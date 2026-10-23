# Local verification

Executed by the coding agent on 2026-10-23 in a Linux amd64 environment using Go 1.27.3, with go.mod declaring Go 1.26.0. Go was absent initially; a toolchain was downloaded from go.dev to `/tmp/reconcilelab-tools/go`. Use `export PATH=/tmp/reconcilelab-tools/go/bin:$PATH` in this session or install Go 1.26+ normally for continued development.

## Commands executed

- `go test ./...`: passed 11 top-level tests, 15 test results including the four example subtests, zero failures.
- `go test -race ./...`: passed across all tested packages.
- `go vet ./...`: passed.
- `go build ./...` and `go build -o bin/reconcilelab ./cmd/reconcilelab`: passed.
- `python3 scripts/cross.py --archive`: compiled Linux/macOS/Windows for amd64 and arm64; generated archives and `dist/SHA256SUMS`. Cross-compilation does not establish native behavior on those platforms.
- `python3 scripts/record_demo.py`: captured four real CLI commands with asserted exit statuses, generated approximately 260 KiB GIF, SVG fallback and raw transcript. A rendered GIF frame was visually inspected.
- `python3 scripts/check_docs.py`: checked local Markdown links and required media assets.
- `git diff --no-index --check /dev/null <file>` across source/documentation files: no whitespace errors, including untracked files.
- `sha256sum -c SHA256SUMS` in `dist`: all six binaries and six archives verified.

## Behavioral evidence

The seven-action basic-controller scenario fails ResourceOwnershipPreserved with the buggy adapter. Chunk reduction produces two actions: external ownership handoff and one reconcile. A fresh replay reproduces the same named failure. The fixed adapter passes both the original and minimized scenario. All four example directories have automated buggy/fixed regression checks.

Tests exercise CRUD and errors, stale cache visibility, stale-update resourceVersion conflicts, fault consumption, copied reads, schema serialization, invalid configuration, cancellation between actions, deterministic replay, deterministic permutation sampling, minimization, built-in invariant predicates, unsupported-controller API errors, CLI reports/exit codes and initialization without clobbering existing configuration.

The actual CLI was also run for init, validate, full stale-cache execution, conflicting-updates with the fixed adapter, report rendering and minimized replay. Buggy runs returned exit 1; fixed runs returned exit 0. Missing files/configuration returned exit 2 in integration tests.

## Unverified boundaries

Docker is installed, but `docker info` failed with permission denied connecting to `/var/run/docker.sock`; image build/run could not be validated. No EnvTest/Kind or live cluster was used. macOS and Windows native test jobs are configured but have not run here. Remote GitHub CI, release upload, human review, human testing and production readiness are unverified. No repository, issues or release were published.

The DevRelay community research connector returned a sign-in requirement; no community-source conclusions were invented. Official controller-runtime fake-client and Kubebuilder EnvTest documentation informed the stated fidelity boundaries.
