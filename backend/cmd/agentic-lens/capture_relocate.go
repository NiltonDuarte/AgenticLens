package main

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"io"
	"time"

	"agenticlens/internal/adapters/capture"
)

func relocateCaptureCommand(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("capture relocate-state", flag.ContinueOnError)
	flags.SetOutput(output)
	project := flags.String("repo", ".", "Exact existing worktree to relocate")
	from := flags.String("from-state-dir", "", "Exact legacy registry directory")
	to := flags.String("to-state-dir", "", "Explicit new configured state base")
	dryRun := flags.Bool("dry-run", false, "Preview without modifying state or configuration")
	timeout := flags.Duration("timeout", 30*time.Second, "Bound maintenance duration")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || *from == "" || *to == "" || *timeout <= 0 || *timeout > 5*time.Minute {
		return errors.New("relocate-state requires explicit --from-state-dir and --to-state-dir and a positive bounded timeout")
	}
	ctx, cancel := context.WithTimeout(context.Background(), *timeout)
	defer cancel()
	result, err := capture.RelocateState(ctx, capture.RelocationOptions{Project: *project, FromDirectory: *from, ToBase: *to, DryRun: *dryRun})
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(result)
}
func inspectCaptureStateCommand(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("capture inspect-state", flag.ContinueOnError)
	flags.SetOutput(output)
	registry := flags.String("state-dir", "", "Exact logical registry directory")
	name := flags.String("project-name", "", "Exact configured name retained by this registry")
	project := flags.String("repo", "", "Exact recorded physical worktree, which may no longer exist")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || *registry == "" || *name == "" || *project == "" {
		return errors.New("inspect-state requires explicit --state-dir, --project-name and --repo; no configuration fallback")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	result, err := capture.InspectRetainedState(ctx, *registry, *name, *project)
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(result)
}
