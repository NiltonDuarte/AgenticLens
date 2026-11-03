package capture

import (
	"bytes"
	"context"
	"crypto/sha256"
	_ "embed"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
)

//go:embed plugin.ts
var pluginTemplate string

const marker = "AgenticLens capture v1"

type InstallOptions struct {
	Repository          string
	StateDir            string
	Binary              string
	Wrapper             string
	Source              string
	SourceStore         string
	DryRun              bool
	Remove              bool
	AllowLegacyManifest bool
	LegacyStateDir      string
	// RequiredTimeout bounds each required installation phase. The optional
	// harness probe has its own allowance; the caller context still bounds all work.
	RequiredTimeout time.Duration
}
type OwnedGroup struct {
	Event string          `json:"event"`
	Value json.RawMessage `json:"value"`
}
type Manifest struct {
	Protocol        int          `json:"protocol"`
	Source          string       `json:"source"`
	Repository      string       `json:"repository"`
	Binary          string       `json:"binary"`
	Wrapper         string       `json:"wrapper,omitempty"`
	WrapperHash     string       `json:"wrapper_hash,omitempty"`
	Version         string       `json:"version,omitempty"`
	SourceStore     string       `json:"source_store,omitempty"`
	TemplateVersion string       `json:"template_version,omitempty"`
	ConfigPath      string       `json:"config_path"`
	Groups          []OwnedGroup `json:"groups,omitempty"`
	PluginHash      string       `json:"plugin_hash,omitempty"`
	Active          bool         `json:"active"`
	HarnessVersion  string       `json:"harness_version"`
	Capability      string       `json:"capability"`
	StateDir        string       `json:"state_dir,omitempty"`
}
type PlannedChange struct {
	Path   string `json:"path"`
	Action string `json:"action"`
}
type InstallResult struct {
	Manifest Manifest        `json:"manifest"`
	Changes  []PlannedChange `json:"changes"`
	DryRun   bool            `json:"dry_run"`
}
type fileChange struct {
	path    string
	content []byte
	remove  bool
	before  []byte
	existed bool
}

func absolute(path string) (string, error) {
	if path == "" {
		return "", errors.New("empty path")
	}
	return ProjectDirectory(path)
}
func hash(b []byte) string                   { h := sha256.Sum256(b); return hex.EncodeToString(h[:]) }
func manifestPath(dir, source string) string { return filepath.Join(dir, "install-"+source+".json") }
func ReadManifest(dir, source string) (Manifest, error) {
	var m Manifest
	b, err := os.ReadFile(manifestPath(dir, source))
	if err != nil {
		return m, err
	}
	err = json.Unmarshal(b, &m)
	if err == nil && m.Protocol != ManifestProtocolVersion {
		err = errors.New("unsupported install manifest protocol")
	}
	return m, err
}
func jsonEqual(a, b []byte) bool {
	var va, vb bytes.Buffer
	return json.Compact(&va, a) == nil && json.Compact(&vb, b) == nil && bytes.Equal(va.Bytes(), vb.Bytes())
}

func probe(ctx context.Context, source string) (string, string) {
	b, err := installCommandOutput(ctx, source, "--version")
	if err != nil {
		return "unavailable", "Harness unavailable; install it and run capture status"
	}
	version := strings.TrimSpace(string(b))
	expected := "codex-cli 0.166.0"
	if source == "opencode" {
		expected = "1.19.5"
	}
	if version != expected {
		return version, "Unsupported/unverified harness version; no runtime capture compatibility claim"
	}
	if source == "codex" {
		features, e := installCommandOutput(ctx, source, "features", "list")
		if e != nil || !strings.Contains(string(features), "hooks") {
			return version, "Hooks unavailable; inspect codex features list"
		}
	}
	return version, "Configuration protocol inspected; runtime callbacks, trust and agent response require smoke verification"
}
func staticCommand(args []string, windows bool) (string, error) {
	quoted := []string{}
	for _, arg := range args {
		if strings.ContainsAny(arg, "\r\n\x00") {
			return "", errors.New("control character in hook command path")
		}
		if windows {
			if strings.ContainsAny(arg, "% !\"&|<>^()") && strings.ContainsAny(arg, "%!\"&|<>^()") {
				return "", errors.New("unsupported Windows shell metacharacter in hook path")
			}
			quoted = append(quoted, "\""+arg+"\"")
		} else {
			quoted = append(quoted, QuoteArgument(arg))
		}
	}
	return strings.Join(quoted, " "), nil
}
func Install(ctx context.Context, options InstallOptions) (InstallResult, error) {
	return installWithWriter(ctx, options, apply)
}

