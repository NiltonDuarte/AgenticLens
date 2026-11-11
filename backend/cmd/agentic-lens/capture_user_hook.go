package main

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex"
	config "agenticlens/internal/adapters/config"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

func nativeCodexUserHook(input io.Reader, output, diagnostic io.Writer) error {
	value, err := capture.ReadHook(input)
	if err != nil {
		return err
	}
	if !filepath.IsAbs(value.WorkingDirectory) {
		return errors.New("native Codex callback requires absolute session cwd")
	}
	shared, err := config.FindConfigFile(value.WorkingDirectory)
	if err != nil {
		return err
	}
	if shared == "" {
		_, err = io.WriteString(output, "{}\n")
		return err
	}
	project := filepath.Dir(shared)
	policy, err := config.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	cfg, err := config.Load(project)
	if err != nil {
		return err
	}
	if _, err := config.ResolveEntrypoint(project); err != nil {
		return err
	}
	store := cfg.Telemetry.Sources.CodexHome
	if strings.TrimSpace(store) == "" {
		return errors.New("selected local configuration requires telemetry.sources.codex.home")
	}
	if !policy.Enabled {
		_, err = io.WriteString(output, "{}\n")
		return err
	}
	runtimeHome := os.Getenv("CODEX_HOME")
	if runtimeHome != "" {
		same, err := capture.SameStore(runtimeHome, store)
		if err != nil {
			return err
		}
		if !same {
			return fmt.Errorf("configured Codex source %s differs from established runtime source %s", store, runtimeHome)
		}
	}
	value, err = capture.NormalizeCodexSubagentStart(value, store, project, policy)
	if err != nil {
		return err
	}
	if value.TranscriptPath != "" {
		if err := codex.ValidateRolloutStore(store, value.TranscriptPath); err != nil {
			return fmt.Errorf("configured Codex source %s and callback transcript %s disagree: %w", store, value.TranscriptPath, err)
		}
		sourceCwd, contextErr := codex.SessionContext(value.TranscriptPath, value.SessionID)
		if contextErr == nil {
			sourceProject, err := config.ProjectRoot(sourceCwd)
			if err != nil {
				return err
			}
			same, err := capture.SameStore(project, sourceProject)
			if err != nil {
				return err
			}
			if !same {
				return fmt.Errorf("Codex session cwd %s conflicts with source context %s", value.WorkingDirectory, sourceCwd)
			}
		} else if errors.Is(contextErr, codex.ErrSessionContextMismatch) {
			return contextErr
		}
	}
	body, err := json.Marshal(value)
	if err != nil {
		return err
	}
	// Source metadata may not exist yet. The durable registration retains the
	// exact pending handoff; the normal bounded worker retries source collection.
	return captureCommandAction([]string{"receive", "--user-level", "--source", "codex", "--repo", project, "--store", store}, bytes.NewReader(body), output)
}
