// Package scenario defines the versioned, data-only scenario format.
package scenario

import (
	"encoding/json"
	"fmt"
	"os"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"sigs.k8s.io/yaml"
)

type Action struct {
	Kind      string                     `json:"kind"`
	Name      string                     `json:"name,omitempty"`
	Namespace string                     `json:"namespace,omitempty"`
	Object    *unstructured.Unstructured `json:"object,omitempty"`
	Operation string                     `json:"operation,omitempty"`
	Error     string                     `json:"error,omitempty"`
}
type Scenario struct {
	APIVersion string                      `json:"apiVersion"`
	Adapter    string                      `json:"adapter"`
	Seed       int64                       `json:"seed"`
	Initial    []unstructured.Unstructured `json:"initial"`
	Actions    []Action                    `json:"actions"`
}

func (s Scenario) Validate() error {
	if s.APIVersion != "reconcilelab/v1alpha1" {
		return fmt.Errorf("unsupported apiVersion %q", s.APIVersion)
	}
	if s.Adapter == "" {
		return fmt.Errorf("adapter is required")
	}
	if len(s.Actions) > 10000 {
		return fmt.Errorf("action limit is 10000")
	}
	seen := map[string]bool{}
	for _, o := range s.Initial {
		k := o.GetNamespace() + "/" + o.GetName()
		if o.GetName() == "" || o.GetKind() != "ConfigMap" || o.GetAPIVersion() != "v1" || seen[k] {
			return fmt.Errorf("initial objects must be unique v1 ConfigMaps with names")
		}
		seen[k] = true
	}
	for i, a := range s.Actions {
		switch a.Kind {
		case "reconcile", "delete":
			if a.Name == "" {
				return fmt.Errorf("action %d requires name", i)
			}
		case "create", "mutate":
			if a.Object == nil || a.Object.GetName() == "" || a.Object.GetKind() != "ConfigMap" || a.Object.GetAPIVersion() != "v1" {
				return fmt.Errorf("action %d requires a v1 ConfigMap", i)
			}
		case "freeze", "sync":
		case "inject":
			if a.Operation != "get" && a.Operation != "create" && a.Operation != "update" && a.Operation != "delete" {
				return fmt.Errorf("invalid injection operation")
			}
			if a.Error != "conflict" && a.Error != "unavailable" {
				return fmt.Errorf("invalid injected error")
			}
		default:
			return fmt.Errorf("unknown action %d: %s", i, a.Kind)
		}
	}
	return nil
}
func Load(path string) (Scenario, error) {
	var s Scenario
	b, e := os.ReadFile(path)
	if e != nil {
		return s, e
	}
	e = yaml.UnmarshalStrict(b, &s)
	if e == nil {
		e = s.Validate()
	}
	return s, e
}
func Save(path string, s Scenario) error {
	b, e := json.Marshal(s)
	if e != nil {
		return e
	}
	b, e = yaml.JSONToYAML(b)
	if e != nil {
		return e
	}
	return os.WriteFile(path, b, 0644)
}
