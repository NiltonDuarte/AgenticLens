package capture

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	adaptersconfig "agenticlens/internal/adapters/config"
)

const installJournalLimit int64 = 8 * 1024 * 1024

type installIdentity struct {
	Repository, StateDir, Source, SourceStore               string
	ConfigVersion, Version, Binary, Wrapper, LegacyStateDir string
	Remove, AllowLegacy                                     bool
}
type installJournalChange struct {
	Path            string
	Before, After   []byte
	Existed, Remove bool
}
type installJournal struct {
	Protocol       int
	Phase          string
	RollbackPrefix int
	Identity       installIdentity
	Manifest       Manifest
	Changes        []installJournalChange
}

func installJournalPath(state, source string) string {
	return filepath.Join(state, "install-transaction-"+source+".json")
}

func acquireInstallLock(state string) (*PublicationLock, error) {
	if err := os.MkdirAll(state, 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(filepath.Join(state, "install.lock"), os.O_CREATE|os.O_RDWR, 0600)
	if err != nil {
		return nil, err
	}
	held, err := tryPublicationLock(file)
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !held {
		return nil, errors.Join(errors.New("capture installation is already in progress; retry after its owner finishes"), file.Close())
	}
	return &PublicationLock{file}, nil
}

func currentInstallIdentity(ctx context.Context, options InstallOptions, repo, state string, m Manifest) (installIdentity, error) {
	source, err := CanonicalStore(options.SourceStore)
	if err != nil {
		return installIdentity{}, err
	}
	if m.Repository != repo || m.StateDir != state || m.Source != options.Source || m.SourceStore != source || m.Active == options.Remove || m.Protocol != ManifestProtocolVersion {
		return installIdentity{}, errors.New("pending installer manifest does not match the exact requested installation")
	}
	entry, err := adaptersconfig.ResolveEntrypoint(repo)
	if err != nil {
		return installIdentity{}, err
	}
	if m.Active {
		native, err := entry.NativePath()
		if err != nil {
			return installIdentity{}, err
		}
		native, err = ProjectDirectory(native)
		if err != nil {
			return installIdentity{}, err
		}
		if m.Version != entry.Version || m.Binary != native {
			return installIdentity{}, errors.New("pending installer release differs from the shared project pin; restore the recorded selection before recovery")
		}
		if err = validateNativeHeader(native); err != nil {
			return installIdentity{}, err
		}
		version, e := exec.CommandContext(ctx, native, "version", "--short").Output()
		if e != nil || strings.TrimSpace(string(version)) != m.Version {
			return installIdentity{}, errors.New("pending installer native version is unavailable or changed")
		}
		protocol, e := exec.CommandContext(ctx, native, "capture", "protocol").Output()
		if e != nil || strings.TrimSpace(string(protocol)) != strconv.Itoa(HandlerProtocolVersion) {
			return installIdentity{}, errors.New("pending installer native lifecycle capability is unavailable or changed")
		}

		if options.Binary != "" {
			requested, err := ProjectDirectory(options.Binary)
			if err != nil {
				return installIdentity{}, err
			}
			if requested != m.Binary {
				return installIdentity{}, errors.New("pending installer native differs from explicit request")
			}
		}
		wrapper := options.Wrapper
		if wrapper == "" {
			wrapper, err = adaptersconfig.UserWrapperPath()
			if err != nil {
				return installIdentity{}, err
			}
		}
		wrapper, err = ProjectDirectory(wrapper)
		if err != nil {
			return installIdentity{}, err
		}
		if wrapper != m.Wrapper {
			return installIdentity{}, errors.New("pending installer wrapper differs from requested entrypoint")
		}
		content, err := readEntrypointFile(wrapper, 128*1024)
		if err != nil {
			return installIdentity{}, err
		}
		if m.WrapperHash == "" || hash(content) != m.WrapperHash {
			return installIdentity{}, errors.New("pending installer wrapper changed; review and restore it before recovery")
		}
	}
	legacy := ""
	if options.AllowLegacyManifest && options.LegacyStateDir != "" {
		legacy, err = ProjectDirectory(options.LegacyStateDir)
		if err != nil {
			return installIdentity{}, err
		}
	}
	return installIdentity{Repository: repo, StateDir: state, Source: options.Source, SourceStore: source, ConfigVersion: entry.Version, Version: m.Version, Binary: m.Binary, Wrapper: m.Wrapper, LegacyStateDir: legacy, Remove: options.Remove, AllowLegacy: options.AllowLegacyManifest}, nil
}

func installAllowedPaths(repo, state, source string) map[string]bool {
	config := filepath.Join(repo, ".codex", "hooks.json")
	if source == "opencode" {
		config = filepath.Join(repo, ".opencode", "plugins", "agenticlens-capture.ts")
	}
	return map[string]bool{config: true, filepath.Join(repo, ".gitignore"): true, manifestPath(state, source): true, filepath.Join(repo, "."+source, "agenticlens-installation.json"): true}
}

func validateInstallJournal(j installJournal, identity installIdentity) error {
	if j.Protocol != 1 || j.Identity != identity || len(j.Changes) < 3 || len(j.Changes) > 4 || j.Phase != "applying" && j.Phase != "rollback" {
		return errors.New("unsupported or differently owned pending installer transaction")
	}
	if j.RollbackPrefix < 0 || j.RollbackPrefix > len(j.Changes) {
		return errors.New("invalid installer rollback prefix")
	}
	allowed := installAllowedPaths(identity.Repository, identity.StateDir, identity.Source)
	seen := map[string]bool{}
	for _, change := range j.Changes {
		if !allowed[change.Path] || seen[change.Path] {
			return errors.New("pending installer transaction contains an unowned or duplicate target")
		}
		seen[change.Path] = true
		if filepath.Dir(change.Path) != identity.StateDir {
			if err := ValidateDestination(identity.Repository, change.Path); err != nil {
				return err
			}
		}
	}
	config := filepath.Join(identity.Repository, ".codex", "hooks.json")
	if identity.Source == "opencode" {
		config = filepath.Join(identity.Repository, ".opencode", "plugins", "agenticlens-capture.ts")
	}
	if j.Manifest.ConfigPath != config {
		return errors.New("pending installer configuration target differs from its exact source")
	}
	if !seen[config] || !seen[manifestPath(identity.StateDir, identity.Source)] || !seen[filepath.Join(identity.Repository, "."+identity.Source, "agenticlens-installation.json")] {
		return errors.New("pending installer transaction lacks required owned targets")
	}
	body, err := json.Marshal(j.Manifest)
	if err != nil {
		return err
	}
	for _, change := range j.Changes {
		if change.Path == manifestPath(identity.StateDir, identity.Source) || change.Path == filepath.Join(identity.Repository, "."+identity.Source, "agenticlens-installation.json") {
			if change.Remove || !jsonEqual(body, change.After) {
				return errors.New("pending installer transaction manifest does not match intended output")
			}
		}
	}
	return nil
}

func encodeInstallJournal(j installJournal) ([]byte, error) {
	data, err := json.Marshal(j)
	if err != nil {
		return nil, err
	}
	if int64(len(data)) > installJournalLimit {
		return nil, errors.New("installer transaction exceeds private recovery journal limit")
	}
	return append(data, '\n'), nil
}
func readInstallJournal(path string) (installJournal, []byte, error) {
	var j installJournal
	info, err := os.Lstat(path)
	if err != nil {
		return j, nil, err
	}
	if !info.Mode().IsRegular() {
		return j, nil, errors.New("installer recovery journal is not a regular private file")
	}
	data, err := readEntrypointFile(path, installJournalLimit)
	if err != nil {
		return j, nil, err
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&j); err != nil {
		return j, nil, err
	}
	if err = decoder.Decode(new(json.RawMessage)); err != io.EOF {
		return j, nil, errors.New("installer recovery journal has trailing content")
	}
	return j, data, nil
}
func saveInstallJournal(path string, j installJournal, before []byte) error {
	if err := matchInstallJournal(path, before); err != nil {
		return err
	}
	after, err := encodeInstallJournal(j)
	if err != nil {
		return err
	}
	return apply(fileChange{path: path, content: after, before: before, existed: before != nil})
}
func matchInstallJournal(path string, before []byte) error {
	info, err := os.Lstat(path)
	if errors.Is(err, os.ErrNotExist) && before == nil {
		return nil
	}
	if err != nil {
		return err
	}
	if before == nil || !info.Mode().IsRegular() {
		return errors.New("installer recovery journal changed; preserve it for explicit diagnosis")
	}
	current, err := readEntrypointFile(path, installJournalLimit)
	if err != nil {
		return err
	}
	if !bytes.Equal(current, before) {
		return errors.New("installer recovery journal changed after validation")
	}
	return nil
}
func removeInstallJournal(path string, before []byte) error {
	if err := matchInstallJournal(path, before); err != nil {
		return err
	}
	return os.Remove(path)
}

func installFileState(change installJournalChange) (before, after bool, err error) {
	info, err := os.Lstat(change.Path)
	if errors.Is(err, os.ErrNotExist) {
		return !change.Existed, change.Remove, nil
	}
	if err != nil {
		return false, false, err
	}
	if !info.Mode().IsRegular() {
		return false, false, fmt.Errorf("installer target is no longer a regular owned file: %s", change.Path)
	}
	current, err := readEntrypointFile(change.Path, installJournalLimit)
	if err != nil {
		return false, false, err
	}
	return change.Existed && bytes.Equal(current, change.Before), !change.Remove && bytes.Equal(current, change.After), nil
}
func validateInstallPrefix(changes []installJournalChange) (allAfter bool, err error) {
	allAfter = true
	seenBefore := false
	for _, change := range changes {
		before, after, e := installFileState(change)
		if e != nil {
			return false, e
		}
		if !before && !after {
			return false, fmt.Errorf("installer target contains unrecorded edits; preserved: %s", change.Path)
		}
		if !after {
			allAfter = false
		}
		if before && after {
			continue
		}
		if before {
			seenBefore = true
		} else if seenBefore {
			return false, errors.New("installer targets do not match a recorded owned write prefix")
		}
	}
	return allAfter, nil
}
func restoreInstallPrefix(j installJournal) error {
	changes := j.Changes[:j.RollbackPrefix]
	var failures []error
	for i := len(changes) - 1; i >= 0; i-- {
		change := changes[i]
		before, after, err := installFileState(change)
		if err != nil {
			failures = append(failures, err)
			continue
		}
		if before {
			continue
		}
		if !after {
			failures = append(failures, fmt.Errorf("installer rollback preserved unrecorded edits: %s", change.Path))
			continue
		}
		if err = apply(fileChange{path: change.Path, content: change.Before, remove: !change.Existed}); err != nil {
			failures = append(failures, err)
		}
	}
	for _, change := range j.Changes[j.RollbackPrefix:] {
		before, _, err := installFileState(change)
		if err != nil {
			failures = append(failures, err)
		} else if !before {
			failures = append(failures, fmt.Errorf("installer unattempted target differs from its recorded prior state; preserved: %s", change.Path))
		}
	}
	return errors.Join(failures...)
}

// recoverInstallTransaction runs before old-group ownership planning. Complete
// output is verified; partial output is restored only from recorded owned bytes.
// The next fresh plan then uses ordinary ownership checks again.
func recoverInstallTransaction(ctx context.Context, options InstallOptions, repo, state string) (*InstallResult, error) {
	for _, source := range []string{"codex", "opencode"} {
		if source != options.Source {
			if _, err := os.Lstat(installJournalPath(state, source)); err == nil {
				return nil, errors.New("another source has an unfinished installer transaction; recover that exact setup first")
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, err
			}
		}
	}
	path := installJournalPath(state, options.Source)
	j, raw, err := readInstallJournal(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	identity, err := currentInstallIdentity(ctx, options, repo, state, j.Manifest)
	if err != nil {
		return nil, err
	}
	if err = validateInstallJournal(j, identity); err != nil {
		return nil, err
	}
	if options.DryRun {
		return nil, errors.New("pending owned installer transaction requires explicit non-dry-run recovery; no files changed")
	}
	if j.Phase == "applying" {
		complete, err := validateInstallPrefix(j.Changes)
		if err != nil {
			return nil, err
		}
		if complete {
			if err = removeInstallJournal(path, raw); err != nil {
				return nil, err
			}
			result := InstallResult{Manifest: j.Manifest, Changes: []PlannedChange{}}
			for _, c := range j.Changes {
				action := "write"
				if c.Remove {
					action = "remove"
				}
				result.Changes = append(result.Changes, PlannedChange{c.Path, action})
			}
			return &result, nil
		}
		j.Phase = "rollback"
		j.RollbackPrefix = len(j.Changes)
		if err = saveInstallJournal(path, j, raw); err != nil {
			return nil, err
		}
		raw, err = encodeInstallJournal(j)
		if err != nil {
			return nil, err
		}
	}
	if err = restoreInstallPrefix(j); err != nil {
		return nil, err
	}
	if err = removeInstallJournal(path, raw); err != nil {
		return nil, err
	}
	return nil, nil
}

func applyInstallTransaction(ctx context.Context, options InstallOptions, repo, state string, m Manifest, changes []fileChange) error {
	return applyInstallTransactionWithWriter(ctx, options, repo, state, m, changes, apply)
}

// The imperative file-writer seam permits deterministic I/O interruption tests;
// ordinary installation always supplies the same production atomic file writer.
func applyInstallTransactionWithWriter(ctx context.Context, options InstallOptions, repo, state string, m Manifest, changes []fileChange, write func(fileChange) error) error {
	identity, err := currentInstallIdentity(ctx, options, repo, state, m)
	if err != nil {
		return err
	}
	j := installJournal{Protocol: 1, Phase: "applying", Identity: identity, Manifest: m}
	for _, c := range changes {
		j.Changes = append(j.Changes, installJournalChange{Path: c.path, Before: c.before, After: c.content, Existed: c.existed, Remove: c.remove})
	}
	if err = validateInstallJournal(j, identity); err != nil {
		return err
	}
	for _, c := range j.Changes {
		before, _, e := installFileState(c)
		if e != nil {
			return e
		}
		if !before {
			return errors.New("installer target changed during planning; retry without overwriting newer content")
		}
	}
	path := installJournalPath(state, options.Source)
	if err = saveInstallJournal(path, j, nil); err != nil {
		return err
	}
	raw, err := encodeInstallJournal(j)
	if err != nil {
		return err
	}
	for i, c := range changes {
		if err = ctx.Err(); err == nil {
			before, after, e := installFileState(j.Changes[i])
			if e != nil {
				err = e
			} else if !before && !after {
				err = errors.New("installer target changed after planning; preserved")
			} else if !after {
				err = write(c)
			}
		}
		if err != nil {
			cause := err
			j.Phase = "rollback"
			j.RollbackPrefix = i + 1
			if e := saveInstallJournal(path, j, raw); e != nil {
				return errors.Join(cause, e)
			}
			raw, e := encodeInstallJournal(j)
			if e != nil {
				return errors.Join(cause, e)
			}
			if e = restoreInstallPrefix(j); e != nil {
				return errors.Join(cause, e)
			}
			return errors.Join(cause, removeInstallJournal(path, raw))
		}
	}
	complete, err := validateInstallPrefix(j.Changes)
	if err != nil {
		return err
	}
	if !complete {
		return errors.New("installer outputs changed before completion; recovery journal retained")
	}
	return removeInstallJournal(path, raw)
}
