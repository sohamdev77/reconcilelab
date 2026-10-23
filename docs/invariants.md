# Invariant assumptions

Default checks run immediately after every reconcile, even if it returned an error. They compare authoritative live state, never cache state. External actions are not checked against these predicates.

- ResourceOwnershipPreserved compares the `owner` label and all ownerReferences for surviving objects. It does not imply Kubernetes RBAC authorization. Use NoDeletion alongside it to forbid removing owned resources.
- NoDeletion requires all objects present before a reconcile to survive it. This intentionally forbids even a legitimate controller deletion unless you replace the check with a policy-aware custom invariant.
- Unchanged compares the supplied snapshot slices after removing only resourceVersion. Use canonical simulator snapshots; order matters for manually constructed slices. It rejects all other persisted changes.
- Idempotent uses the same state comparison with a distinct violation name. Invoke only after declaring convergence, unchanged observed inputs and permitted side effects. It checks net persisted state, not the number of writes; a no-op update is allowed.

Supply a `func(before, after []unstructured.Unstructured) *invariant.Violation` to engine.Run and engine.Minimize. Return nil to pass. Include expected and actual evidence. Callers own policy assumptions and check selection. Bounds alone do not prove convergence or liveness; neither is currently built in.
