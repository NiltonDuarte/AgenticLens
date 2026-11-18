package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io"

	"agenticlens/internal/adapters/capture"
	adaptersconfig "agenticlens/internal/adapters/config"
)

// This sandbox-facing entry point deliberately precedes admission, registry
// opening and worker launch. Only the trusted hook or worker accepts a binding.
func captureSubmitBinding(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("capture submit-binding", flag.ContinueOnError)
	flags.SetOutput(output)
	project := flags.String("repo", "", "Exact project directory from the native hook")
	home := flags.String("store", "", "Exact Codex home from the native hook")
	id := flags.String("session-id", "", "Exact session ID from the native hook")
	request := flags.String("request", "", "Exact inbox request delivered by the native hook")
	var ticket capture.BindingTicket
	flags.StringVar(&ticket.ID, "ticket-id", "", "Explicit canonical ticket ID")
	flags.StringVar(&ticket.Title, "ticket-title", "", "Human-readable ticket title")
	flags.StringVar(&ticket.Type, "ticket-type", "", "Optional conventional ticket type")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || *project == "" || *home == "" {
		return errors.New("capture submit-binding requires exact --repo and --store")
	}
	cfg, err := adaptersconfig.Load(*project)
	if err != nil {
		return err
	}
	if !cfg.Telemetry.Capture.Enabled {
		return errors.New("automatic capture is disabled")
	}
	same, err := capture.SameStore(*home, cfg.Telemetry.Sources.CodexHome)
	if err != nil {
		return err
	}
	if !same {
		return errors.New("requested binding store differs from configured Codex home")
	}
	if err := capture.SubmitCodexBinding(*project, *home, *id, *request, ticket, cfg.Ticket); err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(struct {
		State     string `json:"state"`
		SessionID string `json:"session_id"`
	}{"queued", *id})
}
