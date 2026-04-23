#!/bin/bash
# Ralph Wiggum - Long-running AI agent loop
# Usage: ./ralph.sh [--tool amp|claude|codex] [max_iterations]

set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./ralph.sh [--tool amp|claude|codex] [max_iterations]

Environment:
  RALPH_WORKSPACE_DIR           Directory where the agent should work (default: current Git repo root; or a single nested repo if auto-detected)
  RALPH_PROMPT_FILE             Prompt file to feed into the agent (default: ralph-loops/CLAUDE.md)
  RALPH_PROMPT                  Inline prompt content. Overrides RALPH_PROMPT_FILE when set.
  RALPH_RETRY_SLEEP_SECONDS     Wait between retryable limit errors (default: 300)
  RALPH_RETRY_MAX_SLEEP_SECONDS Max wait when exponential backoff is enabled (default: 1800)
  RALPH_RETRY_BACKOFF_MODE      Retry strategy: fixed or exponential (default: fixed)
EOF
}

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRD_FILE="$SCRIPT_DIR/prd.json"
PROGRESS_FILE="$SCRIPT_DIR/progress.txt"
ARCHIVE_DIR="$SCRIPT_DIR/archive"
LAST_BRANCH_FILE="$SCRIPT_DIR/.last-branch"
PROMPT_FILE="${RALPH_PROMPT_FILE:-$SCRIPT_DIR/CLAUDE.md}"
WORKSPACE_DIR=""
RETRY_SLEEP_SECONDS="${RALPH_RETRY_SLEEP_SECONDS:-300}"
RETRY_MAX_SLEEP_SECONDS="${RALPH_RETRY_MAX_SLEEP_SECONDS:-1800}"
RETRY_BACKOFF_MODE="${RALPH_RETRY_BACKOFF_MODE:-fixed}"

TOOL="amp"  # Default to amp for backwards compatibility
MAX_ITERATIONS=10
TOOL_OUTPUT=""
TOOL_EXIT_STATUS=0
PROMPT_SOURCE=""
TEMP_FILES=()

cleanup() {
  if [[ ${#TEMP_FILES[@]} -gt 0 ]]; then
    rm -f "${TEMP_FILES[@]}"
  fi
}

trap cleanup EXIT

while [[ $# -gt 0 ]]; do
  case "$1" in
    --tool)
      if [[ $# -lt 2 ]]; then
        echo "Error: --tool requires a value."
        usage
        exit 1
      fi
      TOOL="$2"
      shift 2
      ;;
    --tool=*)
      TOOL="${1#*=}"
      shift
      ;;
    -h|--help)
      usage
      exit 0
      ;;
    *)
      if [[ "$1" =~ ^[0-9]+$ ]]; then
        MAX_ITERATIONS="$1"
      else
        echo "Error: Unrecognized argument '$1'."
        usage
        exit 1
      fi
      shift
      ;;
  esac
done

validate_tool() {
  case "$TOOL" in
    amp|claude|codex)
      ;;
    *)
      echo "Error: Invalid tool '$TOOL'. Must be 'amp', 'claude', or 'codex'."
      exit 1
      ;;
  esac
}

validate_number() {
  local label="$1"
  local value="$2"

  if ! [[ "$value" =~ ^[0-9]+$ ]] || (( value <= 0 )); then
    echo "Error: $label must be a positive integer. Got '$value'."
    exit 1
  fi
}

validate_config() {
  validate_tool
  validate_number "max_iterations" "$MAX_ITERATIONS"
  validate_number "RALPH_RETRY_SLEEP_SECONDS" "$RETRY_SLEEP_SECONDS"
  validate_number "RALPH_RETRY_MAX_SLEEP_SECONDS" "$RETRY_MAX_SLEEP_SECONDS"

  case "$RETRY_BACKOFF_MODE" in
    fixed|exponential)
      ;;
    *)
      echo "Error: RALPH_RETRY_BACKOFF_MODE must be 'fixed' or 'exponential'."
      exit 1
      ;;
  esac

  if [[ ! -d "$WORKSPACE_DIR" ]]; then
    echo "Error: RALPH_WORKSPACE_DIR '$WORKSPACE_DIR' does not exist."
    exit 1
  fi

  if ! git -C "$WORKSPACE_DIR" rev-parse --show-toplevel >/dev/null 2>&1; then
    echo "Error: Workspace '$WORKSPACE_DIR' is not inside a Git repository."
    echo "Set RALPH_WORKSPACE_DIR to your project repository root before running Ralph."
    exit 1
  fi
}

