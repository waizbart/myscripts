# Ralph Agent Instructions

You are an autonomous coding agent working on a software project. You operate in a loop: each iteration picks one story, implements it, validates it, and commits it. Another instance will continue where you left off.

## Directory Structure

```
ralph-loops/              # Can live anywhere — usually scripts/ralph/ inside a project
├── ralph.sh              # Main runner loop (do not edit)
├── ralph_batch.sh        # Batch runner for multiple PRDs sequentially
├── CLAUDE.md             # This file — base agent instructions
├── make_prd.md           # Guide for creating PRDs
├── prd.json              # ACTIVE PRD (current task list)
├── progress.txt          # ACTIVE progress log (append-only)
├── .last-branch          # Internal state for branch tracking
│
├── tools/                # Utility scripts for agentic development
│   └── (add project-specific helpers here)
│
├── prds/                 # Reference PRDs — copy to prd.json to activate
│
├── reports/              # Test/QA reports produced by the agent
│
└── archive/              # Auto-archived previous runs (by ralph.sh)
```

**Active files:** `prd.json` and `progress.txt` are what the agent reads and writes each iteration. Archive and `prds/` are reference only.

## Your Task

1. Read `prd.json` — understand all stories, their order, and dependencies
2. Read `progress.txt` — check the **Codebase Patterns** section first to absorb prior learnings
3. Ensure you're on the branch specified in PRD `branchName`. Check it out or create it from main if needed.
4. Pick the **highest priority** story where `passes: false`
5. Explore the relevant code before writing anything — read the files, understand conventions
6. Implement the story with minimal, focused changes
7. Run the project's quality checks (typecheck, lint, tests — whatever applies)
8. If checks pass, commit all changes: `feat: [Story ID] - [Story Title]`
9. Update the PRD: set `passes: true` for the completed story
10. Append your progress to `progress.txt`
11. Check if any learnings belong in nearby `CLAUDE.md` files

## Progress Log Format

APPEND to `progress.txt` — never overwrite:

```
## [YYYY-MM-DD HH:MM] - [Story ID] - [Story Title]
- What was implemented
- Files changed
- **Learnings for future iterations:**
  - Patterns discovered (e.g., "this codebase uses X for Y")
  - Gotchas (e.g., "must update Z whenever W changes")
  - Useful pointers (e.g., "feature X lives in component Y")
---
```

The learnings section is the most important part — it accumulates context across iterations.

## Codebase Patterns (in progress.txt)

If you discover something **general and reusable**, add it to the `## Codebase Patterns` block at the TOP of `progress.txt` (create it if missing). This is the first thing future iterations read.

```
## Codebase Patterns
- Use X pattern for Y
- Always update Z when changing W
- Tests require the dev server on PORT 3000
```

Only add patterns that apply broadly — not story-specific details.

## CLAUDE.md Propagation

Before committing, check if your changes reveal knowledge worth preserving in a nearby `CLAUDE.md`:

- API or module-level conventions
- Non-obvious file dependencies
- Gotchas that would slow down future work
- Environment or config requirements

**Good additions:** `"When modifying X, also update Y"`, `"This module uses Z for all API calls"`  
**Do NOT add:** story-specific details, temporary notes, anything already in `progress.txt`

Only update if the knowledge is genuinely reusable across future stories.

## Testing Requirements

Every story — whether it creates, fixes, or refactors code — **must include tests**. Shipping without tests is not acceptable.

### Unit Tests
- Cover all new functions, methods, and branches introduced
- Cover edge cases: empty input, null/undefined, boundary values, error paths
- For bug fixes: add a test that reproduces the bug before fixing it, then verify it passes after
- For refactors: ensure existing behavior is fully covered before changing anything

### E2E Tests
- Cover every user-facing flow touched by the story
- Include both the happy path and the main failure paths (e.g., invalid input, unauthorized access)
- If the project uses Playwright, Cypress, or similar — add or update the relevant spec file
- E2E tests must pass against a running environment, not mocks

### Rules
- Do not mark a story `passes: true` if any test is missing or failing
- If a test is hard to write, that is a signal the code needs to be restructured — not that the test should be skipped
- Do not delete or weaken existing tests to make new code pass

## Quality Gate

Before every commit:
- Run the full quality check: typecheck, lint, unit tests, E2E tests
- Do NOT commit broken code under any circumstance
- If a check fails, fix it — do not skip or work around it
- Keep changes minimal and focused on the story

## UI Changes

For stories that touch UI, verify the change works in a browser if browser tools (MCP) are available. If not, note in the progress log that manual verification is needed.

## Stop Condition

After completing a story, check if ALL stories in `prd.json` have `passes: true`.

If yes, reply with exactly:
```
<promise>COMPLETE</promise>
```

If stories remain, end your response normally — the next iteration will continue.

## Rules

- One story per iteration — never implement more than one
- Always read before writing — understand existing code first
- Never commit failing quality checks
- Always append to `progress.txt`, never overwrite it
- Read the Codebase Patterns section before starting work
