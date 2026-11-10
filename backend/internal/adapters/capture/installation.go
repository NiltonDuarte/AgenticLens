package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
)

const installationHealthFileLimit = 1024 * 1024

var errInstallationOwnership = errors.New("capture installation ownership conflict")

// InstalledStates discovers owned installation state from stable project files.
// Legacy hook/plugin content is read as data, never executed or broadly imported.
func InstalledStates(project string) ([]string, error) {
	states := []string{}
	for _, source := range []string{"codex", "opencode"} {
		anchor := filepath.Join(project, "."+source, "agenticlens-installation.json")
		body, err := readEntrypointFile(anchor, installationHealthFileLimit)
		state := ""
		if err == nil {
			var m Manifest
			if err = json.Unmarshal(body, &m); err != nil {
				return nil, err
			}
			owner, e := ProjectDirectory(m.Repository)
			if e != nil {
				return nil, fmt.Errorf("%w: resolve discovery project: %w", errInstallationOwnership, e)
			}
			if owner != project || m.Source != source {
				return nil, fmt.Errorf("%w: discovery project/source differs", errInstallationOwnership)
			}
			if m.Protocol != ManifestProtocolVersion {
				return nil, errors.New("installed integration protocol differs; refresh owned setup")
			}
			state = m.StateDir
		} else if !errors.Is(err, os.ErrNotExist) {
			return nil, err
		}
		if state == "" {
			path := filepath.Join(project, ".codex", "hooks.json")
			pattern := `--state-dir\s+'([^']*)'`
			if source == "opencode" {
				path = filepath.Join(project, ".opencode", "plugins", "agenticlens-capture.ts")
				pattern = `const stateDir = ("(?:[^"\\]|\\.)*")`
			}
			body, err = readEntrypointFile(path, installationHealthFileLimit)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return nil, err
			}
			if source == "codex" {
				var config struct {
					Hooks map[string][]struct {
						Hooks []struct {
							Command string `json:"command"`
							Status  string `json:"statusMessage"`
						} `json:"hooks"`
					} `json:"hooks"`
				}
				if err := json.Unmarshal(body, &config); err != nil {
					return nil, err
				}
				for _, groups := range config.Hooks {
					for _, group := range groups {
						for _, h := range group.Hooks {
							if h.Status != marker {
								continue
							}
							if candidate, ok := installedStateArgument(h.Command); ok {
								state = candidate
							}
						}
					}
				}
			} else if match := regexp.MustCompile(pattern).FindSubmatch(body); len(match) > 1 {
				if err := json.Unmarshal(match[1], &state); err != nil {
					return nil, err
				}
			}
		}
		if state == "" {
			continue
		}
		manifestBody, err := readEntrypointFile(manifestPath(state, source), installationHealthFileLimit)
		var m Manifest
		if err == nil {
			err = json.Unmarshal(manifestBody, &m)
		}
		if err == nil && m.Protocol != ManifestProtocolVersion {
			err = errors.New("unsupported install manifest protocol")
		}
		if err != nil {
			return nil, fmt.Errorf("installed %s state %s unavailable: %w; restore its manifest before reinstall", source, state, err)
		}
		owner, err := ProjectDirectory(m.Repository)
		if err != nil {
			return nil, fmt.Errorf("%w: resolve installed state project: %w", errInstallationOwnership, err)
		}
		if owner != project || m.Source != source {
			return nil, fmt.Errorf("%w: installed state belongs to another project/source", errInstallationOwnership)
		}
		states = append(states, state)
	}
	return states, nil
}

// InstallationIssues separates current installation drift from fatal namespace
// ownership conflicts without printing or altering historical exchange facts.
// Strict runtime callers continue to use CheckInstalledState.
func InstallationIssues(project, state string) ([]string, error) {
	err := CheckInstalledState(project, state, false)
	if err == nil {
		return []string{}, nil
	}
	if errors.Is(err, errInstallationOwnership) {
		return nil, err
	}
	return []string{err.Error()}, nil
}

