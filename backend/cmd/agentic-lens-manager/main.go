package main

import (
	"fmt"
	"os"
	"os/exec"

	"agenticlens/internal/adapters/manager"
)

func main() {
	if err := manager.Run(os.Args[1:], os.Stdin, os.Stdout, os.Stderr); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			os.Exit(exit.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "AgenticLens manager:", err)
		os.Exit(1)
	}
}
