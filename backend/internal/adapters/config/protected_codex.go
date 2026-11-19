package config

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters/filesystem"
	"github.com/pelletier/go-toml/v2"
)

type workspaceWrite struct {
	Roots         []string `toml:"writable_roots"`
	ExcludeTmp    *bool    `toml:"exclude_slash_tmp"`
	ExcludeEnvTmp *bool    `toml:"exclude_tmpdir_env_var"`
}

func insidePath(path, root string) bool {
	rel, err := filepath.Rel(root, path)
	return err == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator))
}

type sandboxSettings struct {
	Write    workspaceWrite `toml:"sandbox_workspace_write"`
	Profiles map[string]struct {
		Write workspaceWrite `toml:"sandbox_workspace_write"`
	} `toml:"profiles"`
}

func readSandboxSettings(path string) (sandboxSettings, error) {
	var cfg sandboxSettings
	data, err := readBootstrap(path)
	if errors.Is(err, os.ErrNotExist) {
		return cfg, nil
	}
	if err != nil {
		return cfg, err
	}
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return cfg, fmt.Errorf("read Codex sandbox grants from %s: %w", path, err)
	}
	return cfg, nil
}

func (profile workspaceWrite) grants() []string {
	roots := profile.Roots
	if profile.ExcludeTmp != nil && !*profile.ExcludeTmp {
		roots = append(roots, "/tmp")
	}
	if profile.ExcludeEnvTmp != nil && !*profile.ExcludeEnvTmp && os.Getenv("TMPDIR") != "" {
		roots = append(roots, os.Getenv("TMPDIR"))
	}
	return roots
}

func (cfg sandboxSettings) grants() []string {
	roots := cfg.Write.grants()
	for _, profile := range cfg.Profiles {
		roots = append(roots, profile.Write.grants()...)
	}
	return roots
}

// CheckProtectedCodexPaths rejects paths under the workspace, its configured
// workspace-write grants and normal temporary directories. Additional launcher
// grants are operator policy because hook payloads do not enumerate them.
func CheckProtectedCodexPaths(project, home string, paths ...string) error {
	if !filepath.IsAbs(home) {
		return errors.New("protected Codex capture requires an absolute runtime home")
	}
	roots := []string{filepath.Join(home, CodexInboxName)}
	path := filepath.Join(home, "config.toml")
	cfg, err := readSandboxSettings(path)
	if err != nil {
		return err
	}
	roots = append(roots, cfg.grants()...)
	if cfg.Write.ExcludeTmp == nil || !*cfg.Write.ExcludeTmp {
		roots = append(roots, "/tmp")
	}
	if (cfg.Write.ExcludeEnvTmp == nil || !*cfg.Write.ExcludeEnvTmp) && os.Getenv("TMPDIR") != "" {
		roots = append(roots, os.Getenv("TMPDIR"))
	}
	if project != "" {
		physical, err := filesystem.PhysicalPath(project)
		if err != nil {
			return err
		}
		roots = append(roots, physical)
		// A nested configured project may share a surrounding Git worktree.
		// Every project-level grant adds to the protected-path exclusions.
		for dir := physical; ; dir = filepath.Dir(dir) {
			if dir != physical {
				if _, err := os.Lstat(filepath.Join(dir, ".git")); err == nil {
					roots = append(roots, dir)
				}
			}
			projectSettings, err := readSandboxSettings(filepath.Join(dir, ".codex", "config.toml"))
			if err != nil {
				return err
			}
			roots = append(roots, projectSettings.grants()...)
			if filepath.Dir(dir) == dir {
				break
			}
		}
	}
	// Treat the machine settings and capture receipt as authority only outside
	// the writable roots described by those settings.
	paths = append(paths, path, filepath.Join(home, "agenticlens-user-installation.json"))
	for _, root := range roots {
		if !filepath.IsAbs(root) {
			return fmt.Errorf("cannot establish protected Codex capture with relative writable root %q", root)
		}
		root, err := filesystem.PhysicalPath(root)
		if err != nil {
			return err
		}
		for _, path := range paths {
			if !filepath.IsAbs(path) {
				return errors.New("protected capture requires absolute paths")
			}
			physical, err := filesystem.PhysicalPath(path)
			if err != nil {
				return err
			}
			if insidePath(physical, root) || insidePath(root, physical) || insidePath(filepath.Clean(path), root) {
				return fmt.Errorf("protected Codex capture path %s overlaps writable root %s", path, root)
			}
		}
	}
	return nil
}