// CheckInstalledState diagnoses frozen installed paths without overriding an
// intentional CLI state selection or copying registrations between namespaces.
func CheckInstalledState(project, state string, explicit bool) error {
	states, err := InstalledStates(project)
	if err != nil {
		return err
	}
	for _, installed := range states {
		canonical, err := ProjectDirectory(installed)
		if err != nil {
			return err
		}
		expected, err := ProjectDirectory(state)
		if err != nil {
			return err
		}
		if canonical != expected {
			detail := fmt.Sprintf("installed capture state %s differs from configured/selected state %s; restore telemetry.capture.state_dir to the installed base or use --state-dir %s for status/recovery and owned reinstall/removal; preserve the old registry", installed, state, installed)
			if !explicit {
				return errors.New(detail)
			}
			fmt.Fprintln(os.Stderr, "Capture state drift (explicit CLI override retained):", detail)
		}
	}
	for _, source := range []string{"codex", "opencode"} {
		manifest, found, err := installedHealthManifest(project, state, source, states)
		if err != nil {
			return err
		}
		if !found {
			continue
		}
		if !manifest.Active {
			continue
		}
		entry, err := adaptersconfig.ResolveEntrypoint(project)
		if err != nil {
			return fmt.Errorf("installed version configuration unavailable; explicit setup refresh required: %w", err)
		}
		native, err := entry.NativePath()
		if err != nil {
			return err
		}
		info, err := os.Stat(native)
		if err != nil || !info.Mode().IsRegular() {
			return errors.New("installed native cache artifact missing; explicitly repair selected release")
		}
		if manifest.Version != entry.Version || !sameEntrypointPath(manifest.Binary, native) || manifest.TemplateVersion != installationTemplateVersion() || manifest.Wrapper == "" || manifest.SourceStore == "" {
			return errors.New("installed capture pin/template/source provenance differs from current release; explicitly refresh owned capture setup before claiming runtime health")
		}
		if err := CheckManagerCompatibility(manifest.Wrapper, manifest.WrapperHash); err != nil {
			return err
		}
		if err := checkOwnedIntegration(project, source, manifest); err != nil {
			return err
		}
	}
	return nil
}

// Anchors support discovery; an exact selected/discovered local manifest still
// proves an installation existed if both its anchor and callback file vanished.
// This reads only the caller's project states, never a global registry census.
func installedHealthManifest(project, state, source string, states []string) (Manifest, bool, error) {
	paths := []string{filepath.Join(project, "."+source, "agenticlens-installation.json"), manifestPath(state, source)}
	for _, installed := range states {
		paths = append(paths, manifestPath(installed, source))
	}
	for _, path := range paths {
		body, err := readEntrypointFile(path, installationHealthFileLimit)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return Manifest{}, false, fmt.Errorf("inspect installed %s manifest %s: %w", source, path, err)
		}
		var manifest Manifest
		if err := json.Unmarshal(body, &manifest); err != nil {
			return Manifest{}, false, fmt.Errorf("decode installed %s manifest %s: %w", source, path, err)
		}
		owner, err := ProjectDirectory(manifest.Repository)
		if err != nil {
			return Manifest{}, false, fmt.Errorf("%w: resolve installed %s project: %w", errInstallationOwnership, source, err)
		}
		if owner != project || manifest.Source != source {
			return Manifest{}, false, fmt.Errorf("%w: integration manifest project/source differs", errInstallationOwnership)
		}
		if manifest.Protocol != ManifestProtocolVersion {
			return Manifest{}, false, errors.New("installed integration manifest protocol differs; refresh owned setup")
		}
		return manifest, true, nil
	}
	return Manifest{}, false, nil
}

