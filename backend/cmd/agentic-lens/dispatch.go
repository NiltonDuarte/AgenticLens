package main

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/codex"
	config "agenticlens/internal/adapters/config"
)

// nativeDispatch is the installed machine-wide entrypoint. Selection is native,
// offline and identical for manual invocations and session-context callbacks.
func nativeDispatch(args []string, input io.Reader, output, diagnostic io.Writer) error {
	return dispatchNative(args, input, output, diagnostic, false)
}

func dispatchNative(args []string, input io.Reader, output, diagnostic io.Writer, automatic bool) error {
	if len(args) == 1 && args[0] == "protocol" {
		_, err := io.WriteString(output, "1\n")
		return err
	}
	if len(args) > 0 && args[0] == "codex-hook" {
		return nativeCodexHook(input, output, diagnostic)
	}
	project := "."
	resolve := false
	resolveCache := false
	for len(args) > 0 {
		if args[0] == "--al-offline" {
			args = args[1:]
			continue
		}
		if args[0] == "--al-resolve-cache" {
			resolveCache = true
			args = args[1:]
			continue
		}
		if args[0] == "--al-resolve" {
			resolve = true
			args = args[1:]
			continue
		}
		if args[0] == "--al-project" {
			if len(args) < 2 {
				return errors.New("--al-project requires a directory")
			}
			project = args[1]
			args = args[2:]
			continue
		}
		break
	}
	entry, err := config.ResolveEntrypoint(project)
	if err != nil {
		return err
	}
	if resolveCache {
		native, err := entry.CachedNativePath()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "%s\n%s\n%s\n%s\n", entry.Version, entry.Project, entry.InstallRoot, native)
		return err
	}
	native, err := entry.NativePath()
	if err != nil {
		return err
	}
	validate := config.ValidateManualExecutable
	if automatic || (entry.BinaryDir != "" && strings.HasPrefix(native, entry.BinaryDir+string(filepath.Separator))) {
		validate = config.ValidateNativeExecutable
	}
	identity, err := validate(native)
	if err != nil {
		return fmt.Errorf("selected native %s (exact fallback %s); explicit installation required if absent: %w", native, entry.Version, err)
	}
	if entry.BinaryDir == "" || !strings.HasPrefix(native, entry.BinaryDir+string(filepath.Separator)) {
		if identity.Version != entry.Version {
			return fmt.Errorf("cached native identity %s differs from exact pin %s", identity.Version, entry.Version)
		}
	}
	if resolve {
		_, err = fmt.Fprintf(output, "%s\n%s\n%s\n%s\n", identity.Version, entry.Project, entry.InstallRoot, native)
		return err
	}
	if !automatic && input == os.Stdin && output == os.Stdout && diagnostic == os.Stderr {
		if replaced, err := execManualNative(native, args); replaced || err != nil {
			return err
		}
	}
	command := exec.Command(native, args...)
	command.Stdin = input
	command.Stdout = output
	command.Stderr = diagnostic
	if automatic {
		command.Dir = entry.Project
	}
	return command.Run()
}

func nativeCodexHook(input io.Reader, output, diagnostic io.Writer) error {
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
	if shared == "" || filepath.Base(shared) != ".agenticlens-config.yaml" {
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
	if !policy.Enabled {
		_, err = io.WriteString(output, "{}\n")
		return err
	}
	store := cfg.Telemetry.Sources.CodexHome
	if strings.TrimSpace(store) == "" {
		return errors.New("selected local configuration requires telemetry.sources.codex.home")
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
	value, err = capture.NormalizeCodexSubagentStart(value, store, project)
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
	return dispatchNative([]string{"--al-offline", "--al-project", project, "capture", "receive", "--user-level", "--source", "codex", "--repo", project, "--store", store}, bytes.NewReader(body), output, diagnostic, true)
}

func runNativeDispatch(args []string) {
	executable, err := os.Executable()
	if err == nil {
		active, selectionErr := capture.ActiveDispatcher(executable)
		if selectionErr != nil {
			fmt.Fprintln(os.Stderr, "AgenticLens dispatcher:", selectionErr)
			os.Exit(1)
		}
		if active != "" {
			command := exec.Command(active, append([]string{"dispatch"}, args...)...)
			command.Stdin = os.Stdin
			command.Stdout = os.Stdout
			command.Stderr = os.Stderr
			if err := command.Run(); err != nil {
				var exit *exec.ExitError
				if errors.As(err, &exit) {
					os.Exit(exit.ExitCode())
				}
				fmt.Fprintln(os.Stderr, err)
				os.Exit(1)
			}
			return
		}
	} else {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	if err := nativeDispatch(args, os.Stdin, os.Stdout, os.Stderr); err != nil {
		fmt.Fprintln(os.Stderr, "AgenticLens dispatcher:", err)
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			os.Exit(exit.ExitCode())
		}
		os.Exit(1)
	}
}
