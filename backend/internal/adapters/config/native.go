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
	role, err := exec.CommandContext(ctx, path, "engine", "protocol").Output()
	if err != nil {
		return identity, fmt.Errorf("native engine interface: %w", err)
	}
	if strings.TrimSpace(string(role)) != "1" {
		return identity, errors.New("unsupported native engine interface")
	}
	protocol, err := exec.CommandContext(ctx, path, "capture", "protocol").Output()
	if err != nil {
		return identity, fmt.Errorf("native capture interface: %w", err)
	}
	if strings.TrimSpace(string(protocol)) != "3" {
		return identity, errors.New("unsupported native capture interface")
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

// ValidateManualExecutable validates an explicitly provisioned engine without
// making capture installation an execution prerequisite.
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

// ValidateManagerExecutable admits the separate native installation/router
// command. It must report its own identity, never a selected engine's version.
func ValidateManagerExecutable(path string) (domain.BuildInfo, error) {
	var identity domain.BuildInfo
	if err := validateNativeFormat(path); err != nil {
		return identity, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	protocol, err := exec.CommandContext(ctx, path, "manager", "protocol").Output()
	if err != nil {
		return identity, fmt.Errorf("native manager interface: %w", err)
	}
	if strings.TrimSpace(string(protocol)) != "1" {
		return identity, fmt.Errorf("native manager interface incompatible: %q", strings.TrimSpace(string(protocol)))
	}
	metadata, err := exec.CommandContext(ctx, path, "--manager-version", "--json").Output()
	if err != nil {
		return identity, fmt.Errorf("native manager identity: %w", err)
	}
	if err := json.Unmarshal(metadata, &identity); err != nil {
		return identity, fmt.Errorf("native manager identity: %w", err)
	}
	if err := ValidateReleaseVersion(identity.Version); err != nil {
		return identity, fmt.Errorf("native manager identity: %w", err)
	}
	return identity, nil
}

func validateNativeFormat(path string) error {
	return ValidateNativePlatform(path, runtime.GOOS, runtime.GOARCH)
}

// ValidateNativePlatform checks native format and machine without executing it.
func ValidateNativePlatform(path, osName, arch string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !info.Mode().IsRegular() || (osName != "windows" && info.Mode().Perm()&0111 == 0) {
		return errors.New("native output must be a regular executable")
	}
	valid := false
	switch osName {
	case "linux":
		file, err := elf.Open(path)
		if err != nil {
			return err
		}
		valid = (arch == "amd64" && file.Machine == elf.EM_X86_64) || (arch == "arm64" && file.Machine == elf.EM_AARCH64)
		if err := file.Close(); err != nil {
			return err
		}
	case "darwin":
		file, err := macho.Open(path)
		if err != nil {
			return err
		}
		valid = (arch == "amd64" && file.Cpu == macho.CpuAmd64) || (arch == "arm64" && file.Cpu == macho.CpuArm64)
		if err := file.Close(); err != nil {
			return err
		}
	case "windows":
		file, err := pe.Open(path)
		if err != nil {
			return err
		}
		valid = arch == "amd64" && file.Machine == pe.IMAGE_FILE_MACHINE_AMD64
		if err := file.Close(); err != nil {
			return err
		}
	}
	if !valid {
		return errors.New("native output platform does not match requested platform")
	}
	return nil
}

// ValidateNativeFormat checks the artifact's executable host format before an
// installer or manager starts any interface probe.
func ValidateNativeFormat(path string) error { return validateNativeFormat(path) }
