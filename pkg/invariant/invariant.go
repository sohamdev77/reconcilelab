// Package invariant provides state predicates checked after each reconcile.
package invariant

import (
	"encoding/json"
	"reflect"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type Violation struct {
	Name     string      `json:"name"`
	Expected interface{} `json:"expected"`
	Actual   interface{} `json:"actual"`
}
type Check func(before, after []unstructured.Unstructured) *Violation

func index(s []unstructured.Unstructured) map[string]unstructured.Unstructured {
	m := map[string]unstructured.Unstructured{}
	for _, o := range s {
		m[o.GetNamespace()+"/"+o.GetName()] = o
	}
	return m
}
func OwnershipPreserved(b, a []unstructured.Unstructured) *Violation {
	m := index(a)
	for _, o := range b {
		n, ok := m[o.GetNamespace()+"/"+o.GetName()]
		if ok && (!reflect.DeepEqual(o.GetOwnerReferences(), n.GetOwnerReferences()) || o.GetLabels()["owner"] != n.GetLabels()["owner"]) {
			return &Violation{"ResourceOwnershipPreserved", o.Object, n.Object}
		}
	}
	return nil
}

// NoDeletion forbids deletion of any object present immediately before reconcile.
func NoDeletion(b, a []unstructured.Unstructured) *Violation {
	m := index(a)
	for _, o := range b {
		if _, ok := m[o.GetNamespace()+"/"+o.GetName()]; !ok {
			return &Violation{"UnauthorizedDeletion", o.Object, nil}
		}
	}
	return nil
}

// Unchanged ignores resourceVersion only; callers choose which objects are protected.
func Unchanged(b, a []unstructured.Unstructured) *Violation {
	clean := func(s []unstructured.Unstructured) string {
		v := []map[string]interface{}{}
		for _, o := range s {
			p := o.DeepCopy()
			p.SetResourceVersion("")
			v = append(v, p.Object)
		}
		j, _ := json.Marshal(v)
		return string(j)
	}
	if clean(b) != clean(a) {
		return &Violation{"UnexpectedMutation", b, a}
	}
	return nil
}

// Idempotent is a state predicate: call on snapshots surrounding a repeated reconcile
// only after declaring the input and cache unchanged and convergence reached.
func Idempotent(b, a []unstructured.Unstructured) *Violation {
	v := Unchanged(b, a)
	if v != nil {
		v.Name = "Idempotency"
	}
	return v
}
