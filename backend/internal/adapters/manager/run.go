// Package manager installs and selects native engines without depending on the
// capture engine or its embedded dashboard.
package manager

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
)

const Protocol = 1

// Run executes manager maintenance commands or selects one offline engine.
func Run(args []string, input io.Reader, output, diagnostic io.Writer) error {
	if len(args) == 2 && args[0] == "engine" && args[1] == "protocol" {
		return errors.New("this executable is the AgenticLens manager, not an engine")
	}
	if len(args) == 2 && args[0] == "manager" && (args[1] == "opencode-hook-protocol" || args[1] == "codex-binding-wake-protocol" || args[1] == "protected-codex-hook-protocol") {
		_, err := fmt.Fprintln(output, "1")
		return err
	}
	if len(args) == 3 && args[0] == "manager" && args[1] == "protected-codex-version" {
		version, err := config.ProtectedCodexVersion(args[2], domain.GetBuildInfo().Version)
		if err != nil {
			return err
		}
		_, err = fmt.Fprintln(output, version)
		return err
	}
	if len(args) == 2 && args[0] == "manager" && args[1] == "protocol" {
		_, err := fmt.Fprintln(output, Protocol)
		return err
	}
	if len(args) == 2 && args[0] == "manager" && (args[1] == "--help" || args[1] == "help") {
		_, err := io.WriteString(output, "AgenticLens manager\n  install EXACT [--local-bin PATH --sha256 SHA] [--install-root DIR] [--platform OS-ARCH] [--force]\n  self-update --version EXACT [--local-bin PATH --sha256 SHA] [--dest FILE]\n  versions [--install-root DIR]\n  --manager-version [--json|--short]\nOther commands run the project-selected engine. Normal execution is offline.\n")
		return err
	}
	if len(args) > 0 && args[0] == "--manager-version" {
		info := domain.GetBuildInfo()
		if len(args) > 1 && args[1] == "--json" {
			return json.NewEncoder(output).Encode(info)
		}
		_, err := fmt.Fprintln(output, info.Version)
		return err
	}
	project, resolve, cache := ".", false, false
	for len(args) > 0 {
		switch args[0] {
		case "--al-offline":
			args = args[1:]
		case "--al-project":
			if len(args) < 2 {
				return errors.New("--al-project requires a directory")
			}
			project, args = args[1], args[2:]
		case "--al-resolve":
			resolve, args = true, args[1:]
		case "--al-resolve-cache":
			cache, args = true, args[1:]
		default:
			goto parsed
		}
	}
parsed:
	if len(args) > 0 {
		switch args[0] {
		case "self-update":
			return selfUpdate(args[1:], output)
		case "install":
			return installEngine(project, args[1:], output)
		case "versions":
			return listVersions(project, args[1:], output)
		case "codex-binding-wake":
			return codexBindingWake(input, output, diagnostic)
		case "codex-hook":
			return nativeHook("Codex", []string{"capture", "user-hook"}, input, output, diagnostic)
		case "opencode-hook":
			if len(args) < 2 {
				return errors.New("opencode-hook requires receive, request or bind")
			}
			return nativeHook("OpenCode", append([]string{"capture", "opencode-hook"}, args[1:]...), input, output, diagnostic)
		}
	}
	entry, err := resolveProject(project, standalone(args) && !resolve && !cache)
	if err != nil {
		return err
	}
	args, codexHomes, err := prepareCaptureArgs(args, entry.Project)
	if err != nil {
		return err
	}
	if cache {
		path, err := entry.CachedNativePath()
		if err != nil {
			return err
		}
		_, err = fmt.Fprintf(output, "%s\n%s\n%s\n%s\n", entry.Version, entry.Project, entry.InstallRoot, path)
		return err
	}
	if len(codexHomes) > 1 && !resolve {
		manager, err := os.Executable()
		if err != nil {
			return err
		}
		native, err := config.ResolveProtectedCodexEngine(entry.Project, codexHomes[1], manager)
		if err != nil {
			return err
		}
		return execute(native, args, input, output, diagnostic, "")
	}
	native, identity, err := selectEngine(entry, codexHomes...)
	if err != nil {
		return err
	}
	if resolve {
		_, err = fmt.Fprintf(output, "%s\n%s\n%s\n%s\n", identity.Version, entry.Project, entry.InstallRoot, native)
		return err
	}
	return execute(native, args, input, output, diagnostic, "")
}

