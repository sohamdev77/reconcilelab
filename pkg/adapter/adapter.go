// Package adapter defines a factory contract; every run receives fresh controller state.
package adapter

import (
	"context"
	"fmt"

	corev1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

type Factory func(client.Client) reconcile.Reconciler
type Registry map[string]Factory

func Examples() Registry {
	return Registry{"buggy": func(c client.Client) reconcile.Reconciler { return &Sample{Client: c} }, "fixed": func(c client.Client) reconcile.Reconciler { return &Sample{Client: c, Safe: true} }}
}

// Sample models a controller that assumes every requested ConfigMap belongs to it.
type Sample struct {
	Client client.Client
	Safe   bool
}

func (s *Sample) Reconcile(ctx context.Context, r reconcile.Request) (reconcile.Result, error) {
	o := &corev1.ConfigMap{}
	if e := s.Client.Get(ctx, r.NamespacedName, o); e != nil {
		return reconcile.Result{}, client.IgnoreNotFound(e)
	}
	owner := o.Labels["owner"]
	if s.Safe && owner != "reconcilelab" {
		return reconcile.Result{}, nil
	}
	if o.Data["ready"] == "true" && owner == "reconcilelab" {
		return reconcile.Result{}, nil
	}
	if o.Labels == nil {
		o.Labels = map[string]string{}
	}
	o.Labels["owner"] = "reconcilelab"
	if o.Data == nil {
		o.Data = map[string]string{}
	}
	o.Data["ready"] = "true"
	return reconcile.Result{}, s.Client.Update(ctx, o)
}
func (r Registry) Lookup(name string) (Factory, error) {
	f := r[name]
	if f == nil {
		return nil, fmt.Errorf("unknown adapter %q; register a Go factory", name)
	}
	return f, nil
}
