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
)

const userHookMarker = "AgenticLens native user capture v1"

type UserInstallOptions struct {
	Repository       string
	DispatcherOnly   bool
	LegacyRepository string
	Home             string
	Binary           string
	InstallRoot      string
	Remove           bool
	DryRun           bool
}

// dispatcherInstallation is the sole machine-cache ownership authority. Hook
// receipts may record historical build hashes but cannot veto another home's
// valid dispatcher update or authorize bytes against this canonical record.
type dispatcherInstallation struct {
	Protocol     int    `json:"protocol"`
	Dispatcher   string `json:"dispatcher"`
	Hash         string `json:"hash"`
	ActiveNative string `json:"active_native,omitempty"`
	ActiveHash   string `json:"active_hash,omitempty"`
}

type userInstallation struct {
	Protocol       int          `json:"protocol"`
	Dispatcher     string       `json:"dispatcher"`
	DispatcherHash string       `json:"dispatcher_hash"`
	Groups         []OwnedGroup `json:"groups"`
}

// EffectiveCodexHome is the harness configuration home, independently of any
// project's configured telemetry store. Installation never freezes that store.
func EffectiveCodexHome() (string, error) {
	if value := os.Getenv("CODEX_HOME"); value != "" {
		return absolute(value)
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}
	return filepath.Join(home, ".codex"), nil
}

// InstallUserCodex owns only unchanged groups in the effective user hooks file.
// The ownership receipt is machine-wide; adoption needs no per-project receipt.
func InstallUserCodex(ctx context.Context, options UserInstallOptions) (InstallResult, error) {
	return installUserCodexWithCheckpoint(ctx, options, nil)
}

