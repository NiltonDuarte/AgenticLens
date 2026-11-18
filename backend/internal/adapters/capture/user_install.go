package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
)

const userHookMarker = "AgenticLens native user capture v1"
const userCaptureOwnershipProtocol = 2

type UserInstallOptions struct {
	Repository       string
	LegacyRepository string
	Home             string
	Entrypoint       string
	Remove           bool
	DryRun           bool
}

type userInstallation struct {
	Protocol           int          `json:"protocol"`
	Manager            string       `json:"manager"`
	Groups             []OwnedGroup `json:"groups"`
	BindingInbox       bool         `json:"binding_inbox,omitempty"`
	OwnsInboxGrant     bool         `json:"owns_inbox_grant,omitempty"`
	CreatedInboxConfig bool         `json:"created_inbox_config,omitempty"`
}

// InstallUserCodex owns only unchanged groups in the effective user hooks file.
// The ownership receipt is machine-wide; adoption needs no per-project receipt.
func InstallUserCodex(ctx context.Context, options UserInstallOptions) (InstallResult, error) {
	return installUserCodexWithCheckpoint(ctx, options, nil)
}

func installUserCodexWithCheckpoint(ctx context.Context, options UserInstallOptions, checkpoint func(userInstallStage, int) error) (result InstallResult, returnErr error) {
	return installUser(ctx, options, "codex", checkpoint)
}

