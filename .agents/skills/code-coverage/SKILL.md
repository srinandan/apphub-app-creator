---
name: code-coverage
description: >-
  Coverage targets and exclusions for this repo (Python for the
  orchestrator, Go for pkg and frontend server), and why coverage should be treated
  as a signal rather than a target. Use when assessing whether a change
  has adequate test coverage, or when coverage drops in CI.
metadata:
  audience: coding-agent
  version: "0.2.0"
---

# Code coverage

Coverage is a signal for finding untested code, not a target to write
tests against. Don't add a test whose only purpose is moving the number.

## Targets

- `orchestrator/` (Python): **80% line coverage**, measured with
  `pytest --cov=orchestrator --cov-report=term-missing`
- Go packages (`pkg/`, `frontend/server/`): **80% line coverage**, measured with
  `go test ./... -coverprofile=coverage.out && go tool cover -func=coverage.out`
- `frontend/` (Vue/TS): **60% line coverage**, advisory rather than
  enforced — UI code has a lower ratio of logic to markup, so the number
  is less meaningful here
- The Reviewer/Critic and HITL approval path specifically: treat as
  needing coverage closer to 100%, regardless of the module-wide target —
  this is the governance surface the whole demo depends on being
  trustworthy

## Exclusions

Don't count against the target, and don't chase coverage on:
- Generated code
- `main.go` / `__main__.py` / wiring and dependency-injection setup with
  no branching logic
- Thin wrappers around external SDKs with no logic of their own

## Where enforced

CI gate, once CI config exists (not yet built — see repo status
in the root `README.md`). Until then, run coverage locally before opening
a PR:

```bash
pytest --cov            # orchestrator/
go test ./... -cover    # pkg/ and frontend/server/
npm run test -- --coverage    # frontend/
```

## When coverage drops

Treat a coverage regression as a prompt to ask "what code just got added
without a test," not as a number to patch. If the honest answer is "this
line genuinely doesn't need a test" (e.g. a trivial getter), that's a
valid outcome — say so in the PR rather than padding with a low-value test.
