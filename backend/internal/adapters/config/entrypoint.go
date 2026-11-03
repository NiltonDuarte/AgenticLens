package config

import (
	"bufio"
	"bytes"
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
)

const bootstrapLimit = 1024 * 1024

var releasePattern = regexp.MustCompile(`^v(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)\.(0|[1-9][0-9]*)(-[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?(\+[0-9A-Za-z]+([.-][0-9A-Za-z]+)*)?$`)
var bootstrapKey = regexp.MustCompile(`^([A-Za-z_][A-Za-z0-9_-]*):(?:[ \t]+(.*))?$`)
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
	InstallRoot string
}

// NativePath selects the executing environment, never the installing host.
func (e Entrypoint) NativePath() (string, error) {
	osName, arch := runtime.GOOS, runtime.GOARCH
	if (osName != "linux" && osName != "darwin" && osName != "windows") || (arch != "amd64" && arch != "arm64") || (osName == "windows" && arch != "amd64") {
		return "", fmt.Errorf("unsupported native platform %s-%s", osName, arch)
	}
	name := "agentic-lens-" + osName + "-" + arch
	if osName == "windows" {
		name += ".exe"
	}
	return filepath.Join(e.InstallRoot, "bin", e.Version, osName+"-"+arch, name), nil
}

// UserWrapperPath is stable across projects and release switches.
func UserWrapperPath() (string, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return "", fmt.Errorf("user wrapper home: %w", err)
	}
	if runtime.GOOS == "windows" {
		return filepath.Join(home, ".agenticlens", "agentic-lens.ps1"), nil
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
// JSON remains supported by the legacy native loader, but cannot bootstrap YAML authority.
func ResolveEntrypoint(start string) (Entrypoint, error) {
	var result Entrypoint
	path, err := FindConfigFile(start)
	if err != nil {
		return result, err
	}
	if path == "" {
		return result, errors.New("shared project YAML missing; explicit setup must write agenticlens.version")
	}
	if filepath.Ext(path) == ".json" {
		return result, errors.New("JSON version bootstrap unsupported; migrate shared configuration to .agenticlens-config.yaml")
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
	localPath := filepath.Join(root, ".agenticlens-config.local.yaml")
	localData, err := readBootstrap(localPath)
	if err == nil {
		local, parseErr := parseEntrypoint(localData, true)
		if parseErr != nil {
			return result, fmt.Errorf("local version configuration: %w", parseErr)
		}
		if local.InstallRoot != "" {
			shared.InstallRoot = local.InstallRoot
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	if value := os.Getenv("AGENTIC_LENS_VERSION"); value != "" && value != shared.Version {
		return result, errors.New("AGENTIC_LENS_VERSION conflicts with authoritative shared agenticlens.version")
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return result, err
	}
	installRoot := shared.InstallRoot
	if installRoot == "" {
		installRoot = filepath.Join(home, ".agenticlens")
	} else if strings.HasPrefix(installRoot, "~/") {
		installRoot = home + string(filepath.Separator) + installRoot[2:]
	}
	if !filepath.IsAbs(installRoot) {
		return result, errors.New("agenticlens.install_root is not absolute on the executing platform")
	}
	installRoot, err = filesystem.PhysicalPath(installRoot)
	if err != nil {
		return result, fmt.Errorf("physical install_root: %w", err)
	}
	return Entrypoint{Project: root, ConfigPath: path, Version: shared.Version, InstallRoot: installRoot}, nil
}

// parseEntrypoint matches the documented shell/PowerShell bootstrap profile.
// The full native YAML loader still validates unrelated settings after selection.
func parseEntrypoint(data []byte, local bool) (domain.EntrypointConfig, error) {
	var result domain.EntrypointConfig
	if len(data) > bootstrapLimit {
		return result, errors.New("bootstrap configuration exceeds 1 MiB")
	}
	data = bytes.TrimPrefix(data, []byte{0xef, 0xbb, 0xbf})
	roots, children := map[string]bool{}, map[string]bool{}
	scanner := bufio.NewScanner(bytes.NewReader(data))
	scanner.Buffer(make([]byte, 4096), bootstrapLimit+1)
	active := false
	number := 0
	failure := func(message string) (domain.EntrypointConfig, error) {
		return result, fmt.Errorf("line %d: %s; use the documented literal bootstrap YAML form", number, message)
	}
	for scanner.Scan() {
		number++
		raw := strings.TrimSuffix(scanner.Text(), "\r")
		if strings.HasPrefix(strings.TrimLeft(raw, " "), "\t") {
			return failure("tab indentation unsupported")
		}
		line, err := bootstrapLex(raw)
		if err != nil {
			return failure(err.Error())
		}
		if strings.Trim(line, " \t") == "" {
			continue
		}
		if strings.HasPrefix(line, "---") || strings.HasPrefix(line, "...") || strings.HasPrefix(line, "%") {
			return failure("YAML document markers/directives unsupported")
		}
		if line[0] != ' ' {
			parts := bootstrapKey.FindStringSubmatch(line)
			if parts == nil {
				return failure("use plain top-level mapping keys")
			}
			key := parts[1]
			if roots[key] {
				return failure("duplicate top-level mapping key")
			}
			roots[key] = true
			active = key == "agenticlens"
			if active && strings.Trim(parts[2], " \t") != "" {
				return failure("agenticlens must be a block mapping")
			}
			continue
		}
		if strings.HasPrefix(strings.TrimLeft(line, " "), "<<:") {
			return failure("YAML merges unsupported")
		}
		if !active {
			continue
		}
		if !strings.HasPrefix(line, "  ") {
			return failure("agenticlens children require two spaces")
		}
		parts := bootstrapKey.FindStringSubmatch(line[2:])
		if parts == nil {
			return failure("agenticlens children require exactly two spaces and plain keys")
		}
		key := parts[1]
		if children[key] {
			return failure("duplicate agenticlens key")
		}
		children[key] = true
		value := strings.Trim(parts[2], " \t")
		if value == "" {
			return failure("empty bootstrap scalar")
		}
		if value[0] == '\'' || value[0] == '"' {
			quote := value[0]
			if len(value) < 2 || value[len(value)-1] != quote {
				return failure("unsupported quoted scalar")
			}
			value = value[1 : len(value)-1]
			if strings.ContainsRune(value, rune(quote)) {
				return failure("embedded quotes unsupported")
			}
		} else if strings.ContainsAny(value, " \t#\"'") {
			return failure("quote values containing spaces or comments")
		}
		if value == "" || strings.ContainsAny(value, "\x00\r\n") {
			return failure("empty or unsafe bootstrap scalar")
		}
		switch key {
		case "version":
			if local {
				return failure("local version overrides prohibited; edit shared YAML")
			}
			if err := ValidateReleaseVersion(value); err != nil {
				return failure(err.Error())
			}
			result.Version = value
		case "install_root":
			if len(value) > 4096 || !absoluteInstallRoot.MatchString(value) || strings.ContainsAny(value, "<>|?*") {
				return failure("install_root must be an absolute path or ~/path")
			}
			result.InstallRoot = value
		default:
			return failure("unknown agenticlens key")
		}
	}
	if err := scanner.Err(); err != nil {
		return result, fmt.Errorf("read bootstrap YAML: %w", err)
	}
	if !local && result.Version == "" {
		return result, errors.New("missing agenticlens.version; explicit setup must write an exact release pin")
	}
	return result, nil
}

func bootstrapLex(line string) (string, error) {
	var out strings.Builder
	var quote rune
	var previous rune
	for _, c := range line {
		if quote != 0 {
			if c == '\\' && quote == '"' {
				return "", errors.New("escaped double-quoted scalars unsupported")
			}
			if c == quote {
				quote = 0
			}
			out.WriteRune(c)
		} else {
			if c == '#' && (out.Len() == 0 || previous == ' ' || previous == '\t') {
				break
			}
			if (c == '"' || c == '\'') && (out.Len() == 0 || strings.ContainsRune(" \t:[,", previous)) {
				quote = c
			}
			if strings.ContainsRune("&*!|>", c) {
				return "", errors.New("anchors, aliases, tags and block scalars unsupported")
			}
			out.WriteRune(c)
		}
		previous = c
	}
	if quote != 0 {
		return "", errors.New("unclosed quoted scalar")
	}
	return out.String(), nil
}