func installUser(ctx context.Context, options UserInstallOptions, source string, checkpoint func(userInstallStage, int) error) (result InstallResult, returnErr error) {
	result = InstallResult{DryRun: options.DryRun, Changes: []PlannedChange{}}
	home, err := absolute(options.Home)
	if err != nil {
		return result, err
	}
	root := home
	coordinator := home
	manager := options.Entrypoint
	if manager == "" {
		manager, err = adaptersconfig.UserManagerPath()
		if err != nil {
			return result, err
		}
	}
	manager, err = absolute(manager)
	if err != nil {
		return result, err
	}
	if !options.Remove {
		if err := validateManagerEntrypoint(manager); err != nil {
			return result, fmt.Errorf("native manager unavailable: %w", err)
		}
		if source == "opencode" {
			if err := validateInstallNativeOutput(ctx, manager, "1", "OpenCode user hook protocol", "manager", "opencode-hook-protocol"); err != nil {
				return result, fmt.Errorf("install a manager with OpenCode user hook support: %w", err)
			}
		}
	}
	if !options.Remove && options.Repository != "" {
		if err := checkUserConflicts(ctx, source, options.Repository, options.LegacyRepository); err != nil {
			return result, err
		}
	}
	var legacyChanges []fileChange
	homes := []string{home}
	if options.LegacyRepository != "" {
		if options.Remove {
			return result, errors.New("legacy migration requires user-hook installation")
		}
		legacyHomes, err := legacyHomesFor(source, options.LegacyRepository)
		if err != nil {
			return result, err
		}
		homes = append(homes, legacyHomes...)
		coordinator, err = absolute(legacyHomes[0])
		if err != nil {
			return result, err
		}
	}

	if options.Repository != "" && !options.Remove {
		projects, err := legacyUserProjects(ctx, options.Repository)
		if err != nil {
			return result, err
		}
		leasedProjects := map[string]bool{}
		for _, project := range projects {
			if source == "opencode" && !options.DryRun && !leasedProjects[project] {
				lock, err := lockOpenCodeProject(project)
				if err != nil {
					return result, err
				}
				leasedProjects[project] = true
				defer func() { returnErr = errors.Join(returnErr, lock.Close()) }()
			}
			directory := filepath.Join(project, "."+source)
			info, err := os.Stat(directory)
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				return result, err
			}
			if info.IsDir() {
				homes = append(homes, directory)
			}
		}
	}

	heldUserResources := map[string]bool{}
	if !options.DryRun {
		// Discover pending authority before creating a fresh home. A shared
		// reference to a deleted private owner must remain unknown; mkdir would
		// otherwise make its missing journal look like a completed transaction.
		locks := append([]string(nil), homes...)
		discovered := map[string]bool{}
		for index := 0; index < len(locks); index++ {
			directory, err := absolute(locks[index])
			if err != nil {
				return result, err
			}
			if discovered[directory] {
				continue
			}
			discovered[directory] = true
			pending, err := pendingUserHomes(directory)
			if err != nil {
				return result, err
			}
			locks = append(locks, pending...)
		}
		// Every OpenCode instance loads its XDG directory. Coordinate activation
		// there even when writing the plugin into an additional custom directory.
		// The primary-home lease below covers default-only installs and recovery.
		var loadedHomeLease string
		if source == "opencode" {
			homes, err := loadedOpenCodeHomes(home)
			if err != nil {
				return result, err
			}
			if len(homes) > 1 {
				loadedHomeLease = homes[1]
				lease, err := acquireInstallLock(loadedHomeLease)
				if err != nil {
					return result, err
				}
				defer func() { returnErr = errors.Join(returnErr, lease.Close()) }()
			}
		}
		sort.Strings(locks)
		if err := os.MkdirAll(root, 0700); err != nil {
			return result, err
		}
		machineLock, err := acquireInstallLock(root)
		if err != nil {
			return result, err
		}
		defer func() { returnErr = errors.Join(returnErr, machineLock.Close()) }()
		if checkpoint != nil {
			if err := checkpoint(userInstallResourcesDiscovered, -1); err != nil {
				return result, err
			}
		}
		seen := map[string]bool{root: true}
		heldUserResources = seen
		for _, directory := range locks {
			directory, err = absolute(directory)
			if err != nil {
				return result, err
			}
			if seen[directory] {
				continue
			}
			seen[directory] = true
			if directory == loadedHomeLease {
				// A pending migration may also name this already-held home.
				continue
			}
			if directory == home {
				if err := os.MkdirAll(directory, 0700); err != nil {
					return result, err
				}
			}
			if directory != home {
				info, err := os.Stat(directory)
				if err != nil {
					return result, fmt.Errorf("installation recovery resource unavailable; pending journal preserved: %w", err)
				}
				if !info.IsDir() {
					return result, fmt.Errorf("installation resource is not a directory: %s", directory)
				}
			}
			resourceLock, err := acquireInstallLock(directory)
			if err != nil {
				return result, err
			}
			defer func() { returnErr = errors.Join(returnErr, resourceLock.Close()) }()
		}
		// A journal can appear or acquire new resources before its home lease is
		// held. Pre-lock discovery is only a lock plan, never recovery authority.
		for journalDirectory := range seen {
			fresh, err := pendingUserHomes(journalDirectory)
			if err != nil {
				return result, err
			}
			if err := requireUserLeases(seen, fresh); err != nil {
				return result, err
			}
		}
		if options.LegacyRepository != "" {
			fresh, err := legacyHomesFor(source, options.LegacyRepository)
			if err != nil {
				return result, err
			}
			if err := requireUserLeases(seen, fresh); err != nil {
				return result, err
			}
		}
		if options.LegacyRepository == "" && options.Repository != "" && !options.Remove {
			if err := checkUserConflicts(ctx, source, options.Repository, ""); err != nil {
				return result, err
			}
		}
		// No recovery writes occur until all current and prior journal homes are
		// excluded, including a prior invocation's different effective Codex home.
		for directory := range seen {
			if err := validateUserRecovery(directory); err != nil {
				return result, err
			}
		}
		for directory := range seen {
			if err := recoverUserInstallation(directory); err != nil {
				return result, err
			}
		}

	}

	if !options.Remove && options.Repository != "" {
		if err := checkUserConflicts(ctx, source, options.Repository, options.LegacyRepository); err != nil {
			return result, err
		}
	}
	// Plan migration only after recovery under every affected lease. A durable
	// prefix can have removed hooks or changed the active bit; pre-recovery plans
	// would reject recoverable state or silently leave restored legacy callbacks.
	if options.LegacyRepository != "" {
		legacyChanges, err = legacyChangesFor(source, options.LegacyRepository)
		if err != nil {
			return result, err
		}
		if !options.DryRun {
			required := userJournalHomes(userInstallJournal{Changes: asUserJournalChanges(legacyChanges)})
			if err := requireUserLeases(heldUserResources, required); err != nil {
				return result, err
			}
		}
	}
	var own []fileChange
	if source == "codex" {
		own, result.Manifest, err = planCodexUser(home, manager, options.Repository, options.Remove)
	} else {
		own, result.Manifest, err = planOpenCodeUser(home, manager, options.Remove)
	}
	if err != nil {
		return result, err
	}
	changes := append(legacyChanges, own...)
	// Preserve file identity when owned hook artifacts are unchanged.
	effective := make([]fileChange, 0, len(changes))
	for _, change := range changes {
		if !change.remove && change.existed && bytes.Equal(change.before, change.content) {
			continue
		}
		effective = append(effective, change)
	}
	changes = effective
	for _, change := range changes {
		action := "write"
		if change.mkdir {
			action = "mkdir"
		} else if change.remove {
			action = "remove"
		}
		result.Changes = append(result.Changes, PlannedChange{change.path, action})
	}

	if options.DryRun {
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}

	// Verify previews under the exclusive machine installation lock.
	for _, change := range changes {
		if change.mkdir {
			if _, err := os.Lstat(change.path); !errors.Is(err, os.ErrNotExist) {
				return result, errors.New("inbox changed during planning; retry")
			}
			continue
		}
		if filepath.Base(change.path) == "config.toml" {
			data, mode, exists, err := readCodexConfigFile(change.path)
			if err != nil || exists != change.existed || mode != change.mode || !bytes.Equal(data, change.before) {
				return result, errors.Join(errors.New("Codex config.toml changed during planning; retry"), err)
			}
			continue
		}
		current, err := os.ReadFile(change.path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return result, err
		}
		if (err == nil) != change.existed || string(current) != string(change.before) {
			return result, errors.New("user installation changed during planning; retry")
		}
	}
	if err := applyUserInstallationWithCheckpoint(source, home, coordinator, options.LegacyRepository, changes, checkpoint); err != nil {
		return result, err
	}
	return result, nil
}