func installWithWriter(ctx context.Context, options InstallOptions, write func(fileChange) error) (result InstallResult, returnErr error) {
	result = InstallResult{DryRun: options.DryRun, Changes: []PlannedChange{}}
	requiredTimeout := options.RequiredTimeout
	if requiredTimeout == 0 {
		requiredTimeout = 2 * time.Second
	}
	if requiredTimeout < 0 || requiredTimeout > 5*time.Minute {
		return result, errors.New("required installation timeout must be positive and at most five minutes")
	}
	callerCtx := ctx
	if err := callerCtx.Err(); err != nil {
		return result, fmt.Errorf("required installation setup: %w", err)
	}
	ctx, phaseCancel := context.WithTimeout(callerCtx, requiredTimeout)
	defer phaseCancel()
	if options.Source != "codex" && options.Source != "opencode" {
		return result, errors.New("install source must be codex or opencode")
	}
	repo, err := ProjectDirectory(options.Repository)
	if err != nil {
		return result, err
	}
	dir, err := absolute(options.StateDir)
	if err != nil {
		return result, err
	}
	if options.SourceStore == "" {
		return result, errors.New("empty source store")
	}
	if _, err = validateStatePlacement(repo, logicalRegistryDirectory(dir)); err != nil {
		return result, err
	}
	if !options.DryRun {
		lock, e := acquireInstallLock(dir)
		if e != nil {
			return result, e
		}
		defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
	}
	recovered, e := recoverInstallTransaction(ctx, options, repo, dir)
	if e != nil {
		return result, e
	}
	if recovered != nil {
		return *recovered, nil
	}
	expectedLegacyState := ""
	if options.AllowLegacyManifest && options.LegacyStateDir != "" {
		expectedLegacyState, err = ProjectDirectory(options.LegacyStateDir)
		if err != nil {
			return result, err
		}
	}
	legacyStateMatches := func(state string) bool {
		if expectedLegacyState == "" || state == "" {
			return false
		}
		actual, e := ProjectDirectory(state)
		return e == nil && actual == expectedLegacyState
	}
	previous, loadErr := ReadManifest(dir, options.Source)
	if options.AllowLegacyManifest && loadErr != nil && previous.Protocol == 1 {
		body, e := os.ReadFile(manifestPath(dir, options.Source))
		if e != nil {
			return result, e
		}
		var legacy Manifest
		if e = json.Unmarshal(body, &legacy); e != nil {
			return result, e
		}
		if legacy.Protocol != 1 || !legacyStateMatches(legacy.StateDir) {
			return result, errors.New("legacy manifest does not match explicitly validated relocation state")
		}
		previous = legacy
		loadErr = nil
	}
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) {
		return result, loadErr
	}
	previousRepo := ""
	if loadErr == nil {
		previousRepo, err = ProjectDirectory(previous.Repository)
		if err != nil {
			return result, err
		}
	}
	if loadErr == nil && (previousRepo != repo || previous.Source != options.Source) {
		return result, errors.New("capture manifest belongs to another repository")
	}
	if options.Remove && errors.Is(loadErr, os.ErrNotExist) {
		return result, errors.New("no owned installation manifest; refusing removal")
	}
	binary := options.Binary
	wrapper, version := previous.Wrapper, previous.Version
	if options.SourceStore == "" {
		return result, errors.New("empty path")
	}
	store, err := CanonicalStore(options.SourceStore)
	if err != nil {
		return result, err
	}
	if !options.Remove {
		entry, e := adaptersconfig.ResolveEntrypoint(repo)
		if e != nil {
			return result, e
		}
		selected, e := entry.NativePath()
		if e != nil {
			return result, e
		}
		if binary == "" {
			binary = selected
		}
		binary, e = ProjectDirectory(binary)
		if e != nil {
			return result, fmt.Errorf("selected native capture handler: %w", e)
		}
		selected, e = ProjectDirectory(selected)
		if e != nil || binary != selected {
			return result, errors.New("--binary must identify the exact shared-pin native cache leaf; explicitly install the release first")
		}
		if e = validateNativeHeader(binary); e != nil {
			return result, e
		}
		wrapper = options.Wrapper
		if wrapper == "" {
			wrapper, e = adaptersconfig.UserWrapperPath()
			if e != nil {
				return result, e
			}
		}
		wrapper, e = ProjectDirectory(wrapper)
		if e != nil {
			return result, e
		}
		if e = validateCommonWrapper(ctx, wrapper, entry); e != nil {
			return result, e
		}
		version = entry.Version
		info, e := os.Stat(binary)
		if e != nil || info.IsDir() {
			return result, errors.New("native capture handler missing")
		}
		if e := validateInstallNativeOutput(ctx, binary, "3", "native capture handler protocol", "capture", "protocol"); e != nil {
			return result, e
		}
		if e := validateInstallNativeOutput(ctx, binary, version, "native capture handler version", "version", "--short"); e != nil {
			return result, e
		}
	} else {
		binary = previous.Binary
	}
	m := Manifest{Protocol: ManifestProtocolVersion, Source: options.Source, Repository: repo, Binary: binary, Wrapper: wrapper, Version: version, SourceStore: store, TemplateVersion: installationTemplateVersion(), StateDir: dir, Active: !options.Remove}
	if !options.Remove {
		wrapperBody, e := readEntrypointFile(wrapper, 128*1024)
		if e != nil {
			return result, e
		}
		m.WrapperHash = hash(wrapperBody)
	} else {
		m.WrapperHash = previous.WrapperHash
	}
	if !options.Remove {
		if err := ctx.Err(); err != nil {
			return result, fmt.Errorf("required installation planning: %w", err)
		}
		phaseCancel()
		probeCtx, cancel := context.WithTimeout(callerCtx, 3*time.Second)
		m.HarnessVersion, m.Capability = probe(probeCtx, options.Source)
		cancel()
		if err := callerCtx.Err(); err != nil {
			return result, fmt.Errorf("installation canceled during optional harness inspection: %w", err)
		}
		ctx, phaseCancel = context.WithTimeout(callerCtx, requiredTimeout)
		defer phaseCancel()
	} else {
		m.HarnessVersion = previous.HarnessVersion
		m.Capability = "Disabled; existing local registrations retained"
	}
	changes := []fileChange{}
	if options.Source == "codex" {
		path := filepath.Join(repo, ".codex", "hooks.json")
		m.ConfigPath = path
		content, e := os.ReadFile(path)
		exists := e == nil
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return result, e
		}
		config := map[string]json.RawMessage{}
		if exists {
			if e = json.Unmarshal(content, &config); e != nil {
				return result, fmt.Errorf("existing hook configuration invalid; preserved: %w", e)
			}
		}
		hooks := map[string][]json.RawMessage{}
		if raw, ok := config["hooks"]; ok {
			if e = json.Unmarshal(raw, &hooks); e != nil {
				return result, e
			}
		}
		for _, owned := range previous.Groups {
			found := false
			kept := []json.RawMessage{}
			for _, group := range hooks[owned.Event] {
				if jsonEqual(group, owned.Value) {
					found = true
					continue
				}
				kept = append(kept, group)
			}
			if !found {
				return result, fmt.Errorf("owned %s hook was edited or removed; configuration preserved, resolve conflict before upgrade/removal", owned.Event)
			}
			if len(kept) == 0 {
				delete(hooks, owned.Event)
			} else {
				hooks[owned.Event] = kept
			}
		}
		if !options.Remove {
			for _, groups := range hooks {
				for _, group := range groups {
					if strings.Contains(string(group), marker) {
						return result, errors.New("untracked or locally edited AgenticLens hook; refusing overwrite")
					}
				}
			}
			args := WrapperCommand(wrapper, "--al-offline", "--al-project", repo, "capture", "receive", "--source", "codex", "--store", store, "--state-dir", logicalRegistryDirectory(dir), "--repo", repo, "--wrapper", wrapper)
			command, e := staticCommand(args, false)
			if e != nil {
				return result, e
			}
			windows := ""
			if runtime.GOOS == "windows" {
				windows, e = staticCommand(args, true)
			}
			if e != nil && runtime.GOOS == "windows" {
				return result, e
			}
			for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"} {
				group, e := codexHookGroup(event, command, windows, runtime.GOOS == "windows")
				if e != nil {
					return result, e
				}
				hooks[event] = append(hooks[event], group)
				m.Groups = append(m.Groups, OwnedGroup{event, group})
			}
		}
		raw, e := json.Marshal(hooks)
		if e != nil {
			return result, e
		}
		config["hooks"] = raw
		after, e := json.MarshalIndent(config, "", "  ")
		if e != nil {
			return result, e
		}
		after = append(after, '\n')
		changes = append(changes, fileChange{path: path, content: after, before: content, existed: exists})
	} else {
		path := filepath.Join(repo, ".opencode", "plugins", "agenticlens-capture.ts")
		m.ConfigPath = path
		before, e := os.ReadFile(path)
		exists := e == nil
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return result, e
		}
		if exists && (previous.PluginHash == "" || hash(before) != previous.PluginHash) {
			return result, errors.New("plugin contains conflicting local edits; preserved")
		}
		replacements := []string{}
		for _, item := range [][2]string{{"__WRAPPER__", wrapper}, {"__STATE__", logicalRegistryDirectory(dir)}, {"__STORE__", store}, {"__PROJECT__", repo}} {
			raw, e := json.Marshal(item[1])
			if e != nil {
				return result, e
			}
			replacements = append(replacements, item[0], string(raw))
		}
		// Expand only the original template. A marker in a path is literal data.
		content := []byte(strings.NewReplacer(replacements...).Replace(pluginTemplate))
		if !options.Remove {
			m.PluginHash = hash(content)
		}
		changes = append(changes, fileChange{path: path, content: content, remove: options.Remove, before: before, existed: exists})
	}
	// Repository-local logical state includes registry.db, SQLite sidecars and
	// every coordination leaf. External state does not need a project ignore.
	rel, contained, e := relativeContainedPath(repo, logicalRegistryDirectory(dir))
	if e != nil {
		return result, e
	}
	if contained {
		if rel == "." {
			return result, errors.New("capture state cannot be the repository root")
		}
		ignore := filepath.Join(repo, ".gitignore")
		before, readErr := os.ReadFile(ignore)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return result, readErr
		}
		line := "/" + filepath.ToSlash(rel) + "/"
		if strings.ContainsAny(line, "\n\r*?[]!\\") {
			return result, errors.New("unsupported capture state ignore path")
		}
		present := false
		for _, v := range strings.Split(string(before), "\n") {
			if v == line {
				present = true
			}
		}
		if !present && !options.Remove {
			after := append(append([]byte{}, before...), []byte("\n# Local AgenticLens lifecycle state\n"+line+"\n")...)
			changes = append(changes, fileChange{path: ignore, content: after, before: before, existed: readErr == nil})
		}
	}
	body, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return result, err
	}
	manifestFile := manifestPath(dir, options.Source)
	old, oldErr := os.ReadFile(manifestFile)
	if oldErr != nil && !errors.Is(oldErr, os.ErrNotExist) {
		return result, oldErr
	}
	changes = append(changes, fileChange{path: manifestFile, content: body, before: old, existed: oldErr == nil})

	anchor := filepath.Join(repo, "."+options.Source, "agenticlens-installation.json")
	anchorBefore, anchorErr := os.ReadFile(anchor)
	if anchorErr != nil && !errors.Is(anchorErr, os.ErrNotExist) {
		return result, anchorErr
	}
	if anchorErr == nil {
		var prior Manifest
		if err := json.Unmarshal(anchorBefore, &prior); err != nil {
			return result, err
		}
		priorRepo, err := ProjectDirectory(prior.Repository)
		protocolMatches := prior.Protocol == ManifestProtocolVersion
		if options.AllowLegacyManifest && prior.Protocol == 1 && legacyStateMatches(prior.StateDir) {
			protocolMatches = true
		}
		if err != nil || priorRepo != repo || prior.Source != options.Source || !protocolMatches {
			return result, errors.New("installation discovery file ownership conflict")
		}
	}
	changes = append(changes, fileChange{path: anchor, content: body, before: anchorBefore, existed: anchorErr == nil})
	for _, c := range changes {
		action := "write"
		if c.remove {
			action = "remove"
		}
		result.Changes = append(result.Changes, PlannedChange{c.path, action})
	}
	result.Manifest = m
	if err := ctx.Err(); err != nil {
		return result, fmt.Errorf("required installation planning: %w", err)
	}
	if options.DryRun {
		return result, nil
	}
	if err = applyInstallTransactionWithWriter(ctx, options, repo, dir, m, changes, write); err != nil {
		return result, err
	}
	return result, nil
}