type codexCaptureReceipt struct {
	Protocol int               `json:"protocol"`
	Manager  string            `json:"manager"`
	Inbox    bool              `json:"binding_inbox"`
	Groups   []json.RawMessage `json:"groups"`
}

const CodexUserInstallPending = "agenticlens-user-install.pending.json"

func readCodexCaptureReceipt(home string) (codexCaptureReceipt, error) {
	var r codexCaptureReceipt
	if _, err := os.Lstat(filepath.Join(home, CodexUserInstallPending)); !errors.Is(err, os.ErrNotExist) {
		if err != nil {
			return r, err
		}
		return r, errors.New("Codex user installation transaction is pending; finish installer recovery")
	}
	data, err := readBootstrap(filepath.Join(home, "agenticlens-user-installation.json"))
	if err != nil {
		return r, err
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return r, err
	}
	if r.Protocol != 2 || !filepath.IsAbs(r.Manager) {
		return r, errors.New("invalid Codex user installation receipt; explicitly repair setup")
	}
	return r, nil
}

// ProtectedReconcileHome selects the existing admission or the operator's
// active Codex integration for an enabled project using that same source.
func ProtectedReconcileHome(project string) (string, error) {
	home, err := ProtectedCodexProjectHome(project)
	if err != nil || home != "" {
		return home, err
	}
	policy, err := ResolveCapture(project, "")
	if err != nil || !policy.Enabled {
		return "", err
	}
	cfg, err := Load(project)
	if err != nil || cfg.Telemetry.Sources.CodexHome == "" {
		return "", err
	}
	configured, err := filesystem.PhysicalPath(cfg.Telemetry.Sources.CodexHome)
	if err != nil {
		return "", err
	}
	actual, err := filesystem.PhysicalPath(RuntimeCodexHome())
	if err != nil || actual != configured {
		return "", err
	}
	receipt, err := readCodexCaptureReceipt(actual)
	if errors.Is(err, os.ErrNotExist) {
		return "", nil
	}
	if err != nil || !receipt.Inbox {
		return "", err
	}
	return actual, nil
}

// ProtectedWorkerProtocol requires admitted worktree routing and guarded private
// state throughout the generation child, including lock and journal acquisition.
const ProtectedWorkerProtocol = "2"

const protectedEngineLimit = 256 << 20

// Verify the operator-owned regular file, without following an alias or running
// the file. The digest originates in the explicit manager installation.
func verifiedCodexFile(path, expected string) error {
	if _, err := hex.DecodeString(expected); err != nil || len(expected) != sha256.Size*2 {
		return errors.New("invalid protected executable checksum")
	}
	st, err := os.Lstat(path)
	if err != nil {
		return err
	}
	if !st.Mode().IsRegular() || st.Size() > protectedEngineLimit {
		return fmt.Errorf("invalid protected native executable %s", path)
	}
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	defer f.Close()
	opened, err := f.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(st, opened) {
		return errors.New("protected executable changed while opening")
	}
	h := sha256.New()
	n, err := io.Copy(h, io.LimitReader(f, protectedEngineLimit+1))
	if err != nil {
		return err
	}
	if n != st.Size() || hex.EncodeToString(h.Sum(nil)) != expected {
		return fmt.Errorf("protected executable checksum differs for %s", path)
	}
	return nil
}