// Capture flags use Go flag arity. Consume values before interpreting later
// tokens so a literal "--repo" title cannot change project selection.
func prepareCaptureArgs(args []string, selected string) ([]string, []string, error) {
	if len(args) < 2 || args[0] != "capture" {
		return args, nil, nil
	}
	repository, source, store, userSource := "", "", "", ""
	stateDirectory, projectName := "", ""
	for index := 2; index < len(args); index++ {
		argument := args[index]
		if argument == "--" || !strings.HasPrefix(argument, "-") {
			break
		}
		name, value, inline := strings.Cut(strings.TrimLeft(argument, "-"), "=")
		if name == "user-source" || name == "store" || name == "repo" || name == "state-dir" || name == "project-name" || name == "source" {
			if !inline {
				index++
				if index >= len(args) {
					return nil, nil, fmt.Errorf("capture --%s requires a value", name)
				}
				value = args[index]
			}
			switch name {
			case "user-source":
				userSource = value
			case "store":
				store = value
			case "source":
				source = value
			case "repo":
				repository = value
			case "state-dir":
				stateDirectory = value
			case "project-name":
				projectName = value
			}
			continue
		}
		if inline {
			continue
		}
		switch name {
		case "dry-run", "user-level", "help", "h":
		default:
			index++
		}
	}
	// Earlier engines accepted --source opencode on user installation but
	// operated on Codex. Send an option they reject during flag parsing.
	if source == "opencode" && (args[1] == "install-user" || args[1] == "uninstall-user") {
		args = slices.Insert(slices.Clone(args), 2, "--user-source", source)
	}
	var homes []string
	if args[1] == "install-user" {
		homes = []string{config.RuntimeCodexHome()}
		if source != "opencode" && userSource != "opencode" {
			if store == "" {
				cfg, err := config.Load(selected)
				if err != nil {
					return nil, nil, err
				}
				store = cfg.Telemetry.Sources.CodexHome
			}
			homes = append(homes, store)
		}
	}
	if repository == "" {
		return args, homes, nil
	}
	// Retained-state inspection validates the explicit registry and stored
	// project identity in the engine, even when its checkout no longer exists.
	if args[1] == "inspect-state" && stateDirectory != "" && projectName != "" {
		return args, homes, nil
	}
	root, err := config.ProjectRoot(repository)
	if err != nil {
		return nil, nil, err
	}
	wanted, err := os.Stat(root)
	if err != nil {
		return nil, nil, err
	}
	actual, err := os.Stat(selected)
	if err != nil {
		return nil, nil, err
	}
	if !os.SameFile(wanted, actual) {
		return nil, nil, errors.New("capture --repo differs from selected project; supply matching --al-project DIR before capture")
	}
	return args, homes, nil
}

func defaultRoot() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".agenticlens"), nil
}

func resolveProject(project string, allowCompanion bool) (config.Entrypoint, error) {
	shared, err := config.FindConfigFile(project)
	if err != nil {
		return config.Entrypoint{}, err
	}
	if shared != "" {
		return config.ResolveEntrypoint(project)
	}
	if !allowCompanion {
		return config.ResolveEntrypoint(project)
	}
	root, err := defaultRoot()
	if err != nil {
		return config.Entrypoint{}, err
	}
	info := domain.GetBuildInfo()
	if err := config.ValidateReleaseVersion(info.Version); err != nil {
		return config.Entrypoint{}, err
	}
	return config.Entrypoint{Version: info.Version, InstallRoot: root}, nil
}

func standalone(args []string) bool {
	if len(args) == 0 {
		return true
	}
	switch args[0] {
	case "help", "--help", "-h", "-help", "version", "--version", "-version", "-v", "init":
		return true
	default:
		return false
	}
}

func selectEngine(entry config.Entrypoint, sourceHome ...string) (string, domain.BuildInfo, error) {
	native, err := entry.NativePath(sourceHome...)
	if err != nil {
		return "", domain.BuildInfo{}, err
	}
	identity, err := config.ValidateNativeExecutable(native)
	if err != nil {
		return "", identity, fmt.Errorf("engine %s unavailable or incompatible; explicitly install %s: %w", native, entry.Version, err)
	}
	development := entry.BinaryDir != ""
	if !development && identity.Version != entry.Version {
		return "", identity, fmt.Errorf("cached engine identity %s differs from exact pin %s", identity.Version, entry.Version)
	}
	return native, identity, nil
}

func execute(path string, args []string, input io.Reader, output, diagnostic io.Writer, directory string) error {
	if input == os.Stdin && output == os.Stdout && diagnostic == os.Stderr {
		if replaced, err := replaceProcess(path, args, directory); replaced || err != nil {
			return err
		}
	}
	command := exec.Command(path, args...)
	command.Stdin, command.Stdout, command.Stderr, command.Dir = input, output, diagnostic, directory
	return command.Run()
}

