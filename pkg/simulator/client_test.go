package simulator_test

import (
	"context"
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/simulator"
	corev1 "k8s.io/api/core/v1"
	errors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestAPI(t *testing.T) {
	ctx := context.Background()
	c := simulator.New(nil)
	o := &corev1.ConfigMap{}
	o.Name = "x"
	o.Namespace = "default"
	if e := c.Create(ctx, o); e != nil {
		t.Fatal(e)
	}
	if e := c.Create(ctx, o); !errors.IsAlreadyExists(e) {
		t.Fatal(e)
	}
	stale := o.DeepCopy()
	c.Frozen = true
	o.Data = map[string]string{"v": "new"}
	if e := c.Update(ctx, o); e != nil {
		t.Fatal(e)
	}
	if e := c.Update(ctx, stale); !errors.IsConflict(e) {
		t.Fatal("missing RV conflict", e)
	}
	read := &corev1.ConfigMap{}
	if e := c.Get(ctx, client.ObjectKeyFromObject(o), read); e != nil || read.Data["v"] != "" {
		t.Fatal("stale cache not preserved", e)
	}
	read.Data = map[string]string{"escaped": "copy"}
	if e := c.Get(ctx, client.ObjectKeyFromObject(o), read); e != nil || read.Data["escaped"] != "" {
		t.Fatal("Get retained old fields or leaked a copy", e)
	}
	c.Sync()
	if e := c.Get(ctx, client.ObjectKeyFromObject(o), read); e != nil || read.Data["v"] != "new" {
		t.Fatal(e)
	}
	c.Faults["update"] = []string{"conflict", "unavailable"}
	if e := c.Update(ctx, o); !errors.IsConflict(e) {
		t.Fatal(e)
	}
	if e := c.Update(ctx, o); !errors.IsServiceUnavailable(e) {
		t.Fatal(e)
	}
	if e := c.Update(ctx, o); e != nil {
		t.Fatal(e)
	}
	if e := c.Delete(ctx, o); e != nil {
		t.Fatal(e)
	}
	if e := c.Get(ctx, client.ObjectKeyFromObject(o), read); !errors.IsNotFound(e) {
		t.Fatal(e)
	}
	if e := c.Delete(ctx, o); !errors.IsNotFound(e) {
		t.Fatal(e)
	}
	if len(c.Operations) < 10 {
		t.Fatal("missing operation trace")
	}
	if e := c.List(ctx, &corev1.ConfigMapList{}); e == nil {
		t.Fatal("List silently accepted")
	}
	u := &unstructured.Unstructured{}
	u.SetAPIVersion("v1")
	u.SetKind("Secret")
	if e := c.Create(ctx, u); e == nil {
		t.Fatal("unsupported resource accepted")
	}
}
