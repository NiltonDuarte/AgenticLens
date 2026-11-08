package config

import (
	"context"
	"debug/elf"
	"debug/macho"
	"debug/pe"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"

	"agenticlens/internal/domain"
)

// ValidateNativeExecutable rejects non-native and incompatible artifacts before
// running a bounded interface probe. Version metadata reports the actual build.
func ValidateNativeExecutable(path string) (domain.BuildInfo, error) {
	var identity domain.BuildInfo
	if err := validateNativeFormat(path); err != nil {
		return identity, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	protocol, err := exec.CommandContext(ctx, path, "capture", "protocol").Output()
	if err != nil {
		return identity, fmt.Errorf("native capture interface: %w", err)
	}
	if strings.TrimSpace(string(protocol)) != "3" {
		return identity, errors.New("unsupported native capture interface")
	}
	dispatcher, err := exec.CommandContext(ctx, path, "dispatch", "protocol").Output()
	if err != nil {
		return identity, fmt.Errorf("native dispatcher interface: %w", err)
	}
	if strings.TrimSpace(string(dispatcher)) != "1" {
		return identity, errors.New("unsupported native dispatcher interface")
	}
	metadata, err := exec.CommandContext(ctx, path, "version", "--json").Output()
	if err != nil {
		return identity, fmt.Errorf("native build identity: %w", err)
	}
	if err := json.Unmarshal(metadata, &identity); err != nil {
		return identity, fmt.Errorf("native build identity: %w", err)
	}
	if strings.TrimSpace(identity.Version) == "" {
		return identity, errors.New("native build has no version identity")
	}
	return identity, nil
}

// ValidateManualExecutable preserves explicitly provisioned legacy manual CLI
// support without making capture installation an execution prerequisite.
func ValidateManualExecutable(path string) (domain.BuildInfo, error) {
	var identity domain.BuildInfo
	if err := validateNativeFormat(path); err != nil {
		return identity, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	version, err := exec.CommandContext(ctx, path, "version", "--short").Output()
	if err != nil {
		return identity, fmt.Errorf("native release identity: %w", err)
	}
	identity.Version = strings.TrimSpace(string(version))
	if err := ValidateReleaseVersion(identity.Version); err != nil {
		return identity, fmt.Errorf("native release identity: %w", err)
	}
	return identity, nil
}

func validateNativeFormat(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || (runtime.GOOS != "windows" && info.Mode().Perm()&0111 == 0) {
		return errors.New("native output must be a regular executable")
	}
	valid := false
	switch runtime.GOOS {
	case "linux":
		file, err := elf.Open(path)
		if err != nil {
			return err
		}
		valid = (runtime.GOARCH == "amd64" && file.Machine == elf.EM_X86_64) || (runtime.GOARCH == "arm64" && file.Machine == elf.EM_AARCH64)
		if err := file.Close(); err != nil {
			return err
		}
	case "darwin":
		file, err := macho.Open(path)
		if err != nil {
			return err
		}
		valid = (runtime.GOARCH == "amd64" && file.Cpu == macho.CpuAmd64) || (runtime.GOARCH == "arm64" && file.Cpu == macho.CpuArm64)
		if err := file.Close(); err != nil {
			return err
		}
	case "windows":
		file, err := pe.Open(path)
		if err != nil {
			return err
		}
		valid = runtime.GOARCH == "amd64" && file.Machine == pe.IMAGE_FILE_MACHINE_AMD64
		if err := file.Close(); err != nil {
			return err
		}
	}
	if !valid {
		return errors.New("native output platform does not match executing host")
	}
	return nil
}