// checkOwnedIntegration inspects configuration as data. Recorded exchange and
// human harness trust are independent facts; neither is inferred or changed.
func checkOwnedIntegration(project, source string, manifest Manifest) error {
	path := filepath.Join(project, ".codex", "hooks.json")
	if source == "opencode" {
		path = filepath.Join(project, ".opencode", "plugins", "agenticlens-capture.ts")
	}
	if err := ValidateDestination(project, path); err != nil {
		return fmt.Errorf("installed %s integration path escaped its project: %w", source, err)
	}
	expected, err := ProjectDirectory(path)
	if err != nil {
		return fmt.Errorf("resolve installed %s integration path: %w", source, err)
	}
	configured, err := ProjectDirectory(manifest.ConfigPath)
	if err != nil {
		return fmt.Errorf("resolve installed %s manifest configuration path: %w", source, err)
	}
	if expected != configured {
		return fmt.Errorf("installed %s configuration path differs from owned project integration; refresh owned setup", source)
	}
	body, err := readEntrypointFile(path, installationHealthFileLimit)
	if err != nil {
		return fmt.Errorf("installed %s integration unavailable at %s; restore or refresh owned setup: %w", source, path, err)
	}
	if source == "opencode" {
		if manifest.PluginHash == "" || hash(body) != manifest.PluginHash {
			return fmt.Errorf("installed opencode plugin was edited at %s; preserve local changes and refresh owned setup", path)
		}
		return nil
	}
	var config struct {
		Hooks map[string][]json.RawMessage `json:"hooks"`
	}
	if err := json.Unmarshal(body, &config); err != nil {
		return fmt.Errorf("installed codex hook configuration is invalid at %s: %w", path, err)
	}
	if len(manifest.Groups) == 0 {
		return errors.New("installed codex manifest has no owned hook groups; refresh owned setup")
	}
	for _, owned := range manifest.Groups {
		matches := 0
		for _, group := range config.Hooks[owned.Event] {
			if sameOwnedHookJSON(group, owned.Value) {
				matches++
			}
		}
		if matches != 1 {
			return fmt.Errorf("installed codex owned %s hook was edited, removed or duplicated at %s; preserve unrelated hooks and refresh owned setup", owned.Event, path)
		}
	}
	return nil
}

// Owned JSON object key order and whitespace do not affect execution. Decode
// object members recursively with RawMessage so unknown fields remain evidence
// and are never lost through a narrower handler DTO or an untyped value.
func sameOwnedHookJSON(left, right json.RawMessage) bool {
	return sameOwnedHookJSONDepth(left, right, 0)
}

func sameOwnedHookJSONDepth(left, right json.RawMessage, depth int) bool {
	// Generated groups have three nested containers. Edited manifest metadata
	// must not turn a read-only health comparison into an unbounded recursion.
	if depth > 16 {
		return false
	}
	left, right = bytes.TrimSpace(left), bytes.TrimSpace(right)
	if len(left) == 0 || len(right) == 0 || left[0] != right[0] {
		return false
	}
	switch left[0] {
	case '{':
		var a, b map[string]json.RawMessage
		if json.Unmarshal(left, &a) != nil || json.Unmarshal(right, &b) != nil || len(a) != len(b) {
			return false
		}
		for key, value := range a {
			other, ok := b[key]
			if !ok || !sameOwnedHookJSONDepth(value, other, depth+1) {
				return false
			}
		}
		return true
	case '[':
		var a, b []json.RawMessage
		if json.Unmarshal(left, &a) != nil || json.Unmarshal(right, &b) != nil || len(a) != len(b) {
			return false
		}
		for i := range a {
			if !sameOwnedHookJSONDepth(a[i], b[i], depth+1) {
				return false
			}
		}
		return true
	default:
		return jsonEqual(left, right)
	}
}

// Installed commands contain only statically quoted arguments generated by Install.
// Decode concatenated shell quotes as data, including literal apostrophes.
func installedStateArgument(command string) (string, bool) {
	rest := ""
	for _, flag := range []string{"'--state-dir' ", "\"--state-dir\" ", "--state-dir "} {
		start := strings.Index(command, flag)
		if start >= 0 {
			rest = command[start+len(flag):]
			break
		}
	}
	if rest == "" {
		return "", false
	}

	var result strings.Builder
	quote := byte(0)
	for i := 0; i < len(rest); i++ {
		c := rest[i]
		if quote == 0 {
			if c == ' ' {
				break
			}
			if c == '\'' || c == '"' {
				quote = c
				continue
			}
			return "", false
		}
		if c == quote {
			quote = 0
			continue
		}
		result.WriteByte(c)
	}
	return result.String(), quote == 0 && result.Len() > 0
}
