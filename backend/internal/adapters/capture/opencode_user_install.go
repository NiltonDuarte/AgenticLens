package capture

import (
	"bytes"
	"context"
	_ "embed"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

//go:embed user_plugin.ts
var userPluginTemplate string

const openCodeReceiptName = "agenticlens-opencode-user-installation.json"

type openCodeUserReceipt struct {
	Protocol   int    `json:"protocol"`
	Manager    string `json:"manager"`
	PluginHash string `json:"plugin_hash"`
}

func OpenCodeConfigHome() (string, error) {
	if home := os.Getenv("OPENCODE_CONFIG_DIR"); home != "" {
		if !filepath.IsAbs(home) {
			return "", errors.New("OPENCODE_CONFIG_DIR must be absolute for machine-wide plugin installation")
		}
		return absolute(home)
	}
	return defaultOpenCodeConfigHome()
}

// OpenCode loads its default configuration as well as the selected custom
// directory. Both paths are physical so a symlink alias identifies one load home.
func loadedOpenCodeHomes(selected string) ([]string, error) {
	if selected == "" {
		var err error
		selected, err = OpenCodeConfigHome()
		if err != nil {
			return nil, err
		}
	}
	base, err := defaultOpenCodeConfigHome()
	if err != nil {
		return nil, err
	}
	if base == selected {
		return []string{selected}, nil
	}
	return []string{selected, base}, nil
}

func defaultOpenCodeConfigHome() (string, error) {
	dir := os.Getenv("XDG_CONFIG_HOME")
	if dir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return "", err
		}
		dir = filepath.Join(home, ".config")
	}
	if !filepath.IsAbs(dir) {
		return "", errors.New("XDG_CONFIG_HOME must be absolute")
	}
	return absolute(filepath.Join(dir, "opencode"))
}

func openCodePluginPath(home string) string {
	return filepath.Join(home, "plugins", "agenticlens-capture.ts")
}

func legacyConfigPath(project, source string) string {
	if source == "opencode" {
		return openCodePluginPath(filepath.Join(project, ".opencode"))
	}
	return filepath.Join(project, ".codex", "hooks.json")
}

func InstallUserOpenCode(ctx context.Context, options UserInstallOptions) (InstallResult, error) {
	return installUser(ctx, options, "opencode", nil)
}

// Read owned installation files without adopting a redirected or dangling path.
func openCodeInstallationFile(path string) ([]byte, error) {
	physical, err := absolute(path)
	if err != nil {
		return nil, err
	}
	if physical != path {
		return nil, fmt.Errorf("OpenCode installation path %s resolves elsewhere; preserved", path)
	}
	return readEntrypointFile(path, installationHealthFileLimit)
}

func userPluginReceipt(home string) (openCodeUserReceipt, []byte, error) {
	var receipt openCodeUserReceipt
	body, err := openCodeInstallationFile(filepath.Join(home, openCodeReceiptName))
	if errors.Is(err, os.ErrNotExist) {
		return receipt, nil, nil
	}
	if err != nil {
		return receipt, nil, err
	}
	if err := json.Unmarshal(body, &receipt); err != nil {
		return receipt, nil, err
	}
	if receipt.Protocol != userCaptureOwnershipProtocol || !filepath.IsAbs(receipt.Manager) || len(receipt.PluginHash) != 64 {
		return receipt, nil, errors.New("invalid OpenCode user plugin ownership receipt; preserved")
	}
	return receipt, body, nil
}

// UserOpenCodeInstalled is read-only. A changed plugin cannot reactivate
// legacy worker state merely by leaving an ownership receipt behind.
func UserOpenCodeInstalled() (bool, error) {
	path, err := loadedOpenCodePlugin("")
	return path != "", err
}

