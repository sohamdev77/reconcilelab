package scenario_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/scenario"
)

func TestStrictValidation(t *testing.T) {
	for _, text := range []string{"apiVersion: wrong\nadapter: fixed\n", "apiVersion: reconcilelab/v1alpha1\nadapter: fixed\ntypo: true\n", "apiVersion: reconcilelab/v1alpha1\nadapter: fixed\nactions:\n- kind: sleep\n"} {
		p := filepath.Join(t.TempDir(), "bad.yaml")
		if e := os.WriteFile(p, []byte(text), 0600); e != nil {
			t.Fatal(e)
		}
		if _, e := scenario.Load(p); e == nil {
			t.Fatalf("accepted %s", text)
		}
	}
}
