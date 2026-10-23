package cli

import (
	"bytes"
	"path/filepath"
	"testing"

	"github.com/sohamdev77/reconcilelab/pkg/adapter"
)

func TestCommands(t *testing.T) {
	r := adapter.Examples()
	dir := t.TempDir()
	cases := []struct {
		args []string
		code int
	}{{[]string{"version"}, 0}, {[]string{"validate", "../../fixtures/failure.yaml"}, 0}, {[]string{"test", "../../examples/basic-controller", "--output", dir, "--format", "json"}, 1}, {[]string{"replay", filepath.Join(dir, "counterexample.yaml"), "--adapter", "fixed", "--output", t.TempDir()}, 0}, {[]string{"report", filepath.Join(dir, "run.json")}, 0}, {[]string{"test", "missing.yaml"}, 2}, {[]string{"validate"}, 2}, {[]string{"test", "../../fixtures/failure.yaml", "--adapter", "missing"}, 2}}
	for _, tc := range cases {
		var b bytes.Buffer
		e := Execute(tc.args, &b, &b, r)
		if Code(e) != tc.code {
			t.Fatalf("%v: code %d err %v", tc.args, Code(e), e)
		}
	}
}

func TestInitCreatesExecutableScenarioWithoutClobber(t *testing.T) {
	t.Chdir(t.TempDir())
	var out bytes.Buffer
	if err := Execute([]string{"init"}, &out, &out, adapter.Examples()); err != nil {
		t.Fatal(err)
	}
	if err := Execute([]string{"init"}, &out, &out, adapter.Examples()); Code(err) != 2 {
		t.Fatal("init should refuse an existing file")
	}
	if err := Execute([]string{"test", "reconcilelab.yaml", "--output", "buggy"}, &out, &out, adapter.Examples()); Code(err) != 1 {
		t.Fatalf("starter must expose a real failure: %v", err)
	}
	if err := Execute([]string{"replay", "buggy/counterexample.yaml", "--adapter", "fixed", "--output", "fixed"}, &out, &out, adapter.Examples()); err != nil {
		t.Fatal(err)
	}
}
