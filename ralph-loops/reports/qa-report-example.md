# QA Report
Date: YYYY-MM-DD
Branch: feature/your-branch
PRD: prds/prd-your-feature.json

## Summary
- Total tested: 5
- Passed: 4
- Failed: 1
- New bugs found: 1

## Results

### S1 — Add avatar_url column to users table
**Status:** PASSED
**What was tested:** Migration ran cleanly, column exists in DB, `tsc` passes.
**Notes:** —

---

### S2 — Expose avatar_url in GET /users/:id response
**Status:** PASSED
**What was tested:** `GET /users/1` returns `avatar_url: null` for users without avatar. Returns URL string for users with avatar set.
**Notes:** —

---

### S3 — Accept avatar_url in PATCH /users/:id
**Status:** PASSED
**What was tested:** Valid URL accepted and persisted. Invalid string (non-URL) rejected with 400. Null accepted.
**Notes:** —

---

### S4 — Display avatar in profile page UI
**Status:** FAILED
**What was tested:** Navigated to `/profile` as authenticated user with avatar set.
**Observed behavior:** Avatar image renders but falls back to initials even when `avatar_url` is present.
**Expected behavior:** Should render `<img src={avatar_url}>` when avatar_url is not null.
**Impact:** Medium — cosmetic, no data loss.
**Bug ID:** BUG-1

---

## Bugs Found

| ID | Story | Description | Severity |
|----|-------|-------------|----------|
| BUG-1 | S4 | Avatar image not rendered when avatar_url is present — falls back to initials incorrectly | Medium |
