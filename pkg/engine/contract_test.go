package engine_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/adapter"
	"github.com/sohamdev77/reconcilelab/pkg/engine"
	"github.com/sohamdev77/reconcilelab/pkg/simulator"
	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func TestUnsupportedControllerAPIFailsExecution(t *testing.T) {
	s := fixture(t)
	s.Adapter = "list-controller"
	s.Actions = s.Actions[:1]
	registry := adapter.Registry{"list-controller": func(c client.Client) reconcile.Reconciler {
		return reconcile.Func(func(ctx context.Context, _ reconcile.Request) (reconcile.Result, error) {
			return reconcile.Result{}, c.List(ctx, &corev1.ConfigMapList{})
		})
	}}
	_, err := engine.Run(context.Background(), s, registry)
	if !errors.Is(err, simulator.ErrUnsupported) {
		t.Fatalf("unsupported API must not produce a passing run: %v", err)
	}
}
