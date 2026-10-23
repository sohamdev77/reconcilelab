// Package engine executes bounded sequential schedules and minimizes counterexamples.
package engine

import (
	"context"
	"errors"
	"fmt"
	"math/rand"

	"github.com/sohamdev77/reconcilelab/pkg/adapter"
	"github.com/sohamdev77/reconcilelab/pkg/invariant"
	"github.com/sohamdev77/reconcilelab/pkg/scenario"
	"github.com/sohamdev77/reconcilelab/pkg/simulator"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

const Version = "0.1.0-dev"

type Step struct {
	Operations []simulator.Operation       `json:"operations"`
	Index      int                         `json:"index"`
	Action     scenario.Action             `json:"action"`
	Result     string                      `json:"result"`
	State      []unstructured.Unstructured `json:"state"`
}
type Report struct {
	Version    string                      `json:"version"`
	Scenario   scenario.Scenario           `json:"scenario"`
	Steps      []Step                      `json:"steps"`
	Operations []simulator.Operation       `json:"operations"`
	Violation  *invariant.Violation        `json:"violation,omitempty"`
	Final      []unstructured.Unstructured `json:"final"`
}

func Run(ctx context.Context, s scenario.Scenario, r adapter.Registry, checks ...invariant.Check) (Report, error) {
	out := Report{Version: Version, Scenario: s, Steps: []Step{}}
	if e := s.Validate(); e != nil {
		return out, e
	}
	f, e := r.Lookup(s.Adapter)
	if e != nil {
		return out, e
	}
	c := simulator.New(s.Initial)
	controller := f(c)
	if controller == nil {
		return out, fmt.Errorf("adapter factory returned nil reconciler")
	}
	if len(checks) == 0 {
		checks = []invariant.Check{invariant.OwnershipPreserved, invariant.NoDeletion}
	}
	for i, a := range s.Actions {
		if e = ctx.Err(); e != nil {
			return out, e
		}
		before := c.Snapshot()
		operationStart := len(c.Operations)
		result := "ok"
		switch a.Kind {
		case "freeze":
			c.Frozen = true
		case "sync":
			c.Sync()
		case "inject":
			c.Faults[a.Operation] = append(c.Faults[a.Operation], a.Error)
		case "create":
			e = c.Create(ctx, a.Object.DeepCopy())
			if e != nil {
				return out, fmt.Errorf("action %d: %w", i, e)
			}
		case "mutate":
			o := a.Object.DeepCopy()
			k := client.ObjectKeyFromObject(o).String()
			old := c.Live[k]
			if old == nil {
				return out, fmt.Errorf("mutate missing %s", k)
			}
			o.SetResourceVersion(old.GetResourceVersion())
			e = c.Update(ctx, o)
			if e != nil {
				return out, fmt.Errorf("action %d: %w", i, e)
			}
		case "delete":
			o := &unstructured.Unstructured{}
			o.SetAPIVersion("v1")
			o.SetKind("ConfigMap")
			o.SetName(a.Name)
			o.SetNamespace(a.Namespace)
			if e = c.Delete(ctx, o); e != nil {
				return out, e
			}
		case "reconcile":
			res, err := controller.Reconcile(ctx, reconcile.Request{NamespacedName: client.ObjectKey{Name: a.Name, Namespace: a.Namespace}})
			if err != nil {
				if errors.Is(err, simulator.ErrUnsupported) {
					return out, fmt.Errorf("unsupported adapter API call: %w", err)
				}
				result = err.Error()
			} else if res.RequeueAfter != 0 {
				result = fmt.Sprintf("requeueAfter=%s (not scheduled automatically)", res.RequeueAfter)
			}
		}
		after := c.Snapshot()
		out.Steps = append(out.Steps, Step{Index: i, Action: a, Result: result, State: after, Operations: append([]simulator.Operation{}, c.Operations[operationStart:]...)})
		if a.Kind == "reconcile" {
			for _, check := range checks {
				if v := check(before, after); v != nil {
					out.Violation = v
					break
				}
			}
		}
		if out.Violation != nil {
			break
		}
	}
	out.Final = c.Snapshot()
	out.Operations = c.Operations
	return out, nil
}

// Minimize uses chunk deletion, preserves invariant identity, and replays the result.
func Minimize(ctx context.Context, s scenario.Scenario, r adapter.Registry, checks ...invariant.Check) (Report, error) {
	base, e := Run(ctx, s, r, checks...)
	if e != nil || base.Violation == nil {
		return base, e
	}
	target := base.Violation.Name
	n := 2
	for len(s.Actions) > 0 {
		if err := ctx.Err(); err != nil {
			return base, err
		}
		size := (len(s.Actions) + n - 1) / n
		changed := false
		for start := 0; start < len(s.Actions); start += size {
			end := start + size
			if end > len(s.Actions) {
				end = len(s.Actions)
			}
			candidate := s
			candidate.Actions = append(append([]scenario.Action{}, s.Actions[:start]...), s.Actions[end:]...)
			v, err := Run(ctx, candidate, r, checks...)
			if err == nil && v.Violation != nil && v.Violation.Name == target {
				s = candidate
				changed = true
				n = 2
				break
			}
		}
		if changed {
			continue
		}
		if n >= len(s.Actions) {
			break
		}
		n *= 2
		if n > len(s.Actions) {
			n = len(s.Actions)
		}
	}
	verified, e := Run(ctx, s, r, checks...)
	if e != nil {
		return verified, e
	}
	if verified.Violation == nil || verified.Violation.Name != target {
		return verified, fmt.Errorf("minimized trace failed replay validation")
	}
	return verified, nil
}

// Explore executes a bounded seeded permutation sample. Explicit action ordering is
// unchanged by Run; invalid permutations are returned as execution errors.
func Explore(ctx context.Context, s scenario.Scenario, r adapter.Registry, budget int) ([]Report, error) {
	if budget < 1 || budget > 1000 {
		return nil, fmt.Errorf("budget must be 1..1000")
	}
	rng := rand.New(rand.NewSource(s.Seed))
	out := []Report{}
	for i := 0; i < budget; i++ {
		v := s
		v.Actions = append([]scenario.Action{}, s.Actions...)
		if i > 0 {
			rng.Shuffle(len(v.Actions), func(i, j int) { v.Actions[i], v.Actions[j] = v.Actions[j], v.Actions[i] })
		}
		report, e := Run(ctx, v, r)
		if e != nil {
			return out, e
		}
		out = append(out, report)
		if report.Violation != nil {
			break
		}
	}
	return out, nil
}
