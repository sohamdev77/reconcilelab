package main

import (
	"fmt"
	"os"

	"github.com/sohamdev77/reconcilelab/internal/cli"
	"github.com/sohamdev77/reconcilelab/pkg/adapter"
)

func main() {
	e := cli.Execute(os.Args[1:], os.Stdout, os.Stderr, adapter.Examples())
	if e != nil {
		fmt.Fprintln(os.Stderr, e)
	}
	os.Exit(cli.Code(e))
}
