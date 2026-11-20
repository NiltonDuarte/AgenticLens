package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// nativeOpenCodeHook accepts routing identity only from the plugin's session
// context. Tool arguments carry the ticket response and optional directory.
func nativeOpenCodeHook(args []string, input io.Reader, output io.Writer, protectedHome string) error {
	if len(args) == 0 {
		return errors.New("OpenCode callback requires an operation")
	}
	control := captureControlAction(args[0])
	hook, err := capture.ReadHook(input)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(hook.WorkingDirectory) {
		return errors.New("native OpenCode callback requires absolute session cwd")
	}
	shared, err := config.FindConfigFile(hook.WorkingDirectory)
	if err != nil {
		return err
	}
	if shared == "" {
		if control {
			return errors.New("native OpenCode capture control requires a configured project")
		}
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	project := filepath.Dir(shared)
	policy, err := config.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	if _, err := captureSourceStore("opencode", project); err != nil {
		return err
	}
	if _, err := config.ResolveEntrypoint(project); err != nil {
		return err
	}
	if !policy.Enabled {
		if control {
			return errors.New("automatic capture is disabled for this OpenCode project")
		}
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	if control {
		if hook.SessionID == "" || hook.Event != "agent_capture_control" {
			return errors.New("OpenCode capture control requires native session and operation context")
		}
		return runCaptureControl(args[0], args[1:], output, protectedHome, project, hook.SessionID)
	}
	command := []string{args[0], "--user-level", "--source", "opencode", "--repo", project, "--session-id", hook.SessionID}
	switch args[0] {
	case "receive", "request":
		if len(args) != 1 {
			return errors.New("unexpected OpenCode callback arguments")
		}
	case "bind":
		response := flag.NewFlagSet("OpenCode binding", flag.ContinueOnError)
		response.SetOutput(io.Discard)
		ticket := response.String("ticket-id", "", "Ticket ID")
		title := response.String("ticket-title", "", "Ticket title")
		kind := response.String("ticket-type", "", "Conventional type")
		request := response.String("request", "", "Session request")
		destination := response.String("destination-directory", "", "Publication directory")
		destinationOnly := response.Bool("destination-only", false, "Indicate a destination after binding")
		if err := response.Parse(args[1:]); err != nil {
			return err
		}
		if response.NArg() != 0 || strings.TrimSpace(*ticket) == "" || *request == "" {
			return errors.New("OpenCode response requires ticket ID and session request")
		}
		if !*destinationOnly && (strings.TrimSpace(*title) == "" || strings.TrimSpace(*kind) == "") {
			return errors.New("OpenCode binding requires ticket ID, title, explicit conventional type and session request")
		}
		command = append(command, "--ticket-id", *ticket, "--ticket-title", *title, "--ticket-type", *kind, "--request", *request)
		if *destinationOnly {
			command = append(command, "--destination-only")
		}
		response.Visit(func(f *flag.Flag) {
			if f.Name == "destination-directory" {
				command = append(command, "--destination-directory", *destination)
			}
		})
	default:
		return fmt.Errorf("unknown OpenCode operation %q", args[0])
	}
	body, err := json.Marshal(hook)
	if err != nil {
		return err
	}
	return captureNativeAction(command, bytes.NewReader(body), output, protectedHome, true)
}

// Re-read source selection under the native capture admission before opening
// any registry or source. A runtime-selected database never overrides policy.
func nativeOpenCodeStore(ctx context.Context, project string, cfg domain.ProjectConfig) (string, error) {
	store, err := config.ResolveOpenCodePath("", cfg.Telemetry.Sources.OpenCodeDatabasePath)
	if err != nil {
		return "", err
	}
	if observed := os.Getenv("OPENCODE_DB"); observed != "" {
		if observed == ":memory:" {
			return "", fmt.Errorf("configured OpenCode source %s differs from runtime source :memory:", store)
		}
		if !filepath.IsAbs(observed) {
			base := os.Getenv("XDG_DATA_HOME")
			if base == "" {
				home, err := os.UserHomeDir()
				if err != nil {
					return "", err
				}
				base = filepath.Join(home, ".local", "share")
			}
			if !filepath.IsAbs(base) {
				return "", errors.New("cannot establish absolute runtime OpenCode database from relative XDG_DATA_HOME")
			}
			observed = filepath.Join(base, "opencode", observed)
		}
		same, err := capture.SameStore(store, observed)
		if err != nil {
			return "", err
		}
		if !same {
			return "", fmt.Errorf("configured OpenCode source %s differs from established runtime source %s", store, observed)
		}
	}
	if err := capture.CheckOpenCodeProjectConflict(ctx, project); err != nil {
		return "", err
	}
	return store, nil
}
