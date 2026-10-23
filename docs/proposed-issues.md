# Proposed contribution issues

These are proposals, not published GitHub issues.

## 1. ConfigMap name validation

- Problem: Reject invalid DNS names and namespace syntax.
- Relevant files: `pkg/scenario/scenario.go`.
- Expected behavior: invalid names fail before execution.
- Implementation suggestion: Use Kubernetes validation helpers.
- Acceptance criteria: invalid names fail before execution; preserve existing regression coverage.
- Required tests: table-driven name edge cases.
- Difficulty: easy.
- Labels: `good first issue`, `help wanted`.

## 2. Create resourceVersion preconditions

- Problem: Reject create requests with an existing resourceVersion.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: nonempty create RV is rejected.
- Implementation suggestion: Compare Kubernetes create behavior in EnvTest.
- Acceptance criteria: nonempty create RV is rejected; preserve existing regression coverage.
- Required tests: create RV regression.
- Difficulty: medium.
- Labels: `bug`, `help wanted`.

## 3. Delete preconditions

- Problem: Model resourceVersion and UID delete preconditions.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: stale precondition cannot delete current state.
- Implementation suggestion: Introduce explicit supported delete options.
- Acceptance criteria: stale precondition cannot delete current state; preserve existing regression coverage.
- Required tests: matching/mismatched RV and UID.
- Difficulty: medium.
- Labels: `feature`, `help wanted`.

## 4. Typed client schema methods

- Problem: Provide accurate scheme and RESTMapper for ConfigMaps.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: supported discovery returns ConfigMap GVK.
- Implementation suggestion: Register core v1 API mappings.
- Acceptance criteria: supported discovery returns ConfigMap GVK; preserve existing regression coverage.
- Required tests: typed and unstructured GVK tests.
- Difficulty: medium.
- Labels: `controller-runtime`, `help wanted`.

## 5. Per-key frozen observations

- Problem: Freeze only selected objects.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: one key stays stale while others refresh.
- Implementation suggestion: Extend versioned action schema.
- Acceptance criteria: one key stays stale while others refresh; preserve existing regression coverage.
- Required tests: two-key cache regression.
- Difficulty: medium.
- Labels: `feature`, `help wanted`.

## 6. EnvTest conformance suite

- Problem: Compare supported CRUD behavior with a real API server.
- Relevant files: `pkg/simulator/client_test.go`.
- Expected behavior: document every intentional divergence.
- Implementation suggestion: Keep integration tests separately gated.
- Acceptance criteria: document every intentional divergence; preserve existing regression coverage.
- Required tests: CRUD and conflict comparison.
- Difficulty: hard.
- Labels: `testing`, `help wanted`.

## 7. Operation-level invariant hooks

- Problem: Detect no-op write storms and transient forbidden operations.
- Relevant files: `pkg/engine/engine.go`.
- Expected behavior: checks see operations for one reconcile only.
- Implementation suggestion: Expose immutable operation slices to predicates.
- Acceptance criteria: checks see operations for one reconcile only; preserve existing regression coverage.
- Required tests: transient delete/create regression.
- Difficulty: hard.
- Labels: `feature`, `help wanted`.

## 8. Convergence helper

- Problem: Make bounded convergence expectations explicit.
- Relevant files: `pkg/invariant/invariant.go`.
- Expected behavior: failure explains exceeded bound.
- Implementation suggestion: Require user supplied stable-state policy.
- Acceptance criteria: failure explains exceeded bound; preserve existing regression coverage.
- Required tests: convergent/nonconvergent fixture.
- Difficulty: hard.
- Labels: `feature`, `help wanted`.

## 9. Causal exploration

- Problem: Avoid permutations with missing prerequisites.
- Relevant files: `pkg/engine/engine.go`.
- Expected behavior: only dependency-valid schedules run.
- Implementation suggestion: Track explicit dependency edges.
- Acceptance criteria: only dependency-valid schedules run; preserve existing regression coverage.
- Required tests: create-before-update coverage.
- Difficulty: hard.
- Labels: `feature`, `help wanted`.

## 10. Exploration coverage report

- Problem: Report candidate counts and invalid schedules.
- Relevant files: `pkg/engine/engine.go`.
- Expected behavior: budget and coverage are visible.
- Implementation suggestion: Add versioned exploration summary.
- Acceptance criteria: budget and coverage are visible; preserve existing regression coverage.
- Required tests: deterministic summary tests.
- Difficulty: medium.
- Labels: `testing`, `help wanted`.

