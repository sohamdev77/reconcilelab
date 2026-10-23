# Fidelity and safety limits

This is a sequential, ConfigMap-only simulator, not kube-apiserver or an informer cache. Supported methods: Get/Create/Update/Delete. List, Patch, DeleteAllOf, status and other subresources return explicit errors. Scheme/RESTMapper are minimal and discovery is unsupported. Namespace lookup reports namespaced resources because ConfigMaps are the sole supported kind.

No API admission, schema/defaulting, managedFields, UID allocation, generation changes, finalizer deletion, garbage collection, watch events, workqueue, conflict retries, automatic requeue, real-clock delays or multi-thread synchronization. Create assigns resourceVersion; Update checks it; Delete removes immediately with no preconditions. Metadata is otherwise user data. These choices require EnvTest/Kind comparison before relying on matching real Kubernetes behavior.

The finite scheduler cannot interrupt a stuck reconciler. Context cancellation is checked between actions and passed to clients/controllers. Use process timeouts for uncooperative trusted controllers. No sandboxing is provided. Fresh factories must avoid global mutable state, randomness and time dependencies. Memory use scales with copied objects times action count; resource-size limits are future work.

Permutation exploration can produce invalid prerequisites and returns an execution error in that case. It samples a bounded set, not all schedules. Shrinking preserves the invariant name, can retain a different instance of that predicate failure, and is not globally minimal.

No cluster configuration is loaded. Future real-cluster integrations must be opt-in with explicit authorization, least-privilege RBAC, read-only preference and data redaction. Current reports copy ConfigMap values verbatim; never use private credentials or commit sensitive output.
