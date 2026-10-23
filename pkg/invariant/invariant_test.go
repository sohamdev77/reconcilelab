package invariant_test

import (
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/invariant"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestPredicates(t *testing.T) {
	o := unstructured.Unstructured{Object: map[string]interface{}{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]interface{}{"name": "x", "labels": map[string]interface{}{"owner": "a"}}}}
	b := []unstructured.Unstructured{o}
	p := o.DeepCopy()
	p.SetLabels(map[string]string{"owner": "b"})
	a := []unstructured.Unstructured{*p}
	if invariant.OwnershipPreserved(b, a) == nil || invariant.NoDeletion(b, nil) == nil || invariant.Unchanged(b, a) == nil || invariant.Idempotent(b, a) == nil {
		t.Fatal("missed violation")
	}
	p = o.DeepCopy()
	p.SetResourceVersion("100")
	if invariant.Idempotent(b, []unstructured.Unstructured{*p}) != nil {
		t.Fatal("resourceVersion alone changed semantic state")
	}
	if invariant.OwnershipPreserved(b, b) != nil || invariant.NoDeletion(b, b) != nil {
		t.Fatal("false positive")
	}
}
