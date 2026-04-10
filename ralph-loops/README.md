# Ralph Loops

Autonomous AI agent loop for software development. Ralph runs in iterations — each one picks a single user story from a PRD, implements it, tests it, commits it, and signals the next iteration to continue.

Works with [Claude Code](https://claude.ai/code) and [Amp](https://ampcode.com).

## How It Works

```
ralph.sh runs in a loop
  └─ each iteration:
       1. reads prd.json       → picks the next story
       2. reads progress.txt   → absorbs prior learnings
       3. implements the story
       4. runs quality checks + tests (unit + E2E)
       5. commits
       6. updates prd.json     → marks story as passes: true
       7. appends to progress.txt
       └─ if all stories done → emits <promise>COMPLETE</promise>
```

## Directory Structure

```
ralph-loops/
├── ralph.sh              # Main runner loop
├── ralph_batch.sh        # Batch runner for multiple PRDs sequentially
├── CLAUDE.md             # Agent instructions (generic, drop-in for any project)
├── make_prd.md           # Prompt to help generate prd.json files
├── prd.json              # ACTIVE PRD — edit this to change what Ralph works on
├── progress.txt          # Append-only log of what was done and learned
│
├── tools/                # Utility scripts for agentic development
├── prds/                 # Reference PRDs — copy one to prd.json to activate
├── reports/              # Test/QA reports produced by the agent
└── archive/              # Auto-archived runs when branch changes
```

## Quick Start

**1. Create a PRD**

Use `make_prd.md` as a prompt to Claude to generate a `prd.json`, or write one manually:

```json
{
  "branchName": "feature/my-feature",
  "userStories": [
    {
      "id": "1",
      "title": "Add X to Y",
      "description": "What to do, where to do it, and how to verify it's correct. Verify with: npm test.",
      "passes": false
    }
  ]
}
```

Save it as `prd.json` in this directory.

**2. Run Ralph**

```bash
# Using Claude Code (default: 10 iterations)
./ralph.sh --tool claude

# Using Amp
./ralph.sh --tool amp

# Custom iteration limit
./ralph.sh --tool claude 25
```

**3. Run multiple PRDs in sequence**

Edit `ralph_batch.sh` to list your PRDs, then:

```bash
./ralph_batch.sh
```

## PRD Format

```json
{
  "branchName": "feature/name-in-kebab-case",
  "userStories": [
    {
      "id": "1",
      "title": "Short, specific title",
      "description": "What to do, where (file/folder), and how to verify it's correct.",
      "passes": false
    }
  ]
}
```

**Good story descriptions:**
- "Add `expires_at` column to `sessions` table via migration. Verify with: `npx tsc --noEmit`."
- "Create `getUserById` resolver in `src/resolvers/user.ts` following existing resolver patterns. Verify with: `npx tsc --noEmit` and `npm test`."

**Bad story descriptions:**
- "Implement authentication"
- "Refactor the user module"

Keep each story atomic — one responsibility, one commit.

## Agent Instructions (`CLAUDE.md`)

`CLAUDE.md` is the system prompt passed to the agent each iteration. It is generic and works for any project. It instructs the agent to:

- Read before writing — understand existing conventions first
- Write unit and E2E tests for everything created, fixed, or refactored
- Never commit failing quality checks
- Propagate learnings to `progress.txt` and nearby `CLAUDE.md` files
- Work on exactly one story per iteration

To add project-specific context (stack details, env setup, QA instructions), either append to `CLAUDE.md` or pass a composed prompt via the `RALPH_PROMPT` env var.

## Progress & Learnings

`progress.txt` accumulates context across all iterations. The agent reads it at the start of every run. Structure:

```
## Codebase Patterns          ← read first, updated as patterns are discovered
- Use X for Y
- Always update Z when changing W

## 2026-04-10 14:30 - Story 1 - Title
- What was implemented
- Files changed
- Learnings for future iterations
---
```

## State & Archiving

When `ralph.sh` detects a branch change between runs (via `.last-branch`), it automatically archives the previous `prd.json` and `progress.txt` into `archive/YYYY-MM-DD-branch-name/` and resets `progress.txt` for the new run.

## Requirements

- `jq` — for parsing `prd.json`
- `claude` CLI — for `--tool claude` ([install](https://claude.ai/code))
- `amp` CLI — for `--tool amp` ([install](https://ampcode.com))