func verifyProtectedNames(project, home string, paths ...string) error {
	if err := CheckProtectedCodexPaths(project, home, paths...); err != nil {
		return err
	}
	for _, p := range paths {
		physical, err := filesystem.PhysicalPath(p)
		if err != nil {
			return err
		}
		if info, err := os.Lstat(p); err == nil {
			st, ok := info.Sys().(*syscall.Stat_t)
			if !ok || !info.Mode().IsRegular() || info.Mode().Perm()&0022 != 0 || st.Nlink != 1 || (st.Uid != 0 && int(st.Uid) != os.Geteuid()) {
				return fmt.Errorf("protected installation file must be owned, regular, single-linked and private from other writers: %s", p)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if physical != p {
			return fmt.Errorf("protected installation path contains an alias: %s", p)
		}
	}
	return nil
}

// ProtectedCodexManager checks self-update's existing ownership and checksum
// files before any protocol probe. Repair remains an explicit operator action.
func ProtectedCodexManager(project, home, manager string) (string, error) {
	side := manager + ".sha256"
	record := manager + ".installation.json"
	if err := verifyProtectedNames(project, home, manager, side, record, manager+".pending.json"); err != nil {
		return "", err
	}
	if _, err := os.Lstat(manager + ".pending.json"); !errors.Is(err, os.ErrNotExist) {
		return "", errors.New("manager installation pending or unavailable; resolve its self-update before enabling capture")
	}
	var r struct {
		Protocol              int
		Path, SHA256, Version string
	}
	data, err := readBootstrap(record)
	if err != nil {
		return "", fmt.Errorf("protected Codex capture requires an owned manager; provision it with self-update: %w", err)
	}
	if err := json.Unmarshal(data, &r); err != nil {
		return "", err
	}
	if r.Protocol != 1 || r.Path != manager || ValidateReleaseVersion(r.Version) != nil {
		return "", errors.New("invalid protected manager ownership record; explicitly repair its installation")
	}
	data, err = readBootstrap(side)
	if err != nil {
		return "", err
	}
	if string(data) != r.SHA256+"\n" {
		return "", errors.New("protected manager checksum conflicts with ownership record")
	}
	if err := verifiedCodexFile(manager, r.SHA256); err != nil {
		return "", err
	}
	return r.Version, nil
}

// ResolveProtectedCodexEngine uses only the shared exact release intent and
// operator-owned installation. A development checkout uses the manager companion.
func ResolveProtectedCodexEngine(project, home string, installManager ...string) (string, error) {
	manager := ""
	if len(installManager) > 0 {
		manager = installManager[0]
	} else {
		r, err := readCodexCaptureReceipt(home)
		if err != nil {
			return "", fmt.Errorf("protected Codex installation receipt unavailable; run capture install-user: %w", err)
		}
		if !r.Inbox || r.Manager == "" {
			return "", errors.New("protected Codex installation requires an active inbox receipt")
		}
		manager = r.Manager
	}
	version, err := ProtectedCodexManager(project, home, manager)
	if err != nil {
		return "", err
	}
	version, err = ProtectedCodexVersion(project, version)
	if err != nil {
		return "", err
	}
	user, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	user, err = filesystem.PhysicalPath(user)
	if err != nil {
		return "", err
	}
	root := filepath.Join(user, ".agenticlens")
	entry := Entrypoint{InstallRoot: root, Version: version}
	path, err := entry.CachedNativePath()
	if err != nil {
		return "", err
	}
	if err := verifyProtectedNames(project, home, path, path+".sha256"); err != nil {
		return "", err
	}
	checksum, err := readBootstrap(path + ".sha256")
	if err == nil {
		err = verifiedCodexFile(path, strings.TrimSuffix(string(checksum), "\n"))
	}
	if err != nil {
		return "", fmt.Errorf("protected engine %s unavailable or modified; explicitly install %s --install-root %s using a verified artifact: %w", path, version, root, err)
	}
	identity, err := ValidateNativeExecutable(path)
	if err != nil {
		return "", err
	}
	if identity.Version != version {
		return "", fmt.Errorf("protected engine requires release %s, found %s", version, identity.Version)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	body, err := exec.CommandContext(ctx, path, "capture", "protected-worker-protocol").Output()
	if err != nil || strings.TrimSpace(string(body)) != ProtectedWorkerProtocol {
		return "", fmt.Errorf("protected engine %s requires protected worker protocol %s for guarded publication; provision a compatible release", path, ProtectedWorkerProtocol)
	}
	return path, nil
}

// ProtectedCodexVersion allows a shared release requirement or, for development,
// the release of the independently verified manager companion.
func ProtectedCodexVersion(project, companion string) (string, error) {
	if project == "" {
		return companion, nil
	}
	shared, err := FindConfigFile(project)
	if err != nil || shared == "" {
		return companion, err
	}
	data, err := readBootstrap(shared)
	if err != nil {
		return "", err
	}
	pin, err := parseEntrypoint(data, false)
	if err != nil {
		return "", err
	}
	if pin.BinaryDir != "" {
		return companion, nil
	}
	return pin.Version, nil
}