// Report the unique active owned plugin. Pending writes, ownership drift and
// competing plugins make activation ambiguous, including in the secondary home.
func loadedOpenCodePlugin(selected string) (string, error) {
	homes, err := loadedOpenCodeHomes(selected)
	if err != nil {
		return "", err
	}
	active := ""
	for _, home := range homes {
		if _, pending, err := readPendingUserRecord(home); err != nil {
			return "", err
		} else if pending {
			return "", fmt.Errorf("pending user installation at %s; recover it before activating OpenCode capture", home)
		}
		receipt, prior, err := userPluginReceipt(home)
		if err != nil {
			return "", err
		}
		owned := ""
		if prior != nil {
			owned = openCodePluginPath(home)
			body, err := openCodeInstallationFile(owned)
			if err != nil {
				return "", err
			}
			if hash(body) != receipt.PluginHash {
				return "", fmt.Errorf("owned OpenCode user plugin was edited at %s; preserved", owned)
			}
			if active != "" {
				return "", fmt.Errorf("duplicate owned OpenCode user plugins at %s and %s; resolve explicitly", active, owned)
			}
			active = owned
		}
		if err := checkOtherOpenCodePlugins(home, owned); err != nil {
			return "", err
		}
	}
	return active, nil
}

func planOpenCodeUser(home, manager string, remove bool) ([]fileChange, Manifest, error) {
	path := openCodePluginPath(home)
	var result Manifest
	receipt, prior, err := userPluginReceipt(home)
	if err != nil {
		return nil, result, err
	}
	if prior != nil && receipt.Manager != manager {
		return nil, result, errors.New("OpenCode manager location differs from owned plugin; explicitly uninstall before relocating")
	}
	if remove && prior == nil {
		return nil, result, errors.New("no owned OpenCode user plugin; refusing removal")
	}
	body, err := openCodeInstallationFile(path)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return nil, result, err
	}
	if exists != (prior != nil) || (exists && hash(body) != receipt.PluginHash) {
		return nil, result, fmt.Errorf("OpenCode user plugin ownership conflict at %s; preserved", path)
	}
	if !remove {
		active, err := loadedOpenCodePlugin(home)
		if err != nil {
			return nil, result, err
		}
		if active != "" && active != path {
			return nil, result, fmt.Errorf("OpenCode user plugin at %s would duplicate installation in %s", active, home)
		}
	}

	encoded, err := json.Marshal(manager)
	if err != nil {
		return nil, result, err
	}
	plugin := []byte(strings.ReplaceAll(userPluginTemplate, "__MANAGER__", string(encoded)))
	next, err := json.MarshalIndent(openCodeUserReceipt{userCaptureOwnershipProtocol, manager, hash(plugin)}, "", "  ")
	if err != nil {
		return nil, result, err
	}
	changes := []fileChange{
		{path: path, before: body, existed: exists, content: plugin, remove: remove},
		{path: filepath.Join(home, openCodeReceiptName), before: prior, existed: prior != nil, content: next, remove: remove},
	}
	result = Manifest{Source: "opencode", ConfigPath: path, Binary: manager, Active: !remove, Capability: "OpenCode user plugin configuration; runtime coverage requires observed evidence"}
	return changes, result, nil
}

// OpenCode discovers both plugin and plugins directories. Restrict inspection
// to AgenticLens's own filename so unrelated user extensions stay untouched.
func checkOtherOpenCodePlugins(home, owned string) error {
	for _, directory := range []string{"plugin", "plugins"} {
		for _, name := range []string{"agenticlens-capture.ts", "agenticlens-capture.js"} {
			path := filepath.Join(home, directory, name)
			if path == owned {
				continue
			}
			_, err := os.Lstat(path)
			if err == nil {
				return fmt.Errorf("existing OpenCode capture plugin at %s would duplicate user capture; preserve it and explicitly migrate an unchanged owned project installation", path)
			}
			if !errors.Is(err, os.ErrNotExist) {
				return err
			}
		}
	}
	return nil
}

