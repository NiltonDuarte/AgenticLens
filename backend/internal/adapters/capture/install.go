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
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

//go:embed plugin.ts
var pluginTemplate string

const marker = "AgenticLens capture v1"

type InstallOptions struct {
	Repository  string
	StateDir    string
	Binary      string
	Source      string
	SourceStore string
	DryRun      bool
	Remove      bool
}
type OwnedGroup struct {
	Event string          `json:"event"`
	Value json.RawMessage `json:"value"`
}
type Manifest struct {
	Protocol       int          `json:"protocol"`
	Source         string       `json:"source"`
	Repository     string       `json:"repository"`
	Binary         string       `json:"binary"`
	ConfigPath     string       `json:"config_path"`
	Groups         []OwnedGroup `json:"groups,omitempty"`
	PluginHash     string       `json:"plugin_hash,omitempty"`
	Active         bool         `json:"active"`
	HarnessVersion string       `json:"harness_version"`
	Capability     string       `json:"capability"`
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
	return filepath.Abs(path)
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
	if err == nil && m.Protocol != ProtocolVersion {
		err = errors.New("unsupported install manifest protocol")
	}
	return m, err
}
func jsonEqual(a, b []byte) bool {
	var va, vb bytes.Buffer
	return json.Compact(&va, a) == nil && json.Compact(&vb, b) == nil && bytes.Equal(va.Bytes(), vb.Bytes())
}

func probe(ctx context.Context, source string) (string, string) {
	cmd := exec.CommandContext(ctx, source, "--version")
	b, err := cmd.Output()
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
		features, e := exec.CommandContext(ctx, source, "features", "list").Output()
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
	result := InstallResult{DryRun: options.DryRun, Changes: []PlannedChange{}}
	if options.Source != "codex" && options.Source != "opencode" {
		return result, errors.New("install source must be codex or opencode")
	}
	repo, err := absolute(options.Repository)
	if err != nil {
		return result, err
	}
	dir, err := absolute(options.StateDir)
	if err != nil {
		return result, err
	}
	previous, loadErr := ReadManifest(dir, options.Source)
	if loadErr != nil && !errors.Is(loadErr, os.ErrNotExist) {
		return result, loadErr
	}
	if loadErr == nil && previous.Repository != repo {
		return result, errors.New("capture manifest belongs to another repository")
	}
	if options.Remove && errors.Is(loadErr, os.ErrNotExist) {
		return result, errors.New("no owned installation manifest; refusing removal")
	}
	binary, err := absolute(options.Binary)
	if err != nil {
		return result, err
	}
	store, err := absolute(options.SourceStore)
	if err != nil {
		return result, err
	}
	if !options.Remove {
		info, e := os.Stat(binary)
		if e != nil || info.IsDir() {
			return result, errors.New("native capture handler missing")
		}
		checkCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
		defer cancel()
		out, e := exec.CommandContext(checkCtx, binary, "capture", "protocol").Output()
		if e != nil || strings.TrimSpace(string(out)) != "1" {
			return result, errors.New("installed native handler does not support capture protocol 1")
		}
	}
	m := Manifest{Protocol: ProtocolVersion, Source: options.Source, Repository: repo, Binary: binary, Active: !options.Remove}
	if !options.Remove {
		probeCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
		m.HarnessVersion, m.Capability = probe(probeCtx, options.Source)
		cancel()
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
			args := []string{binary, "capture", "receive", "--source", "codex", "--store", store, "--state-dir", dir}
			command, e := staticCommand(args, false)
			if e != nil {
				return result, e
			}
			windows, e := staticCommand(args, true)
			if e != nil && runtime.GOOS == "windows" {
				return result, e
			}
			for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"} {
				handler := struct {
					Type           string `json:"type"`
					Command        string `json:"command"`
					CommandWindows string `json:"commandWindows,omitempty"`
					Timeout        int    `json:"timeout"`
					Status         string `json:"statusMessage"`
					Limit          int    `json:"additionalContextLimit,omitempty"`
				}{"command", command, windows, 2, marker, 0}
				if event == "SessionStart" || event == "UserPromptSubmit" || event == "SubagentStart" {
					handler.Limit = 1000
				}
				group, e := json.Marshal(struct {
					Hooks []typeofHandler `json:"hooks"`
				}{[]typeofHandler{handler}})
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
		for _, item := range [][2]string{{"__BINARY__", binary}, {"__STATE__", dir}, {"__STORE__", store}} {
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
	// The state directory must be local and ignored. Add an exact relative ignore entry; preserve user content.
	rel, e := filepath.Rel(repo, dir)
	if e != nil {
		return result, e
	}
	if !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && rel != ".." {
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
	for _, c := range changes {
		action := "write"
		if c.remove {
			action = "remove"
		}
		result.Changes = append(result.Changes, PlannedChange{c.path, action})
	}
	result.Manifest = m
	if options.DryRun {
		return result, nil
	}
	// Roll back only files this transaction changed; a failed hook setup must leave unrelated settings intact.
	for i, c := range changes {
		e := apply(c)
		if e != nil {
			rollbackErrors := []error{e}
			for j := i - 1; j >= 0; j-- {
				prior := changes[j]
				prior.content = prior.before
				prior.remove = !prior.existed
				rollbackErrors = append(rollbackErrors, apply(prior))
			}
			return result, errors.Join(rollbackErrors...)
		}
	}
	return result, nil
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
