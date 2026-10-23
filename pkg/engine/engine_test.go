package engine_test

import (
	"context"
	"encoding/json"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/adapter"
	"github.com/sohamdev77/reconcilelab/pkg/engine"
	"github.com/sohamdev77/reconcilelab/pkg/scenario"
)

func fixture(t *testing.T) scenario.Scenario {
	t.Helper()
	s, e := scenario.Load("../../fixtures/failure.yaml")
	if e != nil {
		t.Fatal(e)
	}
	return s
}
func TestRegressionReplayShrink(t *testing.T) {
	ctx := context.Background()
	r := adapter.Examples()
	s := fixture(t)
	first, e := engine.Run(ctx, s, r)
	if e != nil || first.Violation == nil {
		t.Fatalf("missing failure: %v", e)
	}
	again, e := engine.Run(ctx, s, r)
	if e != nil || !reflect.DeepEqual(first, again) {
		t.Fatal("nondeterministic replay", e)
	}
	small, e := engine.Minimize(ctx, s, r)
	if e != nil || small.Violation == nil {
		t.Fatal(e)
	}
	if len(small.Scenario.Actions) != 2 {
		t.Fatalf("expected mutation + reconcile, got %d", len(small.Scenario.Actions))
	}
	p := filepath.Join(t.TempDir(), "case.yaml")
	if e = scenario.Save(p, small.Scenario); e != nil {
		t.Fatal(e)
	}
	loaded, e := scenario.Load(p)
	if e != nil {
		t.Fatal(e)
	}
	replayed, e := engine.Run(ctx, loaded, r)
	if e != nil || !reflect.DeepEqual(small, replayed) {
		t.Fatal("serialized replay differs", e)
	}
	loaded.Adapter = "fixed"
	fixed, e := engine.Run(ctx, loaded, r)
	if e != nil || fixed.Violation != nil {
		t.Fatalf("fixed regressed: %v %+v", e, fixed.Violation)
	}
	s.Adapter = "fixed"
	fixed, e = engine.Run(ctx, s, r)
	if e != nil || fixed.Violation != nil {
		t.Fatal("full fixed scenario failed", e)
	}
	b, _ := json.Marshal(first)
	if !json.Valid(b) {
		t.Fatal("invalid report")
	}
}
func TestScheduleAndExploration(t *testing.T) {
	s := fixture(t)
	s.Adapter = "fixed"
	s.Actions = s.Actions[:1]
	s.Actions = append(s.Actions, s.Actions[0])
	a, e := engine.Explore(context.Background(), s, adapter.Examples(), 10)
	if e != nil || len(a) != 10 {
		t.Fatal(e)
	}
	b, e := engine.Explore(context.Background(), s, adapter.Examples(), 10)
	if e != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("seeded exploration differs")
	}
	if len(a[0].Steps) != 2 {
		t.Fatal("duplicate lost")
	}
	if _, e = engine.Explore(context.Background(), s, adapter.Examples(), 0); e == nil {
		t.Fatal("invalid bound accepted")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, e = engine.Run(ctx, s, adapter.Examples()); e == nil {
		t.Fatal("ignored cancellation")
	}
}

func TestStaleReadConflictAndRecovery(t *testing.T) {
	s := fixture(t)
	// An externally changed object remains invisible while frozen. The unsafe
	// controller attempts a stale write, which must conflict rather than overwrite.
	s.Initial[0].Object["data"] = map[string]interface{}{"ready": "false"}
	s.Actions = s.Actions[1:4]
	report, err := engine.Run(context.Background(), s, adapter.Examples())
	if err != nil || report.Violation != nil {
		t.Fatalf("stale write should conflict: %v %+v", err, report.Violation)
	}
	if len(report.Operations) != 3 || report.Operations[2].Verb != "update" || report.Operations[2].Result == "ok" {
		t.Fatalf("missing stale update conflict: %+v", report.Operations)
	}
	if report.Final[0].GetLabels()["owner"] != "another-controller" {
		t.Fatal("stale update stole ownership")
	}
	s.Actions = append(s.Actions, scenario.Action{Kind: "sync"}, scenario.Action{Kind: "reconcile", Name: "demo", Namespace: "default"})
	report, err = engine.Run(context.Background(), s, adapter.Examples())
	if err != nil || report.Violation == nil {
		t.Fatal("fresh observation should expose unsafe mutation", err)
	}
}

func TestPermutationActuallyChangesRequestOrder(t *testing.T) {
	s := fixture(t)
	s.Adapter = "fixed"
	s.Actions = []scenario.Action{{Kind: "reconcile", Name: "a"}, {Kind: "reconcile", Name: "b"}, {Kind: "reconcile", Name: "c"}}
	reports, err := engine.Explore(context.Background(), s, adapter.Examples(), 12)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for _, r := range reports {
		if !reflect.DeepEqual(r.Scenario.Actions, s.Actions) {
			changed = true
		}
	}
	if !changed {
		t.Fatal("exploration never reordered requests")
	}
}

func TestAllExampleRegressions(t *testing.T) {
	for _, name := range []string{"basic-controller", "duplicate-reconcile", "stale-cache", "conflicting-updates"} {
		t.Run(name, func(t *testing.T) {
			s, err := scenario.Load("../../examples/" + name + "/scenario.yaml")
			if err != nil {
				t.Fatal(err)
			}
			buggy, err := engine.Run(context.Background(), s, adapter.Examples())
			if err != nil || buggy.Violation == nil {
				t.Fatalf("missing bug: %v", err)
			}
			s.Adapter = "fixed"
			fixed, err := engine.Run(context.Background(), s, adapter.Examples())
			if err != nil || fixed.Violation != nil {
				t.Fatalf("fixed regression: %v %+v", err, fixed.Violation)
			}
		})
	}
}
