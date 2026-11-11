package capture

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

// legacyUserChanges removes only exact groups attested by the existing owned
// project receipt. All state and modified/unknown callbacks remain preserved.
func legacyUserChanges(project string) ([]fileChange, error) {
	project, err := ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	anchor := filepath.Join(project, ".codex", "agenticlens-installation.json")
	body, err := os.ReadFile(anchor)
	if err != nil {
		return nil, err
	}
	var legacy Manifest
	if err := json.Unmarshal(body, &legacy); err != nil {
		return nil, err
	}
	owner, err := ProjectDirectory(legacy.Repository)
	if err != nil {
		return nil, err
	}
	if owner != project || legacy.Source != "codex" || legacy.Protocol != ManifestProtocolVersion || legacy.ConfigPath != filepath.Join(project, ".codex", "hooks.json") {
		return nil, errors.New("legacy Codex receipt does not prove this project hook ownership")
	}
	content, err := os.ReadFile(legacy.ConfigPath)
	if errors.Is(err, os.ErrNotExist) && !legacy.Active {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	config := map[string]json.RawMessage{}
	if err := json.Unmarshal(content, &config); err != nil {
		return nil, err
	}
	hooks := map[string][]json.RawMessage{}
	if err := json.Unmarshal(config["hooks"], &hooks); err != nil {
		return nil, err
	}
	removed := 0
	for _, owned := range legacy.Groups {
		kept := []json.RawMessage{}
		matches := 0
		for _, group := range hooks[owned.Event] {
			if jsonEqual(group, owned.Value) {
				matches++
			} else {
				kept = append(kept, group)
			}
		}
		if matches == 0 && !legacy.Active {
			continue
		}
		if matches != 1 {
			return nil, fmt.Errorf("legacy %s hook edited or duplicated; preserved", owned.Event)
		}
		removed++
		if len(kept) == 0 {
			delete(hooks, owned.Event)
		} else {
			hooks[owned.Event] = kept
		}
	}
	if legacyCaptureGroupsRemain(hooks) {
		return nil, fmt.Errorf("unknown or modified AgenticLens legacy callbacks remain at %s after planned removal of unchanged owned groups; preserved; resolve these callbacks explicitly before user-hook activation", legacy.ConfigPath)
	}
	if !legacy.Active && removed == 0 {
		return nil, nil
	}
	raw, err := json.Marshal(hooks)
	if err != nil {
		return nil, err
	}
	config["hooks"] = raw
	after, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, err
	}
	legacy.Active = false
	receipt, err := json.MarshalIndent(legacy, "", "  ")
	if err != nil {
		return nil, err
	}
	statePath := manifestPath(legacy.StateDir, "codex")
	priorState, err := os.ReadFile(statePath)
	if err != nil {
		return nil, err
	}
	var state Manifest
	if err := json.Unmarshal(priorState, &state); err != nil {
		return nil, err
	}
	if state.Repository != legacy.Repository || state.ConfigPath != legacy.ConfigPath || state.Source != "codex" || state.Protocol != ManifestProtocolVersion {
		return nil, errors.New("legacy state manifest ownership differs; preserved")
	}
	return []fileChange{{path: legacy.ConfigPath, content: after, before: content, existed: true}, {path: anchor, content: receipt, before: body, existed: true}, {path: statePath, content: receipt, before: priorState, existed: true}}, nil
}

// legacyUserHomes locates resource leases without requiring the hook bytes to
// match a completed install. Interrupted migration recovery restores those bytes
// under the leases before legacyUserChanges builds a fresh ownership plan.
func legacyUserHomes(project string) ([]string, error) {
	project, err := ProjectDirectory(project)
	if err != nil {
		return nil, err
	}
	anchor := filepath.Join(project, ".codex", "agenticlens-installation.json")
	data, err := os.ReadFile(anchor)
	if err != nil {
		return nil, err
	}
	var manifest Manifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	owner, err := ProjectDirectory(manifest.Repository)
	if err != nil {
		return nil, err
	}
	if owner != project || manifest.Source != "codex" || manifest.Protocol != ManifestProtocolVersion || manifest.ConfigPath != filepath.Join(project, ".codex", "hooks.json") || manifest.StateDir == "" {
		return nil, errors.New("legacy resource receipt does not prove project/state ownership")
	}
	return []string{filepath.Dir(manifest.ConfigPath), manifest.StateDir}, nil
}
