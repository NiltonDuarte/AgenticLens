package manager

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
)

const artifactLimit int64 = 256 * 1024 * 1024

// Validate probes a manager's own interface, without project engine selection.
func Validate(path string) (domain.BuildInfo, error) { return config.ValidateManagerExecutable(path) }

func selfUpdate(args []string, output io.Writer) error {
	flags := flag.NewFlagSet("self-update", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	version := flags.String("version", "", "Exact manager version")
	local := flags.String("local-bin", "", "Verified local manager artifact")
	checksum := flags.String("sha256", "", "Expected local artifact SHA256")
	destination := flags.String("dest", "", "Installed manager file")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("usage: self-update --version EXACT [--local-bin PATH --sha256 SHA] [--dest FILE]")
	}
	if err := config.ValidateReleaseVersion(*version); err != nil {
		return err
	}
	path := *destination
	if path == "" {
		var err error
		path, err = os.Executable()
		if err != nil {
			return err
		}
	}
	if strings.ContainsAny(path, "\x00\r\n") {
		return errors.New("path contains control characters")
	}
	path, err := filesystem.PhysicalPath(path)
	if err != nil {
		return err
	}
	staged, clean, err := stageArtifact(*version, "agentic-lens", *local, *checksum)
	if err != nil {
		return err
	}
	defer clean()
	identity, err := Validate(staged)
	if err != nil {
		return fmt.Errorf("replacement manager rejected: %w", err)
	}
	if identity.Version != *version {
		return fmt.Errorf("replacement manager identity %s differs from requested %s", identity.Version, *version)
	}
	if err := replaceManager(path, staged, identity); err != nil {
		return err
	}
	_, err = fmt.Fprintf(output, "Installed manager %s at %s\n", identity.Version, path)
	return err
}

func installEngine(project string, args []string, output io.Writer) (returnErr error) {
	if len(args) == 0 {
		return errors.New("usage: install EXACT [--local-bin PATH --sha256 SHA] [--install-root DIR] [--force]")
	}
	version, args := args[0], args[1:]
	if err := config.ValidateReleaseVersion(version); err != nil {
		return err
	}
	flags := flag.NewFlagSet("install", flag.ContinueOnError)
	flags.SetOutput(io.Discard)
	local := flags.String("local-bin", "", "Verified local host engine artifact")
	checksum := flags.String("sha256", "", "Expected local artifact SHA256")
	root := flags.String("install-root", "", "Explicit version cache root")
	force := flags.Bool("force", false, "Repair an existing version cache leaf")
	platform := flags.String("platform", runtime.GOOS+"-"+runtime.GOARCH, "Engine target OS-ARCH")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("unexpected install arguments")
	}
	entry, err := resolveProject(project, true)
	if err != nil {
		return err
	}
	entry.Version, entry.BinaryDir = version, ""
	if *root != "" {
		entry.InstallRoot, err = validLocalPath(*root)
		if err != nil {
			return err
		}
	}
	osName, arch, err := parsePlatform(*platform)
	if err != nil {
		return err
	}
	if *local != "" && (osName != runtime.GOOS || arch != runtime.GOARCH) {
		return errors.New("foreign-platform local engine installation is unsupported; use a verified published release or validate the artifact on its target platform")
	}
	path := filepath.Join(entry.InstallRoot, "bin", version, *platform, platformArtifactName("agentic-lens-engine", osName, arch))
	staged, clean, err := stagePlatformArtifact(version, "agentic-lens-engine", *local, *checksum, osName, arch)
	if err != nil {
		return err
	}
	defer clean()
	if err := config.ValidateNativePlatform(staged, osName, arch); err != nil {
		return fmt.Errorf("replacement engine rejected: %w", err)
	}
	if osName == runtime.GOOS && arch == runtime.GOARCH {
		identity, err := config.ValidateNativeExecutable(staged)
		if err != nil {
			return fmt.Errorf("replacement engine rejected: %w", err)
		}
		if identity.Version != version {
			return fmt.Errorf("engine identity %s differs from requested %s", identity.Version, version)
		}
	}
	if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
		return err
	}
	lease, err := lockFile(filepath.Join(filepath.Dir(path), ".install.lock"))
	if err != nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, lease.Close()) }()
	before, exists, err := readRegular(path, artifactLimit)
	if err != nil {
		return err
	}
	after, _, err := readRegular(staged, artifactLimit)
	if err != nil {
		return err
	}
	if exists && digest(before) == digest(after) {
		if err := os.Chmod(path, 0700); err != nil {
			return err
		}
	} else {
		if exists && !*force {
			return fmt.Errorf("immutable engine cache leaf already exists at %s; use --force for explicit repair", path)
		}
		if err := atomicWrite(path, after, 0700); err != nil {
			return err
		}
	}
	_, err = fmt.Fprintf(output, "Installed engine %s at %s\n", version, path)
	if err == nil && (osName != runtime.GOOS || arch != runtime.GOARCH) {
		_, err = fmt.Fprintln(output, "Foreign release checksum and platform verified; engine interface and runtime identity validation are deferred until execution on the target platform.")
	}
	return err
}

func digest(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }

func artifactName(kind string) string {
	return platformArtifactName(kind, runtime.GOOS, runtime.GOARCH)
}