// Lease the existing checkout itself during plugin activation so competing
// project and global installs cannot both pass their conflict checks.
func lockOpenCodeProject(project string) (*os.File, error) {
	directory, err := openPublicationDirectory(project)
	if err != nil {
		return nil, err
	}
	held, err := tryPublicationLock(directory)
	if err != nil || !held {
		return nil, errors.Join(fmt.Errorf("OpenCode project installation busy at %s; retry", project), err, directory.Close())
	}
	return directory, nil
}

func CheckOpenCodeProjectConflict(ctx context.Context, project string) error {
	if _, err := loadedOpenCodePlugin(""); err != nil {
		return err
	}
	return checkUserConflicts(ctx, "opencode", project, "")
}

func checkUserConflicts(ctx context.Context, source, project, migration string) error {
	if source == "codex" {
		return checkLegacyUserConflicts(ctx, project, migration)
	}
	projects, err := legacyUserProjects(ctx, project)
	if err != nil {
		return err
	}
	if migration != "" {
		migration, err = ProjectDirectory(migration)
		if err != nil {
			return err
		}
	}
	for _, p := range projects {
		p, err = ProjectDirectory(p)
		if err != nil {
			return err
		}
		home := filepath.Join(p, ".opencode")
		allowed := ""
		if p == migration {
			allowed = openCodePluginPath(home)
		} else if _, exists, err := readPendingUserRecord(home); err != nil {
			return err
		} else if exists {
			return fmt.Errorf("pending OpenCode migration at %s; use --migrate-legacy-repo to recover", home)
		}
		if err := checkOtherOpenCodePlugins(home, allowed); err != nil {
			return err
		}
	}
	return nil
}

func legacyHomesFor(source, project string) ([]string, error) {
	if source == "codex" {
		return legacyUserHomes(project)
	}
	legacy, _, err := readLegacyOpenCode(project)
	if err != nil {
		return nil, err
	}
	return []string{filepath.Join(legacy.Repository, ".opencode"), legacy.StateDir}, nil
}

func readLegacyOpenCode(project string) (Manifest, []byte, error) {
	var legacy Manifest
	project, err := ProjectDirectory(project)
	if err != nil {
		return legacy, nil, err
	}
	anchor := filepath.Join(project, ".opencode", "agenticlens-installation.json")
	body, err := openCodeInstallationFile(anchor)
	if err != nil {
		return legacy, nil, err
	}
	if err := json.Unmarshal(body, &legacy); err != nil {
		return legacy, nil, err
	}
	if legacy.Protocol != ManifestProtocolVersion || legacy.Source != "opencode" || legacy.Repository != project || legacy.ConfigPath != legacyConfigPath(project, "opencode") || !filepath.IsAbs(legacy.StateDir) {
		return legacy, nil, errors.New("legacy OpenCode receipt does not prove project and state ownership")
	}
	return legacy, body, nil
}

func legacyChangesFor(source, project string) ([]fileChange, error) {
	if source == "codex" {
		return legacyUserChanges(project)
	}
	m, prior, err := readLegacyOpenCode(project)
	if err != nil {
		return nil, err
	}
	state := manifestPath(m.StateDir, "opencode")
	beforeState, err := openCodeInstallationFile(state)
	if err != nil {
		return nil, err
	}
	if !bytes.Equal(beforeState, prior) {
		return nil, errors.New("legacy OpenCode project and state receipts differ; preserved")
	}
	before, err := openCodeInstallationFile(m.ConfigPath)
	if errors.Is(err, os.ErrNotExist) && !m.Active {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if m.PluginHash == "" || hash(before) != m.PluginHash {
		return nil, errors.New("legacy OpenCode project plugin was edited; preserved")
	}
	m.Active = false
	after, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return nil, err
	}
	return []fileChange{
		{path: m.ConfigPath, before: before, existed: true, remove: true},
		{path: filepath.Join(m.Repository, ".opencode", "agenticlens-installation.json"), before: prior, existed: true, content: after},
		{path: state, before: beforeState, existed: true, content: after},
	}, nil
}
