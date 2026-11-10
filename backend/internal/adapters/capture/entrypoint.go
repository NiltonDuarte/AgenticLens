package capture

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"runtime"
	"strings"
	"time"

	adaptersconfig "agenticlens/internal/adapters/config"
)

func installationTemplateVersion() string {
	return hash([]byte(marker + pluginTemplate + "native-manager-v1;engine-v1;codex-user2"))
}

const nativeEntrypointLimit = 256 * 1024 * 1024

// CheckManagerCompatibility verifies the common native manager independently of
// the project's historical installation. Owned manager upgrades remain valid.
func CheckManagerCompatibility(wrapper, historicalHash string) error {
	if err := validateManagerEntrypoint(wrapper); err != nil {
		return err
	}
	body, err := readEntrypointFile(wrapper, nativeEntrypointLimit)
	if err != nil {
		return fmt.Errorf("native manager unavailable: %w", err)
	}
	current := hash(body)
	if current == historicalHash {
		return nil
	}
	owned, err := readEntrypointFile(wrapper+".sha256", 256)
	if err != nil || strings.TrimSpace(string(owned)) != current {
		return errors.New("native manager changed without matching installer ownership; edited/unowned command preserved, explicit repair required")
	}
	return nil
}

func validateManagerEntrypoint(path string) error {
	_, err := adaptersconfig.ValidateManagerExecutable(path)
	return err
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

func validateCommonManager(ctx context.Context, wrapper string, entry adaptersconfig.Entrypoint) error {
	if err := validateManagerEntrypoint(wrapper); err != nil {
		return err
	}
	command := []string{wrapper, "--al-offline", "--al-project", entry.Project, "--al-resolve"}
	probeCtx, cancel := context.WithTimeout(ctx, 2*time.Second)
	defer cancel()
	output, err := installCommandOutput(probeCtx, command[0], command[1:]...)
	if err != nil {
		return fmt.Errorf("native manager cannot resolve project pin; refresh setup: %w", err)
	}
	parts := strings.Split(strings.TrimRight(string(output), "\r\n"), "\n")
	if len(parts) != 4 {
		return errors.New("native manager resolution protocol incompatible")
	}
	for i := range parts {
		parts[i] = strings.TrimSuffix(parts[i], "\r")
	}
	native, err := entry.NativePath()
	if err != nil {
		return err
	}
	identity, err := adaptersconfig.ValidateManualExecutable(native)
	if err != nil {
		return err
	}
	if parts[0] != identity.Version || !sameEntrypointPath(parts[1], entry.Project) || !sameEntrypointPath(parts[2], entry.InstallRoot) || !sameEntrypointPath(parts[3], native) {
		return errors.New("native manager resolution differs from shared version/configuration; refresh setup")
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