## 11. Shrinker causal fingerprints

- Problem: Preserve a specific bug instance during reduction.
- Relevant files: `pkg/engine/engine.go`.
- Expected behavior: same-name different-object failures distinguished.
- Implementation suggestion: Allow caller-provided violation fingerprint.
- Acceptance criteria: same-name different-object failures distinguished; preserve existing regression coverage.
- Required tests: two-failure fixture.
- Difficulty: hard.
- Labels: `bug`, `help wanted`.

## 12. Shrinking budget

- Problem: Bound the number of reduction evaluations.
- Relevant files: `pkg/engine/engine.go`.
- Expected behavior: returns best verified candidate on exhaustion.
- Implementation suggestion: Add explicit max evaluations option.
- Acceptance criteria: returns best verified candidate on exhaustion; preserve existing regression coverage.
- Required tests: tiny budget tests.
- Difficulty: medium.
- Labels: `performance`, `help wanted`.

## 13. Resource size limits

- Problem: Limit large object and scenario memory use.
- Relevant files: `pkg/scenario/scenario.go`.
- Expected behavior: oversized input fails clearly.
- Implementation suggestion: Configurable byte ceilings.
- Acceptance criteria: oversized input fails clearly; preserve existing regression coverage.
- Required tests: boundary-size fixtures.
- Difficulty: medium.
- Labels: `feature`, `help wanted`.

## 14. Golden trace fixtures

- Problem: Track intentional trace format changes.
- Relevant files: `pkg/engine/engine_test.go`.
- Expected behavior: review sees trace changes.
- Implementation suggestion: Commit canonical synthetic JSON report.
- Acceptance criteria: review sees trace changes; preserve existing regression coverage.
- Required tests: golden replay comparison.
- Difficulty: easy.
- Labels: `testing`, `help wanted`.

## 15. Invariant policy tutorial

- Problem: Teach ownership and permitted deletion policy.
- Relevant files: `docs/invariants.md`.
- Expected behavior: tutorial example runs as Go test.
- Implementation suggestion: Add compiled custom check example.
- Acceptance criteria: tutorial example runs as Go test; preserve existing regression coverage.
- Required tests: documentation example test.
- Difficulty: easy.
- Labels: `documentation`, `help wanted`.

## 16. Generic object keying

- Problem: Support additional kinds without identity collisions.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: same name across kinds remains independent.
- Implementation suggestion: Include GVK and namespace/name in key.
- Acceptance criteria: same name across kinds remains independent; preserve existing regression coverage.
- Required tests: ConfigMap and custom kind tests.
- Difficulty: hard.
- Labels: `feature`, `help wanted`.

## 17. CLI signal cancellation

- Problem: Pass interrupt cancellation to scenario runs.
- Relevant files: `internal/cli/cli.go`.
- Expected behavior: cooperative controllers stop on interrupt.
- Implementation suggestion: Use signal.NotifyContext in main.
- Acceptance criteria: cooperative controllers stop on interrupt; preserve existing regression coverage.
- Required tests: subprocess signal test on supported OS.
- Difficulty: medium.
- Labels: `feature`, `help wanted`.

## 18. Release provenance

- Problem: Provide signed build provenance.
- Relevant files: `.github/workflows/release.yml`.
- Expected behavior: attestations accompany release archives.
- Implementation suggestion: Use GitHub artifact attestation with least privilege.
- Acceptance criteria: attestations accompany release archives; preserve existing regression coverage.
- Required tests: workflow validation and verification docs.
- Difficulty: medium.
- Labels: `help wanted`, `help wanted`.

## 19. Demo accessibility

- Problem: Add terminal text timing transcript.
- Relevant files: `scripts/record_demo.py`.
- Expected behavior: static fallback conveys all results.
- Implementation suggestion: Emit WebVTT alongside captured output.
- Acceptance criteria: static fallback conveys all results; preserve existing regression coverage.
- Required tests: transcript consistency test.
- Difficulty: easy.
- Labels: `documentation`, `help wanted`.

## 20. Snapshot profiling

- Problem: Measure allocation cost before optimization.
- Relevant files: `pkg/simulator/client.go`.
- Expected behavior: benchmark is reproducible and documented.
- Implementation suggestion: Add deterministic benchmarks without claimed numbers.
- Acceptance criteria: benchmark is reproducible and documented; preserve existing regression coverage.
- Required tests: benchmark plus alias isolation tests.
- Difficulty: medium.
- Labels: `performance`, `help wanted`.