resolve_workspace_dir() {
  local requested_dir
  local git_root
  local repo_candidates=()
  local candidate

  requested_dir="${RALPH_WORKSPACE_DIR:-$PWD}"

  if [[ ! -d "$requested_dir" ]]; then
    WORKSPACE_DIR="$requested_dir"
    return
  fi

  git_root="$(git -C "$requested_dir" rev-parse --show-toplevel 2>/dev/null || true)"
  if [[ -n "$git_root" ]]; then
    WORKSPACE_DIR="$git_root"
    return
  fi

  if [[ -z "${RALPH_WORKSPACE_DIR:-}" ]]; then
    while IFS= read -r candidate; do
      repo_candidates+=("$(dirname "$candidate")")
    done < <(find "$requested_dir" -mindepth 2 -maxdepth 2 -type d -name .git 2>/dev/null | sort)

    if (( ${#repo_candidates[@]} == 1 )); then
      WORKSPACE_DIR="${repo_candidates[0]}"
      echo "Auto-detected workspace Git repository: $WORKSPACE_DIR"
      return
    fi
  fi

  WORKSPACE_DIR="$requested_dir"
}

required_command() {
  case "$TOOL" in
    amp)
      echo "amp"
      ;;
    claude)
      echo "claude"
      ;;
    codex)
      echo "codex"
      ;;
  esac
}

ensure_tool_installed() {
  local command_name
  command_name="$(required_command)"

  if ! command -v "$command_name" >/dev/null 2>&1; then
    echo "Error: Required command '$command_name' was not found in PATH."
    exit 1
  fi
}

prepare_prompt_source() {
  if [[ -n "${RALPH_PROMPT:-}" ]]; then
    PROMPT_SOURCE="$(mktemp)"
    TEMP_FILES+=("$PROMPT_SOURCE")
    printf '%s\n' "$RALPH_PROMPT" > "$PROMPT_SOURCE"
    return
  fi

  if [[ ! -f "$PROMPT_FILE" ]]; then
    echo "Error: Prompt file '$PROMPT_FILE' was not found."
    exit 1
  fi

  PROMPT_SOURCE="$PROMPT_FILE"
}

run_command_with_prompt() {
  local output_file="$1"
  shift
  local status=0

  set +e
  "$@" < "$PROMPT_SOURCE" 2>&1 | tee /dev/stderr > "$output_file"
  status=$?
  set -e

  return "$status"
}

run_tool_once() {
  local output_file="$1"

  case "$TOOL" in
    amp)
      run_command_with_prompt "$output_file" amp --dangerously-allow-all
      ;;
    claude)
      run_command_with_prompt "$output_file" claude --dangerously-skip-permissions --print
      ;;
    codex)
      run_command_with_prompt "$output_file" codex exec --full-auto --skip-git-repo-check -C "$WORKSPACE_DIR"
      ;;
  esac
}

is_retryable_limit_error() {
  local output="$1"
  local retry_pattern='usage limit|rate limit|insufficient_quota|quota exceeded|exceeded your current quota|exceeded your .* limit|too many requests|429[^[:alnum:]]+too many requests|credit balance is too low|retry after [0-9]+|please try again in [0-9]+'

  grep -Eiq "$retry_pattern" <<<"$output"
}

next_retry_sleep() {
  local attempt="$1"
  local sleep_seconds="$RETRY_SLEEP_SECONDS"
  local current_attempt=1

  if [[ "$RETRY_BACKOFF_MODE" == "exponential" ]]; then
    while (( current_attempt < attempt && sleep_seconds < RETRY_MAX_SLEEP_SECONDS )); do
      sleep_seconds=$((sleep_seconds * 2))
      if (( sleep_seconds > RETRY_MAX_SLEEP_SECONDS )); then
        sleep_seconds="$RETRY_MAX_SLEEP_SECONDS"
      fi
      current_attempt=$((current_attempt + 1))
    done
  fi

  if (( sleep_seconds > RETRY_MAX_SLEEP_SECONDS )); then
    sleep_seconds="$RETRY_MAX_SLEEP_SECONDS"
  fi

  echo "$sleep_seconds"
}

