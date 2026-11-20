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
	controlAction := flags.String("capture-action", "", "Discard or restore an exact capture through its native request")
	captureID := flags.String("capture-id", "", "Exact capture ID supplied by the native control invitation")
	destinationOnly := flags.Bool("destination-only", false, "Indicate a destination for the already bound ticket using its separate request")
	var ticket capture.BindingTicket
	flags.StringVar(&ticket.ID, "ticket-id", "", "Explicit canonical ticket ID")
	flags.StringVar(&ticket.Title, "ticket-title", "", "Human-readable ticket title")
	flags.StringVar(&ticket.Type, "ticket-type", "", "Required conventional ticket type when binding or completing ticket metadata")
	flags.StringVar(&ticket.Destination, "destination-directory", "", "Exact existing directory to indicate for first publication")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 || *project == "" || *home == "" {
		return errors.New("capture submit-binding requires exact --repo and --store")
	}
	provided := make(map[string]bool)
	flags.Visit(func(f *flag.Flag) { provided[f.Name] = true })
	control := provided["capture-action"] || provided["capture-id"]
	if control {
		for _, name := range []string{"ticket-id", "ticket-title", "ticket-type", "destination-directory", "destination-only"} {
			if provided[name] {
				return errors.New("native capture control requires its separate request and cannot include binding arguments")
			}
		}
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
	if control {
		err = capture.SubmitCodexCaptureControl(*project, *home, *id, *captureID, *controlAction, *request)
	} else if *destinationOnly {
		err = capture.SubmitCodexDestination(*project, *home, *id, *request, ticket, cfg.Ticket)
	} else {
		err = capture.SubmitCodexBinding(*project, *home, *id, *request, ticket, cfg.Ticket)
	}
	if err != nil {
		return err
	}
	return json.NewEncoder(output).Encode(struct {
		State     string `json:"state"`
		SessionID string `json:"session_id"`
	}{"queued", *id})
}