// The native handoff remains two seconds. Native Windows additionally needs
// bounded PowerShell startup time before that handoff begins.
func codexHookGroup(event, command, windowsCommand string, windows bool) (json.RawMessage, error) {
	timeout := 2
	if windows {
		timeout = 5
	}
	handler := typeofHandler{"command", command, windowsCommand, timeout, marker, 0}
	if event == "SessionStart" || event == "UserPromptSubmit" || event == "SubagentStart" {
		handler.Limit = 1000
	}
	return json.Marshal(struct {
		Hooks []typeofHandler `json:"hooks"`
	}{[]typeofHandler{handler}})
}

type typeofHandler = struct {
	Type           string `json:"type"`
	Command        string `json:"command"`
	CommandWindows string `json:"commandWindows,omitempty"`
	Timeout        int    `json:"timeout"`
	Status         string `json:"statusMessage"`
	Limit          int    `json:"additionalContextLimit,omitempty"`
}

func apply(c fileChange) error {
	if c.remove {
		err := os.Remove(c.path)
		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
		return err
	}
	if err := os.MkdirAll(filepath.Dir(c.path), 0700); err != nil {
		return err
	}
	file, err := os.CreateTemp(filepath.Dir(c.path), ".agenticlens-write-")
	if err != nil {
		return err
	}
	name := file.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "capture temporary cleanup failed:", e)
		}
	}()
	if _, err = file.Write(c.content); err != nil {
		return errors.Join(err, file.Close())
	}
	if err = file.Sync(); err != nil {
		return errors.Join(err, file.Close())
	}
	if err = file.Close(); err != nil {
		return err
	}
	return os.Rename(name, c.path)
}
