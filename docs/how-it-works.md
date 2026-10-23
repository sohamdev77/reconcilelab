# How it works

1. Strictly decode a versioned data scenario and resolve its named factory.
2. Deep copy initial ConfigMaps into live storage, assign deterministic resource versions, synchronize observations.
3. Execute actions in order without sleeps or goroutines.
4. For reconcile actions, snapshot live state, call the controller, record errors/requeue requests and evaluate invariants on resulting live state.
5. Stop at the first violation. Serialize evidence with no wall-clock timestamps.
6. Remove action chunks, rerun from initial state with a fresh factory, keep reductions reproducing the same invariant, and verify the final candidate again.

The cache does not spontaneously catch up. RequeueAfter is recorded but never automatically scheduled; add explicit reconcile actions. Conflicts are observed errors, not automatically retried. The scenario controls every subsequent reconcile.
