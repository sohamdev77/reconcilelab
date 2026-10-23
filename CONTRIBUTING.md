# Contributing

Read docs/architecture.md, docs/limitations.md and docs/invariants.md first. Use Go 1.26+. Run `go mod download`, `make test race vet build docs` before opening a PR. Add a focused regression with a deterministic fixture and expected invariant for every correctness change. Explain assumptions and API fidelity boundaries. Keep core paths cluster-free.

Public packages are experimental. Discuss format changes before implementation, preserve replay behavior intentionally and record changes in CHANGELOG.md. Do not add dependencies without justification. Never commit credentials, kubeconfig, private object data, binaries or generated results.

The twenty proposals in docs/proposed-issues.md have suggested labels and acceptance tests. No issues were opened automatically. Maintainers can copy approved proposals into GitHub. Contributors may add adapters, invariants, fixtures, documentation and API-conformance tests. Human review is needed before a public release.
