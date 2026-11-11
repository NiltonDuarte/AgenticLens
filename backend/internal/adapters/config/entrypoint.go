package config

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
	"gopkg.in/yaml.v3"
)

const bootstrapLimit = 1024 * 1024

var releasePattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$`)
var absoluteInstallRoot = regexp.MustCompile(`^(/|~/|[A-Za-z]:[\\/]|\\\\[^\\]+\\[^\\]+)`)

// ValidateReleaseVersion accepts only an exact, path-safe supported release tag.
func ValidateReleaseVersion(version string) error {
	if len(version) > 128 || !releasePattern.MatchString(version) {
		return errors.New("agenticlens.version must be an exact vMAJOR.MINOR.PATCH[-prerelease][+build] release, never latest")
	}
	return nil
}

// Entrypoint is resolved once from shared release authority and local path policy.
type Entrypoint struct {
	Project     string
	ConfigPath  string
	Version     string
	BinaryDir   string
	InstallRoot string
}

// NativePath selects the executing environment, never the installing host.
func (e Entrypoint) NativePath() (string, error) {
	if e.BinaryDir != "" {
		name := "agentic-lens-engine-" + runtime.GOOS + "-" + runtime.GOARCH
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		development := filepath.Join(e.BinaryDir, name)
		if _, err := os.Lstat(development); err == nil {
			if _, err := ValidateNativeExecutable(development); err != nil {
				return "", fmt.Errorf("development output %s: %w", development, err)
			}
			return development, nil
		} else {
			return "", fmt.Errorf("development output %s: %w", development, err)
		}
	}
	return e.CachedNativePath()
}

// CachedNativePath identifies only the exact shared pin; provisioning is explicit.
func (e Entrypoint) CachedNativePath() (string, error) {
	osName, arch := runtime.GOOS, runtime.GOARCH
	if (osName != "linux" && osName != "darwin" && osName != "windows") || (arch != "amd64" && arch != "arm64") || (osName == "windows" && arch != "amd64") {
		return "", fmt.Errorf("unsupported native platform %s-%s", osName, arch)
	}
	name := "agentic-lens-engine-" + osName + "-" + arch
	if osName == "windows" {
		name += ".exe"
	}
	return filepath.Join(e.InstallRoot, "bin", e.Version, osName+"-"+arch, name), nil
}

// UserManagerPath is the native manager command, stable across engine switches.
func UserManagerPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user wrapper home: %w", err)
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".agenticlens", "agentic-lens.exe"), nil
	}
	return filepath.Join(home, ".local", "bin", "agentic-lens"), nil
}

func readBootstrap(path string) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		if _, linkErr := os.Lstat(path); linkErr == nil {
			return nil, fmt.Errorf("bootstrap configuration exists but is unavailable: %s", path)
		}
		return nil, err
	}
	if !info.Mode().IsRegular() {
		return nil, errors.New("bootstrap configuration must be a regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	data, err := io.ReadAll(io.LimitReader(file, bootstrapLimit+1))
	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, err
	}
	if len(data) > bootstrapLimit {
		return nil, errors.New("bootstrap configuration exceeds 1 MiB")
	}
	return data, nil
}

// ResolveEntrypoint preserves FindConfigFile ownership/suffix/Git boundaries.
// Shared bootstrap and engine configuration use the same YAML discovery contract.
func ResolveEntrypoint(start string) (Entrypoint, error) {
	var result Entrypoint
	path, err := FindConfigFile(start)
	if err != nil {
		return result, err
	}
	if path == "" {
		return result, errors.New("shared project YAML missing; explicit setup must write agenticlens.version")
	}
	data, err := readBootstrap(path)
	if err != nil {
		return result, fmt.Errorf("shared version configuration: %w", err)
	}
	shared, err := parseEntrypoint(data, false)
	if err != nil {
		return result, fmt.Errorf("shared version configuration %s: %w", path, err)
	}
	root, err := filepath.EvalSymlinks(filepath.Dir(path))
	if err != nil {
		return result, err
	}
	selection, selectErr := SelectConfig(root)
	if selectErr == nil {
		localData, err := readSelectedLocal(selection)
		if err != nil {
			return result, err
		}
		local, err := parseEntrypoint(localData, true)
		if err != nil {
			return result, fmt.Errorf("local version configuration %s: %w", selection.LocalPath, err)
		}
		if local.InstallRoot != "" {
			shared.InstallRoot = local.InstallRoot
		}
	} else if !errors.Is(selectErr, os.ErrNotExist) {
		return result, selectErr
	}

	if value := os.Getenv("AGENTIC_LENS_VERSION"); value != "" && value != shared.Version {
		return result, errors.New("AGENTIC_LENS_VERSION conflicts with authoritative shared agenticlens.version")
	}
	installRoot := shared.InstallRoot
	if installRoot == "" || strings.HasPrefix(installRoot, "~/") {
		home, homeErr := os.UserHomeDir()
		if homeErr != nil || strings.TrimSpace(home) == "" {
			return result, errors.New("selected installation root requires an available user home")
		}
		if installRoot == "" {
			installRoot = filepath.Join(home, ".agenticlens")
		} else {
			installRoot = home + string(filepath.Separator) + installRoot[2:]
		}
	}
	if !filepath.IsAbs(installRoot) {
		return result, errors.New("agenticlens.install_root is not absolute on the executing platform")
	}
	installRoot, err = filesystem.PhysicalPath(installRoot)
	if err != nil {
		return result, fmt.Errorf("physical install_root: %w", err)
	}
	binaryDir := shared.BinaryDir
	if binaryDir != "" {
		binaryDir, err = ResolveConfiguredPath(root, binaryDir)
		if err != nil {
			return result, fmt.Errorf("agenticlens.binary_dir: %w", err)
		}
		binaryDir, err = filesystem.PhysicalPath(binaryDir)
		if err != nil {
			return result, fmt.Errorf("agenticlens.binary_dir: %w", err)
		}
	}
	return Entrypoint{Project: root, ConfigPath: path, Version: shared.Version, InstallRoot: installRoot, BinaryDir: binaryDir}, nil
}

// parseEntrypoint reads release authority from the same YAML document contract
// as the engine. Unrelated scalar styles/comments do not change bootstrap identity.
func parseEntrypoint(data []byte, local bool) (domain.EntrypointConfig, error) {
	var result domain.EntrypointConfig
	if len(data) > bootstrapLimit {
		return result, errors.New("bootstrap configuration exceeds 1 MiB")
	}
	nodes, err := configurationNodes(data)
	if err != nil {
		return result, err
	}
	entry := nodes["agenticlens"]
	if entry == nil {
		if local {
			return result, nil
		}
		return result, errors.New("missing agenticlens.version; setup must write an exact approved release pin")
	}
	for entry.Kind == yaml.AliasNode && entry.Alias != nil {
		entry = entry.Alias
	}
	if entry.Kind != yaml.MappingNode {
		return result, errors.New("agenticlens must be a YAML mapping")
	}
	for i := 0; i < len(entry.Content); i += 2 {
		key, value := entry.Content[i].Value, entry.Content[i+1]
		for value.Kind == yaml.AliasNode && value.Alias != nil {
			value = value.Alias
		}
		if value.Kind != yaml.ScalarNode || value.Tag != "!!str" || strings.TrimSpace(value.Value) == "" || strings.ContainsAny(value.Value, "\x00\r\n") {
			return result, fmt.Errorf("agenticlens.%s must be an explicit nonblank single-line string", key)
		}
		switch key {
		case "version":
			if local {
				return result, errors.New("agenticlens.version belongs only in shared YAML")
			}
			if err := ValidateReleaseVersion(value.Value); err != nil {
				return result, err
			}
			result.Version = value.Value
		case "binary_dir":
			if local {
				return result, errors.New("agenticlens.binary_dir belongs only in shared YAML")
			}
			if len(value.Value) > 4096 {
				return result, errors.New("agenticlens.binary_dir exceeds path limit")
			}
			result.BinaryDir = value.Value
		case "install_root":
			if len(value.Value) > 4096 || !absoluteInstallRoot.MatchString(value.Value) {
				return result, errors.New("agenticlens.install_root must be absolute or explicitly ~/path")
			}
			result.InstallRoot = value.Value
		default:
			return result, fmt.Errorf("unsupported agenticlens setting %q", key)
		}
	}
	if !local && result.Version == "" {
		return result, errors.New("missing agenticlens.version; setup must write an exact approved release pin")
	}
	return result, nil
}

// ReadReleasePin reads only explicit shared release authority for setup completion.
func ReadReleasePin(path string) (string, error) {
	data, err := readBootstrap(path)
	if err != nil {
		return "", err
	}
	entry, err := parseEntrypoint(data, false)
	if err != nil {
		return "", err
	}
	return entry.Version, nil
}