func platformArtifactName(kind, osName, arch string) string {
	name := kind + "-" + osName + "-" + arch
	return name
}

func stageArtifact(version, kind, local, expected string) (string, func(), error) {
	return stagePlatformArtifact(version, kind, local, expected, runtime.GOOS, runtime.GOARCH)
}

func parsePlatform(value string) (string, string, error) {
	parts := strings.Split(value, "-")
	if len(parts) != 2 {
		return "", "", fmt.Errorf("unsupported native platform %s", value)
	}
	osName, arch := parts[0], parts[1]
	if (osName == "linux" || osName == "darwin") && (arch == "amd64" || arch == "arm64") {
		return osName, arch, nil
	}
	return "", "", fmt.Errorf("unsupported native platform %s", value)
}

func stagePlatformArtifact(version, kind, local, expected, osName, arch string) (string, func(), error) {
	directory, err := os.MkdirTemp("", "agenticlens-verified-")
	if err != nil {
		return "", nil, err
	}
	clean := func() {
		if err := os.RemoveAll(directory); err != nil {
			fmt.Fprintln(os.Stderr, "AgenticLens temporary artifact cleanup:", err)
		}
	}
	name := platformArtifactName(kind, osName, arch)
	path := filepath.Join(directory, name)
	if local != "" {
		if decoded, err := hex.DecodeString(expected); err != nil || len(decoded) != 32 {
			clean()
			return "", nil, errors.New("--local-bin requires an explicit SHA256 digest")
		}
		file, err := os.Open(local)
		if err != nil {
			clean()
			return "", nil, err
		}
		err = copyVerified(path, file, expected)
		closeErr := file.Close()
		if err = errors.Join(err, closeErr); err != nil {
			clean()
			return "", nil, err
		}
	} else {
		if expected != "" {
			clean()
			return "", nil, errors.New("--sha256 requires --local-bin")
		}
		repository := os.Getenv("AGENTIC_LENS_REPO")
		if repository == "" {
			repository = "NiltonDuarte/AgenticLens"
		}
		parts := strings.Split(repository, "/")
		if len(parts) != 2 || !repositoryPart(parts[0]) || !repositoryPart(parts[1]) {
			clean()
			return "", nil, errors.New("AGENTIC_LENS_REPO must be an owner/repository name")
		}
		base := "https://github.com/" + repository + "/releases/download/" + version
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		client := &http.Client{Timeout: 90 * time.Second}
		if err := downloadVerified(ctx, client, base, name, path); err != nil {
			clean()
			return "", nil, err
		}
	}
	return path, clean, nil
}

func repositoryPart(value string) bool {
	if value == "" || value == "." || value == ".." {
		return false
	}
	for _, r := range value {
		if !((r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= '0' && r <= '9') || r == '-' || r == '_' || r == '.') {
			return false
		}
	}
	return true
}

func downloadVerified(ctx context.Context, client *http.Client, base, name, destination string) error {
	checksums, err := fetch(ctx, client, base+"/SHA256SUMS")
	if err != nil {
		return err
	}
	body, readErr := io.ReadAll(io.LimitReader(checksums.Body, 1048577))
	closeErr := checksums.Body.Close()
	if err := errors.Join(readErr, closeErr); err != nil {
		return err
	}
	if len(body) > 1048576 {
		return errors.New("release checksum file exceeds 1 MiB")
	}
	expected, matches := "", 0
	for _, line := range strings.Split(string(body), "\n") {
		parts := strings.Fields(line)
		if len(parts) != 2 {
			continue
		}
		candidate := strings.TrimPrefix(parts[1], "*")
		if candidate == name || candidate == "bin/"+name {
			expected, matches = parts[0], matches+1
		}
	}
	if decoded, err := hex.DecodeString(expected); matches != 1 || err != nil || len(decoded) != 32 {
		return fmt.Errorf("missing, malformed or ambiguous checksum for %s", name)
	}
	response, err := fetch(ctx, client, base+"/"+name)
	if err != nil {
		return err
	}
	err = copyVerified(destination, response.Body, expected)
	return errors.Join(err, response.Body.Close())
}

func fetch(ctx context.Context, client *http.Client, address string) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, address, nil)
	if err != nil {
		return nil, err
	}
	response, err := client.Do(request)
	if err != nil {
		return nil, err
	}
	if response.StatusCode != http.StatusOK {
		return nil, errors.Join(fmt.Errorf("release download returned HTTP %d for %s", response.StatusCode, address), response.Body.Close())
	}
	return response, nil
}

func copyVerified(destination string, source io.Reader, expected string) error {
	file, err := os.OpenFile(destination, os.O_CREATE|os.O_EXCL|os.O_WRONLY, 0700)
	if err != nil {
		return err
	}
	hasher := sha256.New()
	written, copyErr := io.Copy(io.MultiWriter(file, hasher), io.LimitReader(source, artifactLimit+1))
	syncErr := file.Sync()
	closeErr := file.Close()
	if err := errors.Join(copyErr, syncErr, closeErr); err != nil {
		return err
	}
	if written > artifactLimit {
		return errors.New("native artifact exceeds 256 MiB")
	}
	if !strings.EqualFold(hex.EncodeToString(hasher.Sum(nil)), expected) {
		return errors.New("native artifact SHA256 mismatch; installation preserved")
	}
	return nil
}
