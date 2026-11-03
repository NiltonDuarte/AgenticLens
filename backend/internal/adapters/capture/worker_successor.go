package capture

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
)

// WorkerSuccessor selects one already installed native for the exact existing
// state namespace. It never dispatches a wrapper, downloads or migrates state.
// Incompatible or unavailable installations leave their generation pending.
func WorkerSuccessor(ctx context.Context, project, state string) (string, error) {
	entry, err := adaptersconfig.ResolveEntrypoint(project)
	if err != nil {
		return "", err
	}
	policy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return "", err
	}
	if !policy.Enabled {
		return "", errors.New("automatic capture disabled; pending work retained")
	}
	configured, err := PreviewStateLocation(policy.StateDir, project)
	if err != nil {
		return "", err
	}
	if configured.Provisional || !sameEntrypointPath(configured.CoordinationDirectory, state) {
		return "", errors.New("capture state directory changed; pending work retained in its original namespace for explicit recovery")
	}
	native, err := entry.NativePath()
	if err != nil {
		return "", err
	}
	if err = validateNativeHeader(native); err != nil {
		return "", fmt.Errorf("selected worker native unavailable: %w", err)
	}
	active := false
	installations := map[string]string{}
	for _, source := range []string{"codex", "opencode"} {
		manifest, err := ReadManifest(state, source)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if !manifest.Active {
			continue
		}
		active = true
		identity, _, err := InstallationIdentity(state, source)
		if err != nil {
			return "", err
		}
		installations[source] = identity
		if manifest.Protocol != ManifestProtocolVersion || manifest.Version != entry.Version || !sameEntrypointPath(manifest.Binary, native) || !sameEntrypointPath(manifest.Repository, project) || !sameEntrypointPath(manifest.StateDir, state) || manifest.TemplateVersion != installationTemplateVersion() || manifest.Wrapper == "" || manifest.SourceStore == "" {
			return "", errors.New("selected worker installation is incompatible; explicitly refresh setup and recover pending work")
		}
		if err := CheckWrapperCompatibility(manifest.Wrapper, manifest.WrapperHash); err != nil {
			return "", fmt.Errorf("selected worker wrapper incompatible; pending work retained: %w", err)
		}
	}
	if !active {
		return "", errors.New("no active compatible installation for selected worker; pending work retained")
	}
	version, err := exec.CommandContext(ctx, native, "version", "--short").Output()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(version)) != entry.Version {
		return "", errors.New("selected worker native version differs from project pin")
	}
	protocol, err := exec.CommandContext(ctx, native, "capture", "protocol").Output()
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(string(protocol)) != strconv.Itoa(HandlerProtocolVersion) {
		return "", errors.New("selected worker lifecycle protocol is incompatible; pending work retained")
	}
	// Selection may change while the bounded native probes run. Do not launch
	// a now-obsolete successor or reinterpret work in a different namespace.
	current, err := adaptersconfig.ResolveEntrypoint(project)
	if err != nil {
		return "", err
	}
	selected, err := current.NativePath()
	if err != nil {
		return "", err
	}
	if current.Version != entry.Version || !sameEntrypointPath(selected, native) {
		return "", errors.New("project release changed during native takeover; pending work retained")
	}
	currentPolicy, err := adaptersconfig.ResolveCapture(project, "")
	if err != nil {
		return "", err
	}
	currentState, err := PreviewStateLocation(currentPolicy.StateDir, project)
	if err != nil {
		return "", err
	}
	if !currentPolicy.Enabled || currentState.Provisional || !sameEntrypointPath(currentState.CoordinationDirectory, state) {
		return "", errors.New("capture policy/state changed during native takeover; pending work retained")
	}
	remaining := false
	for _, source := range []string{"codex", "opencode"} {
		identity, _, err := InstallationIdentity(state, source)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return "", err
		}
		if identity == "" {
			continue
		}
		remaining = true
		if identity != installations[source] {
			return "", errors.New("active capture installation changed during native takeover; pending work retained")
		}
	}
	if !remaining {
		return "", errors.New("capture integrations disabled during native takeover; pending work retained")
	}
	return native, nil
}
