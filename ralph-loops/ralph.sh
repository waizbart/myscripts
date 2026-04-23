#!/bin/bash
# Ralph Wiggum - Long-running AI agent loop
# Usage: ./ralph.sh [--tool amp|claude|codex] [max_iterations]

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PRD_FILE="$SCRIPT_DIR/prd.json"
PROGRESS_FILE="$SCRIPT_DIR/progress.txt"
ARCHIVE_DIR="$SCRIPT_DIR/archive"
LAST_BRANCH_FILE="$SCRIPT_DIR/.last-branch"
PROMPT_FILE="${RALPH_PROMPT_FILE:-$SCRIPT_DIR/CLAUDE.md}"
RETRY_SLEEP_SECONDS="${RALPH_RETRY_SLEEP_SECONDS:-300}"
RETRY_MAX_SLEEP_SECONDS="${RALPH_RETRY_MAX_SLEEP_SECONDS:-1800}"
RETRY_BACKOFF_MODE="${RALPH_RETRY_BACKOFF_MODE:-fixed}"

detect_workspace_dir() {
  if [[ -n "${RALPH_WORKSPACE_DIR:-}" ]]; then
    printf '%s\n' "$RALPH_WORKSPACE_DIR"
    return
  fi

  if git -C "$PWD" rev-parse --show-toplevel >/dev/null 2>&1; then
    git -C "$PWD" rev-parse --show-toplevel
    return
  fi

  local git_matches=()
  local git_path

  while IFS= read -r git_path; do
    git_matches+=("$git_path")
  done < <(find "$PWD" -mindepth 2 -maxdepth 4 \( -type d -name .git -o -type f -name .git \) -print 2>/dev/null)

  if [[ ${#git_matches[@]} -eq 1 ]]; then
    dirname "${git_matches[0]}"
    return
  fi

  printf '%s\n' "$PWD"
}

get_tool_bin() {
  case "$TOOL" in
    amp)
      printf 'amp\n'
      ;;
    claude)
      printf 'claude\n'
      ;;
    codex)
      printf 'codex\n'
      ;;
  esac
}

run_selected_tool() {
  case "$TOOL" in
    amp)
      (
        cd "$WORKSPACE_DIR"
        amp --dangerously-allow-all < "$PROMPT_FILE"
      )
      ;;
    claude)
      (
        cd "$WORKSPACE_DIR"
        claude --dangerously-skip-permissions --print < "$PROMPT_FILE"
      )
      ;;
    codex)
      (
        cd "$WORKSPACE_DIR"
        codex exec --dangerously-bypass-approvals-and-sandbox -C "$WORKSPACE_DIR" < "$PROMPT_FILE"
      )
      ;;
  esac
}

is_retryable_limit_error() {
  local output="$1"

  grep -Eiq \
    'usage limit|rate limit|too many requests|429|quota|credit balance|try again later' \
    <<< "$output"
}

# Parse arguments
TOOL="amp"  # Default to amp for backwards compatibility
MAX_ITERATIONS=10

while [[ $# -gt 0 ]]; do
  case $1 in
    --tool)
      TOOL="$2"
      shift 2
      ;;
    --tool=*)
      TOOL="${1#*=}"
      shift
      ;;
    *)
      # Assume it's max_iterations if it's a number
      if [[ "$1" =~ ^[0-9]+$ ]]; then
        MAX_ITERATIONS="$1"
      fi
      shift
      ;;
  esac
done

WORKSPACE_DIR="$(detect_workspace_dir)"

# Validate tool choice
if [[ "$TOOL" != "amp" && "$TOOL" != "claude" && "$TOOL" != "codex" ]]; then
  echo "Error: Invalid tool '$TOOL'. Must be 'amp', 'claude', or 'codex'."
  exit 1
fi

if [[ ! -d "$WORKSPACE_DIR" ]]; then
  echo "Error: Workspace directory '$WORKSPACE_DIR' does not exist."
  exit 1
fi

if [[ ! -f "$PROMPT_FILE" ]]; then
  echo "Error: Prompt file '$PROMPT_FILE' not found."
  exit 1
fi

TOOL_BIN="$(get_tool_bin)"

if ! command -v "$TOOL_BIN" >/dev/null 2>&1; then
  echo "Error: Required CLI '$TOOL_BIN' is not installed or not on PATH."
  exit 1
fi