run_tool_with_retry() {
  local output_file
  local attempt=1
  local sleep_seconds
  local timestamp

  output_file="$(mktemp)"
  TEMP_FILES+=("$output_file")

  while true; do
    if run_tool_once "$output_file"; then
      TOOL_EXIT_STATUS=0
    else
      TOOL_EXIT_STATUS=$?
    fi

    TOOL_OUTPUT="$(<"$output_file")"

    if is_retryable_limit_error "$TOOL_OUTPUT"; then
      sleep_seconds="$(next_retry_sleep "$attempt")"
      timestamp="$(date '+%Y-%m-%d %H:%M:%S')"
      echo "[$timestamp] $TOOL hit a usage limit. Retry attempt $attempt will wait ${sleep_seconds}s before retrying."
      attempt=$((attempt + 1))
      sleep "$sleep_seconds"
      continue
    fi

    return 0
  done
}

resolve_workspace_dir
validate_config
ensure_tool_installed
prepare_prompt_source

cd "$WORKSPACE_DIR"

# Archive previous run if branch changed
if [[ -f "$PRD_FILE" && -f "$LAST_BRANCH_FILE" ]]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  LAST_BRANCH=$(cat "$LAST_BRANCH_FILE" 2>/dev/null || echo "")

  if [[ -n "$CURRENT_BRANCH" && -n "$LAST_BRANCH" && "$CURRENT_BRANCH" != "$LAST_BRANCH" ]]; then
    DATE=$(date +%Y-%m-%d)
    FOLDER_NAME=$(echo "$LAST_BRANCH" | sed 's|^ralph/||')
    ARCHIVE_FOLDER="$ARCHIVE_DIR/$DATE-$FOLDER_NAME"

    echo "Archiving previous run: $LAST_BRANCH"
    mkdir -p "$ARCHIVE_FOLDER"
    [[ -f "$PRD_FILE" ]] && cp "$PRD_FILE" "$ARCHIVE_FOLDER/"
    [[ -f "$PROGRESS_FILE" ]] && cp "$PROGRESS_FILE" "$ARCHIVE_FOLDER/"
    echo "   Archived to: $ARCHIVE_FOLDER"

    echo "# Ralph Progress Log" > "$PROGRESS_FILE"
    echo "Started: $(date)" >> "$PROGRESS_FILE"
    echo "---" >> "$PROGRESS_FILE"
  fi
fi

# Track current branch
if [[ -f "$PRD_FILE" ]]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  if [[ -n "$CURRENT_BRANCH" ]]; then
    echo "$CURRENT_BRANCH" > "$LAST_BRANCH_FILE"
  fi
fi

# Initialize progress file if it doesn't exist
if [[ ! -f "$PROGRESS_FILE" ]]; then
  echo "# Ralph Progress Log" > "$PROGRESS_FILE"
  echo "Started: $(date)" >> "$PROGRESS_FILE"
  echo "---" >> "$PROGRESS_FILE"
fi

echo "Starting Ralph - Tool: $TOOL - Max iterations: $MAX_ITERATIONS - Workspace: $WORKSPACE_DIR"

for i in $(seq 1 "$MAX_ITERATIONS"); do
  echo ""
  echo "==============================================================="
  echo "  Ralph Iteration $i of $MAX_ITERATIONS ($TOOL)"
  echo "==============================================================="

  run_tool_with_retry

  if (( TOOL_EXIT_STATUS != 0 )); then
    echo ""
    echo "Ralph stopped because '$TOOL' exited with status $TOOL_EXIT_STATUS."
    echo "The failure was not classified as a retryable usage-limit error."
    exit "$TOOL_EXIT_STATUS"
  fi

  if grep -q "<promise>COMPLETE</promise>" <<<"$TOOL_OUTPUT"; then
    echo ""
    echo "Ralph completed all tasks!"
    echo "Completed at iteration $i of $MAX_ITERATIONS"
    exit 0
  fi

  echo "Iteration $i complete. Continuing..."
  sleep 2
done

echo ""
echo "Ralph reached max iterations ($MAX_ITERATIONS) without completing all tasks."
echo "Check $PROGRESS_FILE for status."
exit 1
