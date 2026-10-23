# Scenario format (reconcilelab/v1alpha1)

YAML and JSON share the schema in `pkg/scenario/scenario.go`. Unknown struct fields and duplicate YAML keys are rejected by strict decoding. Objects retain Kubernetes extension fields, but only v1 ConfigMaps are accepted. Names are required; initial namespace/name pairs must be unique. At most 10,000 actions are accepted. API validation of DNS names and ConfigMap contents is not modeled.

Required top-level fields: `apiVersion`, `adapter`. `seed` is an int64. `initial` is an array of objects; `actions` is an ordered array.

| kind | Fields | Effect |
|---|---|---|
| reconcile | name, namespace | One direct call |
| create | object | Create live object |
| mutate | object | Replace an existing live object, using its current RV |
| delete | name, namespace | Remove live object immediately |
| freeze | none | Hold cached observations |
| sync | none | Refresh cache and resume write visibility |
| inject | operation, error | Queue one failure for the next matching API call |

`operation`: get/create/update/delete. `error`: conflict/unavailable. Faults apply to scenario mutations as well as controller operations; an external mutation error terminates execution with code 2. A reconcile error is recorded and execution continues. A propagated simulator ErrUnsupported instead terminates execution with code 2, preventing an unsupported API call from being reported as a passing run. Freeze affects all keys, not a per-key informer. Mutate is a complete replacement, not a JSON merge patch.

See `scenarios/example.yaml`. Unknown adapters require registration in a compiled Go harness, never dynamic code evaluation. `--seed` changes exploration order; plain replay always preserves the explicit action sequence.
