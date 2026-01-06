# Contract v1 freeze checklist

- [ ] All `contract` package tests pass (`go test ./contract` plus compliance guards).
- [ ] The package stays stdlib-only with no external dependencies.
- [ ] Condition and sort JSON shapes remain stable and match the documented expectations.
- [ ] Semantic versioning policy in `STABILITY.md` is still accurate and unchanged.
- [ ] Assumptions about API JSON shapes are verified against the TypeScript client and API docs before making SDK changes.
