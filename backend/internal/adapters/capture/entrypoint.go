package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
)

func installationTemplateVersion() string {
	return hash([]byte(marker + pluginTemplate + "common-wrapper-v1;codex-windows-outer5-native2"))
}

// CheckWrapperCompatibility verifies current shared-wrapper integrity separately
// from a project's historical installation identity. A compatible installer-owned
// global update must not invalidate another project's unchanged manifest.
func CheckWrapperCompatibility(wrapper, historicalHash string) error {
	body, err := readEntrypointFile(wrapper, 128*1024)
	if err != nil {
		return fmt.Errorf("common wrapper unavailable; explicit setup repair required: %w", err)
	}
	compatible := false
	for _, line := range strings.Split(string(body), "\n") {
		if strings.HasPrefix(line, "# AgenticLens common entrypoint ") {
			if !strings.HasPrefix(line, "# AgenticLens common entrypoint v1:") {
				return errors.New("common wrapper contract incompatible; explicitly refresh setup")
			}
			compatible = true
		}
	}
	if !compatible {
		return errors.New("common wrapper contract unavailable; explicitly refresh setup")
	}
	current := hash(body)
	if current == historicalHash {
		return nil
	}
	owned, err := readEntrypointFile(wrapper+".sha256", 256)
	if err != nil || strings.TrimSpace(string(owned)) != current {
		return errors.New("shared common wrapper changed without matching installer ownership; edited/unowned wrapper preserved, explicit repair required")
	}
	return nil
}

func readEntrypointFile(path string, limit int64) ([]byte, error) {
	info, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !info.Mode().IsRegular() || info.Size() > limit {
		return nil, errors.New("entrypoint metadata must be a bounded regular file")
	}
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	body, err := io.ReadAll(io.LimitReader(file, limit+1))
	err = errors.Join(err, file.Close())
	if err != nil {
		return nil, err
	}
	if int64(len(body)) > limit {
		return nil, errors.New("entrypoint metadata exceeds size limit")
	}
	return body, nil
}

func validateNativeHeader(path string) error {
	file, err := os.Open(path)
	if err != nil {
		return err
	}
	var prefix [4]byte
	_, err = io.ReadFull(file, prefix[:])
	err = errors.Join(err, file.Close())
	if err != nil {
		return err
	}
	valid := false
	switch runtime.GOOS {
	case "linux":
		valid = bytes.Equal(prefix[:], []byte{0x7f, 'E', 'L', 'F'})
	case "windows":
		valid = prefix[0] == 'M' && prefix[1] == 'Z'
	case "darwin":
		for _, magic := range [][4]byte{{0xcf, 0xfa, 0xed, 0xfe}, {0xce, 0xfa, 0xed, 0xfe}, {0xfe, 0xed, 0xfa, 0xcf}, {0xfe, 0xed, 0xfa, 0xce}, {0xca, 0xfe, 0xba, 0xbe}} {
			if prefix == magic {
				valid = true
			}
		}
	}
	if !valid {
		return errors.New("selected cache artifact is not a native executable; explicit repair required")
	}
	return nil
}

// WrapperCommand keeps lifecycle selection in one stable, offline entrypoint.
// Native workers deliberately use their already selected executable instead.
func WrapperCommand(wrapper string, args ...string) []string {
	if strings.EqualFold(filepath.Ext(wrapper), ".ps1") {
		shell := "powershell.exe"
		if runtime.GOOS == "windows" {
			shell = filepath.Join(os.Getenv("SystemRoot"), "System32", "WindowsPowerShell", "v1.0", "powershell.exe")
		}
		result := []string{shell, "-NoLogo", "-NoProfile", "-NonInteractive", "-File", wrapper}
		return append(result, args...)
	}
	return append([]string{wrapper}, args...)
}

func validateCommonWrapper(ctx context.Context, wrapper string, entry adaptersconfig.Entrypoint) error {
	body, err := readEntrypointFile(wrapper, 128*1024)
	if err != nil {
		return fmt.Errorf("common wrapper unavailable; explicitly refresh setup: %w", err)
	}
	if len(body) > 128*1024 || !bytes.Contains(body, []byte("AgenticLens common entrypoint v1")) {
		return errors.New("unsupported common wrapper; refresh owned user-level setup")
	}
	command := WrapperCommand(wrapper, "--al-offline", "--al-project", entry.Project, "--al-resolve")
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	output, err := installCommandOutput(probeCtx, command[0], command[1:]...)
	if err != nil {
		return fmt.Errorf("common wrapper cannot resolve project pin; refresh setup: %w", err)
	}
	parts := strings.Split(strings.TrimRight(string(output), "\r\n"), "\n")
	if len(parts) != 4 {
		return errors.New("common wrapper resolution protocol incompatible")
	}
	for i := range parts {
		parts[i] = strings.TrimSuffix(parts[i], "\r")
	}
	native, err := entry.NativePath()
	if err != nil {
		return err
	}
	if parts[0] != entry.Version || !sameEntrypointPath(parts[1], entry.Project) || !sameEntrypointPath(parts[2], entry.InstallRoot) || !sameEntrypointPath(parts[3], native) {
		return errors.New("common wrapper resolution differs from shared version/configuration; refresh setup")
	}
	return nil
}

func sameEntrypointPath(left, right string) bool {
	var err error
	left, err = ProjectDirectory(left)
	if err != nil {
		return false
	}
	right, err = ProjectDirectory(right)
	if err != nil {
		return false
	}
	if runtime.GOOS == "windows" {
		return strings.EqualFold(left, right)
	}
	return left == right
}