func installUserCodexWithCheckpoint(ctx context.Context, options UserInstallOptions, checkpoint func(userInstallStage, int) error) (result InstallResult, returnErr error) {
	result = InstallResult{DryRun: options.DryRun, Changes: []PlannedChange{}}
	var admittedPath string
	var admittedBytes []byte
	if !options.Remove {
		staging, err := os.MkdirTemp("", "agenticlens-native-admission-")
		if err != nil {
			return result, err
		}
		defer func() { returnErr = errors.Join(returnErr, os.RemoveAll(staging)) }()
		admittedPath, err = RetainExecutable(options.Binary, staging)
		if err != nil {
			return result, fmt.Errorf("replacement native rejected before installation: %w", err)
		}
		admittedBytes, err = os.ReadFile(admittedPath)
		if err != nil {
			return result, err
		}
	}
	home, err := absolute(options.Home)
	if err != nil {
		return result, err
	}
	root, err := absolute(options.InstallRoot)
	if err != nil {
		return result, err
	}
	if !options.DispatcherOnly && !options.Remove && options.Repository != "" {
		if err := checkLegacyUserConflicts(ctx, options.Repository, options.LegacyRepository); err != nil {
			return result, err
		}
	}
	var legacyChanges []fileChange
	homes := []string{home}
	if options.LegacyRepository != "" {
		if options.Remove || options.DispatcherOnly {
			return result, errors.New("legacy migration requires user-hook installation")
		}
		legacyHomes, err := legacyUserHomes(options.LegacyRepository)
		if err != nil {
			return result, err
		}
		homes = append(homes, legacyHomes...)
	}

	heldUserResources := map[string]bool{}
	if !options.DryRun {
		if err := os.MkdirAll(root, 0700); err != nil {
			return result, err
		}
		machineLock, err := acquireInstallLock(root)
		if err != nil {
			return result, err
		}
		defer func() { returnErr = errors.Join(returnErr, machineLock.Close()) }()
		pending, err := pendingUserHomes(root)
		if err != nil {
			return result, err
		}
		prior, err := pendingUserHomes(home)
		if err != nil {
			return result, err
		}
		locks := append(append(append([]string(nil), homes...), pending...), prior...)
		sort.Strings(locks)
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
		for _, journalDirectory := range []string{root, home} {
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
		// No recovery writes occur until all current and prior journal homes are
		// excluded, including a prior invocation's different effective Codex home.
		if err := recoverUserInstallation(root); err != nil {
			return result, err
		}
		if root != home {
			if err := recoverUserInstallation(home); err != nil {
				return result, err
			}
		}
	}

	if !options.DispatcherOnly && !options.Remove && options.Repository != "" {
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
	receipt := userInstallation{Protocol: 1}
	prior, err := os.ReadFile(receiptPath)
	if err == nil {
		if err := json.Unmarshal(prior, &receipt); err != nil {
			return result, err
		}
		if receipt.Protocol != 1 {
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
		if !options.DispatcherOnly {
			if len(kept) == 0 {
				delete(hooks, owned.Event)
			} else {
				hooks[owned.Event] = kept
			}
		}
	}
	if !options.DispatcherOnly {
		for _, groups := range hooks {
			for _, group := range groups {
				if strings.Contains(string(group), userHookMarker) || strings.Contains(string(group), marker) {
					return result, errors.New("unowned/legacy AgenticLens user hook conflict; preserve and explicitly migrate before installation")
				}
			}
		}
	}
	dispatcher := filepath.Join(root, "dispatcher")
	if runtime.GOOS == "windows" {
		dispatcher += ".exe"
	}
	if receipt.Dispatcher != "" && receipt.Dispatcher != dispatcher {
		return result, errors.New("user dispatcher location differs from owned installation; explicitly remove before relocating")
	}
	ownerPath := filepath.Join(root, "dispatcher-installation.json")
	ownerBefore, ownerErr := os.ReadFile(ownerPath)
	if ownerErr != nil && !errors.Is(ownerErr, os.ErrNotExist) {
		return result, ownerErr
	}
	owner := dispatcherInstallation{Protocol: 1, Dispatcher: dispatcher}
	if ownerErr == nil {
		if err := json.Unmarshal(ownerBefore, &owner); err != nil {
			return result, err
		}
		if owner.Protocol != 1 || owner.Dispatcher != dispatcher || owner.Hash == "" {
			return result, errors.New("native dispatcher ownership record invalid; preserved")
		}
	}
	if current, err := os.ReadFile(dispatcher); err == nil {
		expected := owner.Hash
		if ownerErr != nil && receipt.Dispatcher == dispatcher && receipt.DispatcherHash != "" {
			// Promote the earlier full hook receipt only after exact old ownership
			// verification. A checksum sidecar alone never authorizes adoption.
			expected = receipt.DispatcherHash
		}
		if expected == "" || hash(current) != expected {
			return result, errors.New("native dispatcher was modified or is unowned; preserved")
		}
		owner.Hash = expected
	} else if !errors.Is(err, os.ErrNotExist) {
		return result, err
	}

	next := userInstallation{Protocol: 1, Dispatcher: dispatcher, DispatcherHash: owner.Hash}
	if options.DispatcherOnly {
		next.Groups = append([]OwnedGroup(nil), receipt.Groups...)
	}
	changes := []fileChange{}
	changes = append(changes, legacyChanges...)

	if !options.Remove {
		binary := admittedBytes

		if runtime.GOOS == "windows" {
			// Windows cannot replace a loaded dispatcher image. Keep the stable
			// launcher immutable and activate retained build bytes in the root record.
			owner.ActiveHash = hash(binary)
			owner.ActiveNative = filepath.Join(root, "native", owner.ActiveHash+".exe")
			if !options.DryRun {
				retained, err := RetainExecutable(admittedPath, root)
				if err != nil {
					return result, err
				}
				owner.ActiveNative = retained
			}
			_, launcherErr := os.Stat(dispatcher)
			if launcherErr != nil && !errors.Is(launcherErr, os.ErrNotExist) {
				return result, launcherErr
			}
			next.DispatcherHash = stableDispatcherHash(owner.Hash, owner.ActiveHash, launcherErr == nil)
		} else {
			next.DispatcherHash = hash(binary)
		}
		if !options.DispatcherOnly {
			command, err := staticCommand([]string{dispatcher, "dispatch", "codex-hook"}, runtime.GOOS == "windows")
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
		before, readErr := os.ReadFile(dispatcher)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return result, readErr
		}
		if runtime.GOOS != "windows" || readErr != nil {
			changes = append(changes, fileChange{path: dispatcher, content: binary, before: before, existed: readErr == nil})
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
	if !options.Remove {
		checksumPath := dispatcher + ".sha256"
		before, readErr := os.ReadFile(checksumPath)
		if readErr != nil && !errors.Is(readErr, os.ErrNotExist) {
			return result, readErr
		}
		changes = append(changes, fileChange{path: checksumPath, content: []byte(next.DispatcherHash + "\n"), before: before, existed: readErr == nil})
	}
	if !options.DispatcherOnly {
		changes = append(changes, fileChange{path: hooksPath, content: after, before: content, existed: exists})
	}
	owner.Hash = next.DispatcherHash
	ownerBody, err := json.MarshalIndent(owner, "", "  ")
	if err != nil {
		return result, err
	}
	changes = append(changes, fileChange{path: ownerPath, content: ownerBody, before: ownerBefore, existed: ownerErr == nil})
	body, err := json.MarshalIndent(next, "", "  ")
	if err != nil {
		return result, err
	}
	if !options.DispatcherOnly || len(prior) > 0 {
		changes = append(changes, fileChange{path: receiptPath, content: body, before: prior, existed: len(prior) > 0})
	}
	// Byte-identical owned artifacts retain their file identity. In particular,
	// Windows cannot atomically replace a running dispatcher ancestor on reinstall.
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
	result.Manifest = Manifest{Source: "codex", ConfigPath: hooksPath, Binary: dispatcher, Active: len(next.Groups) > 0, Capability: "User hooks installed; native trust and runtime coverage require observed evidence"}
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
	if err := applyUserInstallationWithCheckpoint(root, dispatcher, changes, homes, checkpoint); err != nil {
		return result, err
	}
	return result, nil
}

// stableDispatcherHash describes exactly the image a successful Windows repair
// leaves at the stable path: a present loaded image stays immutable, while a
// missing image is recreated from the newly admitted candidate.
func stableDispatcherHash(existingHash, candidateHash string, launcherExists bool) string {
	if launcherExists {
		return existingHash
	}
	return candidateHash
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
