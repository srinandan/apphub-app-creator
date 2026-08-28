---
name: unit-testing
description: >-
  Testing conventions for this repo — table-driven Go tests for pkg
  and frontend server, pytest conventions for the orchestrator, fakes for external
  calls, coverage expectations for golden and error paths. Use when
  writing or reviewing any new function, handler, or component with
  logic beyond pure markup.
metadata:
  audience: coding-agent
  version: "0.2.0"
---

# Unit testing

Applies to all code in this repo. Every new exported function, HTTP
handler, or Vue component with logic beyond pure markup needs a test
before the PR is considered done — not added later.

Orchestrator is Python, packages and frontend server are Go — see
[ADR-0008](../../../docs/adr/0008-python-for-orchestrator.md). Don't assume
one language's conventions apply to both.

## Python (`orchestrator/`)

- pytest, parametrized tests (`@pytest.mark.parametrize`) as the Python
  equivalent of table-driven cases — co-located as `test_*.py` or under
  `tests/` mirroring the source layout.
- Test behavior, not implementation — assert on outputs/side effects, not
  on internal call sequences, unless the sequence itself is the contract
  (e.g. "Reviewer must run before execution").
- External calls (Alpaca, BigQuery, Firestore, Agent Registry, Interactions
  API) go behind a small interface/protocol at the point of use, with
  fakes for tests — never hit real external services in unit tests.
- Pydantic models: test validation failures explicitly (e.g. an
  out-of-enum `risk_tolerance`), not just the happy path — this is the
  layer that's supposed to catch schema-invalid data before it reaches
  Firestore.
- Minimum per function: one golden-path case, one error-path case. Add
  edge cases (empty input, boundary values) where the domain has them —
  e.g. zero-quantity trades, expired skill TTL.
- For the Reviewer/Critic and HITL gate specifically: test the *rejection*
  paths as thoroughly as the approval path. This is the governance
  surface — undertested rejection logic defeats the point of the demo.

## Go (`pkg/` and `frontend/server/`)

- Table-driven tests, co-located as `*_test.go` next to the code under
  test.
- Test behavior, not implementation — assert on outputs/side effects, not
  on internal call sequences.
- External calls go behind small interfaces at the point of use, with
  fakes for tests — never hit real external services in unit tests.
- Minimum per function: one golden-path case, one error-path case.

## TypeScript / Vue (`frontend/`)

- Vitest + Testing Library. Test user-visible behavior (what renders,
  what a click does), not component internals or implementation details.
- Every component that renders conditionally (e.g. the approval card's
  approve/edit/reject states) needs a test per state.

## What NOT to do

- Don't write a test just to move a coverage number — see
  `code-coverage/SKILL.md`.
- Don't mock what you don't own carelessly — if a fake for an external
  API drifts from its real behavior, the test suite gives false
  confidence. Keep fakes minimal and reviewed when the real API's
  contract changes.
- Don't test framework code (e.g. that a Pydantic field assigns
  correctly, or that a Go struct's field assignment works) — test your
  logic.

