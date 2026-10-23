package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"

	"github.com/sohamdev77/reconcilelab/pkg/adapter"
	"github.com/sohamdev77/reconcilelab/pkg/engine"
	"github.com/sohamdev77/reconcilelab/pkg/scenario"
	"github.com/spf13/cobra"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// ExitError lets embedders preserve CI exit codes without parsing messages.
type ExitError struct {
	Code int
	Err  error
}

func (e *ExitError) Error() string { return e.Err.Error() }
func Code(e error) int {
	if e == nil {
		return 0
	}
	if v, ok := e.(*ExitError); ok {
		return v.Code
	}
	return 2
}

// Execute supports custom compiled adapter registries. Scenarios never load code.
func Execute(args []string, out, errOut io.Writer, r adapter.Registry) error {
	root := &cobra.Command{Use: "reconcilelab", Short: "Break the Reconcile Loop Before Production Does.", SilenceUsage: true, SilenceErrors: true}
	root.SetArgs(args)
	root.SetOut(out)
	root.SetErr(errOut)
	root.AddCommand(&cobra.Command{Use: "version", Args: cobra.NoArgs, RunE: func(c *cobra.Command, a []string) error { _, e := fmt.Fprintln(out, engine.Version); return e }})
	root.AddCommand(&cobra.Command{Use: "init", Args: cobra.NoArgs, RunE: func(c *cobra.Command, a []string) error {
		path := "reconcilelab.yaml"
		if _, e := os.Stat(path); e == nil {
			return fmt.Errorf("%s already exists", path)
		}
		s := Default()
		if e := scenario.Save(path, s); e != nil {
			return e
		}
		_, e := fmt.Fprintln(out, "Created reconcilelab.yaml (local simulation)")
		return e
	}})
	root.AddCommand(&cobra.Command{Use: "validate SCENARIO", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, a []string) error {
		s, e := scenario.Load(a[0])
		if e != nil {
			return e
		}
		if _, e = r.Lookup(s.Adapter); e != nil {
			return e
		}
		_, e = fmt.Fprintln(out, "Valid scenario:", a[0])
		return e
	}})
	var seed int64
	var format, dir, adapterName string
	var minimize bool
	var budget int
	run := func(c *cobra.Command, a []string) error {
		path := a[0]
		if info, e := os.Stat(path); e == nil && info.IsDir() {
			path = filepath.Join(path, "scenario.yaml")
		}
		s, e := scenario.Load(path)
		if e != nil {
			return e
		}
		if c.Flags().Changed("seed") {
			s.Seed = seed
		}
		if adapterName != "" {
			s.Adapter = adapterName
		}
		if budget < 1 || budget > 1000 {
			return fmt.Errorf("explore must be 1..1000")
		}
		if format != "text" && format != "json" {
			return fmt.Errorf("format must be text or json")
		}
		report, e := engine.Run(c.Context(), s, r)
		if e != nil {
			return e
		}
		if budget > 1 {
			reports, err := engine.Explore(c.Context(), s, r, budget)
			if err != nil {
				return err
			}
			report = reports[len(reports)-1]
		}
		originalReport := report
		original := len(report.Scenario.Actions)
		if minimize && report.Violation != nil {
			report, e = engine.Minimize(c.Context(), report.Scenario, r)
			if e != nil {
				return e
			}
		}
		if e = os.MkdirAll(dir, 0755); e != nil {
			return e
		}
		full, e := json.MarshalIndent(originalReport, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(dir, "run.full.json"), append(full, '\n'), 0644); e != nil {
			return e
		}
		b, e := json.MarshalIndent(report, "", "  ")
		if e != nil {
			return e
		}
		if e = os.WriteFile(filepath.Join(dir, "run.json"), append(b, '\n'), 0644); e != nil {
			return e
		}
		if report.Violation != nil {
			if e = scenario.Save(filepath.Join(dir, "counterexample.yaml"), report.Scenario); e != nil {
				return e
			}
		}
		if format == "json" {
			if _, e = fmt.Fprintln(out, string(b)); e != nil {
				return e
			}
		} else {
			if _, e = fmt.Fprintln(out, Text(report)); e != nil {
				return e
			}
			if minimize && report.Violation != nil {
				fmt.Fprintf(out, "Minimized: %d -> %d actions (replay verified)\n", original, len(report.Scenario.Actions))
			}
			fmt.Fprintf(out, "Saved: %s\n", filepath.Join(dir, "run.json"))
			if report.Violation != nil {
				fmt.Fprintf(out, "Reproduce: reconcilelab replay %s\n", filepath.Join(dir, "counterexample.yaml"))
			}
		}
		if report.Violation != nil {
			return &ExitError{1, fmt.Errorf("invariant violated: %s", report.Violation.Name)}
		}
		return nil
	}
	for _, name := range []string{"test", "replay"} {
		cmd := &cobra.Command{Use: name + " SCENARIO_OR_DIRECTORY", Args: cobra.ExactArgs(1), RunE: run}
		cmd.Flags().Int64Var(&seed, "seed", 42, "Seed for bounded exploration")
		cmd.Flags().StringVar(&format, "format", "text", "text or json")
		cmd.Flags().StringVar(&dir, "output", "results", "Result directory")
		cmd.Flags().StringVar(&adapterName, "adapter", "", "Override registered adapter")
		cmd.Flags().BoolVar(&minimize, "minimize", true, "Shrink violations and verify replay")
		cmd.Flags().IntVar(&budget, "explore", 1, "Seeded permutation budget (1..1000)")
		root.AddCommand(cmd)
	}
	root.AddCommand(&cobra.Command{Use: "report RUN_JSON", Args: cobra.ExactArgs(1), RunE: func(c *cobra.Command, a []string) error {
		b, e := os.ReadFile(a[0])
		if e != nil {
			return e
		}
		var v engine.Report
		if e = json.Unmarshal(b, &v); e != nil {
			return e
		}
		if v.Version == "" {
			return fmt.Errorf("missing report version")
		}
		if _, e = fmt.Fprintln(out, Text(v)); e != nil {
			return e
		}
		for _, step := range v.Steps {
			if _, e = fmt.Fprintf(out, "Action %d: %s -> %s\n", step.Index, step.Action.Kind, step.Result); e != nil {
				return e
			}
			for _, op := range step.Operations {
				if _, e = fmt.Fprintf(out, "  API %s %s: %s\n", op.Verb, op.Key, op.Result); e != nil {
					return e
				}
			}
		}
		if v.Violation != nil {
			evidence, err := json.MarshalIndent(v.Violation, "", "  ")
			if err != nil {
				return err
			}
			_, e = fmt.Fprintln(out, string(evidence))
		}
		return e
	}})
	return root.ExecuteContext(context.Background())
}
func Text(r engine.Report) string {
	status := "PASS"
	if r.Violation != nil {
		status = "VIOLATION DETECTED: " + r.Violation.Name
	}
	return fmt.Sprintf("ReconcileLab %s | local simulation\nSeed: %s | Adapter: %s\n%s\nExecuted: %d actions | API operations: %d", r.Version, strconv.FormatInt(r.Scenario.Seed, 10), r.Scenario.Adapter, status, len(r.Steps), len(r.Operations))
}
func Default() scenario.Scenario {
	return scenario.Scenario{APIVersion: "reconcilelab/v1alpha1", Adapter: "buggy", Seed: 42, Initial: []unstructured.Unstructured{{Object: map[string]interface{}{"apiVersion": "v1", "kind": "ConfigMap", "metadata": map[string]interface{}{"name": "demo", "namespace": "default", "labels": map[string]interface{}{"owner": "another-controller"}}}}}, Actions: []scenario.Action{{Kind: "reconcile", Name: "demo", Namespace: "default"}}}
}