func nativeHook(harness string, engineArgs []string, input io.Reader, output, diagnostic io.Writer) error {
	body, err := io.ReadAll(io.LimitReader(input, 65537))
	if err != nil {
		return err
	}
	if len(body) > 65536 {
		return fmt.Errorf("%s hook input exceeds 64 KiB", harness)
	}
	var event struct {
		WorkingDirectory string `json:"cwd"`
		SessionID        string `json:"session_id"`
		Event            string `json:"hook_event_name"`
	}
	if err := json.Unmarshal(body, &event); err != nil {
		return fmt.Errorf("%s hook context: %w", harness, err)
	}
	if !filepath.IsAbs(event.WorkingDirectory) {
		return fmt.Errorf("native %s callback requires absolute session cwd", harness)
	}
	if event.SessionID == "" || event.Event == "" {
		return errors.New("hook requires session_id and hook_event_name")
	}
	shared, err := config.FindConfigFile(event.WorkingDirectory)
	if err != nil {
		return err
	}
	if shared == "" {
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	project := filepath.Dir(shared)
	policy, err := config.ResolveCapture(project, "")
	if err != nil {
		return err
	}
	configuration, err := config.Load(project)
	if err != nil {
		return err
	}
	if harness == "OpenCode" && strings.TrimSpace(configuration.Telemetry.Sources.OpenCodeDatabasePath) == "" {
		return errors.New("selected local configuration requires telemetry.sources.opencode.database_path")
	}
	if harness == "Codex" && strings.TrimSpace(configuration.Telemetry.Sources.CodexHome) == "" {
		return errors.New("selected local configuration requires telemetry.sources.codex.home")
	}
	if !policy.Enabled {
		_, err := io.WriteString(output, "{}\n")
		return err
	}
	home := config.RuntimeCodexHome()
	if harness == "Codex" {
		actual, err := filesystem.PhysicalPath(home)
		if err != nil {
			return err
		}
		configured, err := filesystem.PhysicalPath(configuration.Telemetry.Sources.CodexHome)
		if err != nil {
			return err
		}
		if actual != configured {
			return fmt.Errorf("configured Codex source %s differs from established runtime source %s", configured, actual)
		}
	}
	if harness != "Codex" {
		err := config.WithOrdinaryCapture(project, func() error {
			entry, err := config.ResolveEntrypoint(project)
			if err != nil {
				return err
			}
			native, _, err := selectEngine(entry, home, configuration.Telemetry.Sources.CodexHome)
			if err != nil {
				return err
			}
			return execute(native, engineArgs, bytes.NewReader(body), output, diagnostic, project)
		})
		if !errors.Is(err, config.ErrProtectedCodexProject) {
			return err
		}
		home, err = config.ProtectedCodexProjectHome(project)
		if err != nil {
			return err
		}
		if home == "" {
			return errors.New("protected admission changed during callback; retry after checking operator setup")
		}
	}
	actual, err := os.Executable()
	if err != nil {
		return err
	}
	if _, err := config.ProtectedCodexManager(project, home, actual); err != nil {
		return err
	}
	native, err := config.ResolveProtectedCodexEngine(project, home)
	if err != nil {
		return err
	}
	if harness == "Codex" {
		err := config.PrepareProtectedCodexProject(project, home, func() error {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			cmd := exec.CommandContext(ctx, native, "capture", "admit-protected-worktree", "--repo", project, "--protected-codex-home", home)
			cmd.Dir, cmd.Stderr = project, diagnostic
			if err := cmd.Run(); err != nil {
				return fmt.Errorf("protected capture admission failed; check the current worker or install a compatible protected engine: %w", err)
			}
			return nil
		})
		if err != nil {
			return err
		}
	}
	return execute(native, append(engineArgs, "--protected-codex-home", home), bytes.NewReader(body), output, diagnostic, project)

}

func listVersions(project string, args []string, output io.Writer) error {
	entry, err := resolveProject(project, true)
	if err != nil {
		return err
	}
	if len(args) > 0 {
		if len(args) != 2 || args[0] != "--install-root" {
			return errors.New("usage: versions [--install-root DIR]")
		}
		entry.InstallRoot, err = validLocalPath(args[1])
		if err != nil {
			return err
		}
	}
	items, err := os.ReadDir(filepath.Join(entry.InstallRoot, "bin"))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	versions := []string{}
	for _, item := range items {
		if !item.IsDir() || config.ValidateReleaseVersion(item.Name()) != nil {
			continue
		}
		platforms, err := os.ReadDir(filepath.Join(entry.InstallRoot, "bin", item.Name()))
		if err != nil {
			return err
		}
		for _, platform := range platforms {
			if !platform.IsDir() {
				continue
			}
			osName, arch, err := parsePlatform(platform.Name())
			if err != nil {
				continue
			}
			path := filepath.Join(entry.InstallRoot, "bin", item.Name(), platform.Name(), platformArtifactName("agentic-lens-engine", osName, arch))
			if err := config.ValidateNativePlatform(path, osName, arch); err == nil {
				versions = append(versions, item.Name()+"/"+platform.Name())
			}
		}
	}
	result := struct {
		Manager     domain.BuildInfo `json:"manager"`
		Engines     []string         `json:"engines"`
		InstallRoot string           `json:"install_root"`
	}{domain.GetBuildInfo(), versions, entry.InstallRoot}
	return json.NewEncoder(output).Encode(result)
}

func validLocalPath(value string) (string, error) {
	if strings.ContainsAny(value, "\x00\r\n") {
		return "", errors.New("path contains control characters")
	}
	return filesystem.PhysicalPath(value)
}