func requireUserLeases(held map[string]bool, required []string) error {
	for _, directory := range required {
		canonical, err := absolute(directory)
		if err != nil {
			return err
		}
		if !held[canonical] {
			return fmt.Errorf("installation recovery resource set changed; retry under fresh leases; unheld resource %s", canonical)
		}
	}
	return nil
}

func asUserJournalChanges(changes []fileChange) []userInstallChange {
	result := make([]userInstallChange, 0, len(changes))
	for _, change := range changes {
		result = append(result, userInstallChange{Path: change.path})
	}
	return result
}

func planCodexUser(home, manager, project string, remove bool) ([]fileChange, Manifest, error) {
	hooksPath := filepath.Join(home, "hooks.json")
	receiptPath := filepath.Join(home, "agenticlens-user-installation.json")
	receipt := userInstallation{Protocol: userCaptureOwnershipProtocol}
	prior, _, hasReceipt, err := readCodexConfigFile(receiptPath)
	if err != nil {
		return nil, Manifest{}, err
	}
	if hasReceipt {
		if err := json.Unmarshal(prior, &receipt); err != nil {
			return nil, Manifest{}, err
		}
		if receipt.Protocol != userCaptureOwnershipProtocol {
			return nil, Manifest{}, errors.New("unsupported user capture ownership protocol")
		}
	}
	if remove && len(prior) == 0 {
		return nil, Manifest{}, errors.New("no owned user capture installation; refusing removal")
	}
	content, _, exists, err := readCodexConfigFile(hooksPath)
	if err != nil {
		return nil, Manifest{}, err
	}
	config := map[string]json.RawMessage{}
	if exists {
		if err := json.Unmarshal(content, &config); err != nil {
			return nil, Manifest{}, fmt.Errorf("user hooks invalid; preserved: %w", err)
		}
	}
	hooks := map[string][]json.RawMessage{}
	if raw := config["hooks"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			return nil, Manifest{}, err
		}
	}
	for _, owned := range receipt.Groups {
		kept := []json.RawMessage{}
		matches := 0
		for _, group := range hooks[owned.Event] {
			if jsonEqual(group, owned.Value) {
				matches++
			} else {
				kept = append(kept, group)
			}
		}
		if matches != 1 {
			return nil, Manifest{}, fmt.Errorf("owned user %s hook changed or duplicated; preserved", owned.Event)
		}
		if len(kept) == 0 {
			delete(hooks, owned.Event)
		} else {
			hooks[owned.Event] = kept
		}
	}
	for _, groups := range hooks {
		for _, group := range groups {
			if strings.Contains(string(group), userHookMarker) || strings.Contains(string(group), marker) {
				return nil, Manifest{}, errors.New("unowned/legacy AgenticLens user hook conflict; preserve and explicitly migrate before installation")
			}
		}
	}
	if receipt.Manager != "" && receipt.Manager != manager {
		return nil, Manifest{}, errors.New("user manager location differs from owned hooks; explicitly uninstall before relocating")
	}
	next := userInstallation{Protocol: userCaptureOwnershipProtocol, Manager: manager, Groups: []OwnedGroup{}}
	changes, err := planCodexInbox(home, project, remove, receipt, &next)
	if err != nil {
		return nil, Manifest{}, err
	}
	if !remove {
		command, err := staticCommand([]string{manager, "codex-hook"})
		if err != nil {
			return nil, Manifest{}, err
		}
		for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"} {
			group, err := codexHookGroup(event, command)
			if err != nil {
				return nil, Manifest{}, err
			}
			group = json.RawMessage(strings.ReplaceAll(string(group), marker, userHookMarker))
			hooks[event] = append(hooks[event], group)
			next.Groups = append(next.Groups, OwnedGroup{event, group})
		}
	}
	raw, err := json.Marshal(hooks)
	if err != nil {
		return nil, Manifest{}, err
	}
	config["hooks"] = raw
	after, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return nil, Manifest{}, err
	}
	changes = append(changes, fileChange{path: hooksPath, content: after, before: content, existed: exists})
	body, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return nil, Manifest{}, err
	}
	changes = append(changes, fileChange{path: receiptPath, content: body, before: prior, existed: len(prior) > 0})
	return changes, Manifest{Source: "codex", ConfigPath: hooksPath, Binary: manager, Active: len(next.Groups) > 0, Capability: "User hooks installed; native trust and runtime coverage require observed evidence"}, nil
}