if [[ "$RETRY_BACKOFF_MODE" != "fixed" && "$RETRY_BACKOFF_MODE" != "exponential" ]]; then
  echo "Error: Invalid RALPH_RETRY_BACKOFF_MODE '$RETRY_BACKOFF_MODE'. Must be 'fixed' or 'exponential'."
  exit 1
fi

# Archive previous run if branch changed
if [ -f "$PRD_FILE" ] && [ -f "$LAST_BRANCH_FILE" ]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  LAST_BRANCH=$(cat "$LAST_BRANCH_FILE" 2>/dev/null || echo "")
  
  if [ -n "$CURRENT_BRANCH" ] && [ -n "$LAST_BRANCH" ] && [ "$CURRENT_BRANCH" != "$LAST_BRANCH" ]; then
    # Archive the previous run
    DATE=$(date +%Y-%m-%d)
    # Strip "ralph/" prefix from branch name for folder
    FOLDER_NAME=$(echo "$LAST_BRANCH" | sed 's|^ralph/||')
    ARCHIVE_FOLDER="$ARCHIVE_DIR/$DATE-$FOLDER_NAME"
    
    echo "Archiving previous run: $LAST_BRANCH"
    mkdir -p "$ARCHIVE_FOLDER"
    [ -f "$PRD_FILE" ] && cp "$PRD_FILE" "$ARCHIVE_FOLDER/"
    [ -f "$PROGRESS_FILE" ] && cp "$PROGRESS_FILE" "$ARCHIVE_FOLDER/"
    echo "   Archived to: $ARCHIVE_FOLDER"
    
    # Reset progress file for new run
    echo "# Ralph Progress Log" > "$PROGRESS_FILE"
    echo "Started: $(date)" >> "$PROGRESS_FILE"
    echo "---" >> "$PROGRESS_FILE"
  fi
fi

# Track current branch
if [ -f "$PRD_FILE" ]; then
  CURRENT_BRANCH=$(jq -r '.branchName // empty' "$PRD_FILE" 2>/dev/null || echo "")
  if [ -n "$CURRENT_BRANCH" ]; then
    echo "$CURRENT_BRANCH" > "$LAST_BRANCH_FILE"
  fi
fi

# Initialize progress file if it doesn't exist
if [ ! -f "$PROGRESS_FILE" ]; then
  echo "# Ralph Progress Log" > "$PROGRESS_FILE"
  echo "Started: $(date)" >> "$PROGRESS_FILE"
  echo "---" >> "$PROGRESS_FILE"
fi

echo "Starting Ralph - Tool: $TOOL - Max iterations: $MAX_ITERATIONS - Workspace: $WORKSPACE_DIR"

for i in $(seq 1 $MAX_ITERATIONS); do
  echo ""
  echo "==============================================================="
  echo "  Ralph Iteration $i of $MAX_ITERATIONS ($TOOL)"
  echo "==============================================================="

  CURRENT_RETRY_SLEEP="$RETRY_SLEEP_SECONDS"

  while true; do
    set +e
    OUTPUT=$(run_selected_tool 2>&1 | tee /dev/stderr)
    STATUS=$?
    set -e

    if [[ $STATUS -eq 0 ]]; then
      break
    fi

    if [[ "$TOOL" != "amp" ]] && is_retryable_limit_error "$OUTPUT"; then
      echo ""
      echo "$TOOL hit a usage or quota limit. Retrying iteration $i after ${CURRENT_RETRY_SLEEP}s..."
      sleep "$CURRENT_RETRY_SLEEP"

      if [[ "$RETRY_BACKOFF_MODE" == "exponential" && "$CURRENT_RETRY_SLEEP" -lt "$RETRY_MAX_SLEEP_SECONDS" ]]; then
        CURRENT_RETRY_SLEEP=$((CURRENT_RETRY_SLEEP * 2))
        if [[ "$CURRENT_RETRY_SLEEP" -gt "$RETRY_MAX_SLEEP_SECONDS" ]]; then
          CURRENT_RETRY_SLEEP="$RETRY_MAX_SLEEP_SECONDS"
        fi
      fi

      continue
    fi

    echo ""
    echo "Ralph stopped because $TOOL exited with status $STATUS on iteration $i."
    exit "$STATUS"
  done
  
  # Check for completion signal
  if echo "$OUTPUT" | grep -q "<promise>COMPLETE</promise>"; then
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
