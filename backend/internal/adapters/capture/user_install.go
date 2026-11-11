package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
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
	Protocol int          `json:"protocol"`
	Manager  string       `json:"manager"`
	Groups   []OwnedGroup `json:"groups"`
}

// InstallUserCodex owns only unchanged groups in the effective user hooks file.
// The ownership receipt is machine-wide; adoption needs no per-project receipt.
func InstallUserCodex(ctx context.Context, options UserInstallOptions) (InstallResult, error) {
	return installUserCodexWithCheckpoint(ctx, options, nil)
}

func installUserCodexWithCheckpoint(ctx context.Context, options UserInstallOptions, checkpoint func(userInstallStage, int) error) (result InstallResult, returnErr error) {
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
	}
	if !options.Remove && options.Repository != "" {
		if err := checkLegacyUserConflicts(ctx, options.Repository, options.LegacyRepository); err != nil {
			return result, err
		}
	}
	var legacyChanges []fileChange
	homes := []string{home}
	if options.LegacyRepository != "" {
		if options.Remove {
			return result, errors.New("legacy migration requires user-hook installation")
		}
		legacyHomes, err := legacyUserHomes(options.LegacyRepository)
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
		for _, project := range projects {
			directory := filepath.Join(project, ".codex")
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
			fresh, err := legacyUserHomes(options.LegacyRepository)
			if err != nil {
				return result, err
			}
			if err := requireUserLeases(seen, fresh); err != nil {
				return result, err
			}
		}
		if options.LegacyRepository == "" && options.Repository != "" && !options.Remove {
			if err := checkLegacyUserConflicts(ctx, options.Repository, ""); err != nil {
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
		if err := checkLegacyUserConflicts(ctx, options.Repository, options.LegacyRepository); err != nil {
			return result, err
		}
	}
	// Plan migration only after recovery under every affected lease. A durable
	// prefix can have removed hooks or changed the active bit; pre-recovery plans
	// would reject recoverable state or silently leave restored legacy callbacks.
	if options.LegacyRepository != "" {
		legacyChanges, err = legacyUserChanges(options.LegacyRepository)
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
	hooksPath := filepath.Join(home, "hooks.json")
	receiptPath := filepath.Join(home, "agenticlens-user-installation.json")
	receipt := userInstallation{Protocol: userCaptureOwnershipProtocol}
	prior, err := os.ReadFile(receiptPath)
	if err == nil {
		if err := json.Unmarshal(prior, &receipt); err != nil {
			return result, err
		}
		if receipt.Protocol != userCaptureOwnershipProtocol {
			return result, errors.New("unsupported user capture ownership protocol")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	if options.Remove && len(prior) == 0 {
		return result, errors.New("no owned user capture installation; refusing removal")
	}
	content, err := os.ReadFile(hooksPath)
	exists := err == nil
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return result, err
	}
	config := map[string]json.RawMessage{}
	if exists {
		if err := json.Unmarshal(content, &config); err != nil {
			return result, fmt.Errorf("user hooks invalid; preserved: %w", err)
		}
	}
	hooks := map[string][]json.RawMessage{}
	if raw := config["hooks"]; len(raw) > 0 {
		if err := json.Unmarshal(raw, &hooks); err != nil {
			return result, err
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
			return result, fmt.Errorf("owned user %s hook changed or duplicated; preserved", owned.Event)
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
				return result, errors.New("unowned/legacy AgenticLens user hook conflict; preserve and explicitly migrate before installation")
			}
		}
	}
	if receipt.Manager != "" && receipt.Manager != manager {
		return result, errors.New("user manager location differs from owned hooks; explicitly uninstall before relocating")
	}
	next := userInstallation{Protocol: userCaptureOwnershipProtocol, Manager: manager, Groups: []OwnedGroup{}}
	changes := append([]fileChange(nil), legacyChanges...)
	if !options.Remove {
		command, err := staticCommand([]string{manager, "codex-hook"}, runtime.GOOS == "windows")
		if err != nil {
			return result, err
		}
		for _, event := range []string{"SessionStart", "UserPromptSubmit", "SubagentStart", "SubagentStop", "Stop", "SessionEnd"} {
			group, err := codexHookGroup(event, command, "", runtime.GOOS == "windows")
			if err != nil {
				return result, err
			}
			group = json.RawMessage(strings.ReplaceAll(string(group), marker, userHookMarker))
			hooks[event] = append(hooks[event], group)
			next.Groups = append(next.Groups, OwnedGroup{event, group})
		}
	}
	raw, err := json.Marshal(hooks)
	if err != nil {
		return result, err
	}
	config["hooks"] = raw
	after, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return result, err
	}
	changes = append(changes, fileChange{path: hooksPath, content: after, before: content, existed: exists})
	body, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return result, err
	}
	changes = append(changes, fileChange{path: receiptPath, content: body, before: prior, existed: len(prior) > 0})
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
		if change.remove {
			action = "remove"
		}
		result.Changes = append(result.Changes, PlannedChange{change.path, action})
	}
	result.Manifest = Manifest{Source: "codex", ConfigPath: hooksPath, Binary: manager, Active: len(next.Groups) > 0, Capability: "User hooks installed; native trust and runtime coverage require observed evidence"}
	if options.DryRun {
		return result, nil
	}
	if err := ctx.Err(); err != nil {
		return result, err
	}

	// Verify previews under the exclusive machine installation lock.
	for _, change := range changes {
		current, err := os.ReadFile(change.path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return result, err
		}
		if (err == nil) != change.existed || string(current) != string(change.before) {
			return result, errors.New("user installation changed during planning; retry")
		}
	}
	if err := applyUserInstallationWithCheckpoint(home, coordinator, options.LegacyRepository, changes, checkpoint); err != nil {
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
