# Release Hardening Checklist

This file tracks the pre-Wails hardening work for the current Go CLI/core engine.

## Must Fix

- [ ] Return config errors instead of terminating the process from library code.
- [ ] Support custom prompts correctly.
- [ ] Remove config reloading from AI execution paths.
- [ ] Fix output/apply path handling and duplicate-name edge cases.
- [ ] Fix logging directory consistency and revert path assumptions.
- [ ] Replace machine-local tests with deterministic fixture-driven tests.
- [ ] Get `go test ./...` green in a clean environment.

## Should Fix

- [ ] Trim or implement overstated file extraction/document support.
- [ ] Separate core execution from CLI presentation concerns.
- [ ] Normalize naming and remove obviously dead or misleading code.
- [ ] Tighten release/build docs to match real behavior.
