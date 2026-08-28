---
name: concise-code
description: >-
  Engineering conventions for concise, maintainable code — avoiding
  premature abstraction, dead code, and swallowed errors. Use when
  writing or reviewing any Python, Go, TypeScript, or Vue code in this
  repository.
metadata:
  audience: coding-agent
  version: "0.2.0"
---

# Concise code

Applies to all code in this repo (`orchestrator/` — Python, `pkg/` and `frontend/server/` —
Go, `frontend/src/` — TypeScript/Vue). See
[ADR-0008](../../../docs/adr/0008-python-for-orchestrator.md) for why the
languages split this way.

## Rules

1. **Prefer the straightforward implementation.** Don't introduce an
   interface/protocol, abstraction layer, or generic type until there are
   at least two concrete cases that need it. A single implementation
   doesn't need an interface in front of it "for testability" — use a
   real fake/mock at the call site instead if needed.

2. **One function, one job.** If a function needs a comment to explain
   what its middle section does, that section is probably a function.
   Flag (don't silently accept) any function over ~50 lines or file over
   ~400 lines — either is a signal to split, not a hard rule to enforce
   blindly.

3. **No dead code.** No commented-out code, no unused functions "in case
   we need them later," no speculative config options without a caller.
   Delete it; git history is the archive.

4. **Errors are handled, never swallowed.**
   - Go: every error is either returned (wrapped with
     `fmt.Errorf("doing X: %w", err)` for context) or explicitly handled
     — never `_ = err` or a bare ignored return.
   - Python: catch specific exception types, never a bare `except:` or
     `except Exception: pass`. Re-raise with context (`raise
     RuntimeError("doing X") from err`) rather than swallowing.
   - No generic `catch (e) {}` blocks in TypeScript.

5. **Standard library over dependency, dependency over hand-rolled.** In
   that order of preference. Don't add a package for something the
   language's stdlib or a few lines of idiomatic code already does well
   — applies equally to Go's stdlib and Python's.

6. **Naming says what, types say shape.** Avoid Hungarian-style or
   redundant naming (`stringName`, `dataObj`). A name plus its type
   signature/annotation should make a comment unnecessary for
   straightforward code — this includes Python type hints, not just Go
   types and TypeScript.

## Anti-patterns to flag in review

- A new interface/protocol with exactly one implementation and no test
  double using it
- A config struct/class with fields nothing reads yet
- A generic `utils.go` / `utils.py` / `helpers.ts` grab-bag file
- Deep nesting (>3 levels) instead of early returns / guard clauses
