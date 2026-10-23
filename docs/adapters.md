# Adapter contract

The CLI resolves a directory to `scenario.yaml`; it does not load Go source from the directory. Shipped names `buggy` and `fixed` resolve to real sample factories. To test another controller, import the public engine and construct a registry:

```go
registry := adapter.Registry{
    "my-controller": func(c client.Client) reconcile.Reconciler {
        return &MyReconciler{Client: c}
    },
}
report, err := engine.Run(context.Background(), scenarioValue, registry)
```

Imports are `github.com/sohamdev77/reconcilelab/pkg/{adapter,engine}` and controller-runtime `pkg/{client,reconcile}`. The factory must return a fresh, non-nil controller every call; do not retain global mutable state, start background goroutines, use time/randomness, or connect to real clusters. Only supported ConfigMap client calls are simulated. The engine is sequential and clients are not goroutine-safe. A custom Go test can register invariants as extra Run/Minimize arguments; supplying checks replaces defaults. Keep ownership/deletion checks explicitly if desired. Custom CLI embedding is currently internal; use the public Go engine contract.

Example directories contain data fixtures. `pkg/adapter/adapter.go` contains their shared executable controller; `pkg/engine/engine_test.go` proves both implementations against identical scenarios.
