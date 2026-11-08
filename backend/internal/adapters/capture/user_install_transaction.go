package capture

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
)

const userInstallJournalVersion = 2
const userInstallPendingFile = "agenticlens-user-install.pending.json"

var errNoPendingUserJournal = fmt.Errorf("no matching pending user journal: %w", os.ErrNotExist)

type userInstallRecordKind string

const (
	userInstallJournalKind   userInstallRecordKind = "journal"
	userInstallReferenceKind userInstallRecordKind = "reference"
)

type userInstallChange struct {
	Path       string
	Content    []byte
	Before     []byte
	Remove     bool
	Existed    bool
	Executable bool
}

type userInstallJournal struct {
	Version          int                   `json:"version"`
	Kind             userInstallRecordKind `json:"kind"`
	OwnerHome        string                `json:"owner_home"`
	Coordinator      string                `json:"coordinator"`
	LegacyRepository string                `json:"legacy_repository,omitempty"`
	Homes            []string              `json:"homes,omitempty"`
	Changes          []userInstallChange   `json:"changes"`
}

type userInstallReference struct {
	Version         int                   `json:"version"`
	Kind            userInstallRecordKind `json:"kind"`
	OwnerHome       string                `json:"owner_home"`
	Coordinator     string                `json:"coordinator"`
	JournalChecksum string                `json:"journal_checksum"`
	Checksum        string                `json:"checksum"`
}

type userPendingRecord struct {
	Kind      userInstallRecordKind
	Journal   userInstallJournal
	Reference userInstallReference
	Checksum  string
}

type userInstallStage uint8

const (
	userInstallJournaled userInstallStage = iota
	userInstallApplied
	userInstallBeforeCommit
	userInstallResourcesDiscovered
	userInstallReferenced
	userInstallCommitted
)

func readPendingUserRecord(directory string) (userPendingRecord, bool, error) {
	var record userPendingRecord
	data, err := readEntrypointFile(filepath.Join(directory, userInstallPendingFile), installJournalLimit)
	if errors.Is(err, os.ErrNotExist) {
		return record, false, nil
	}
	if err != nil {
		return record, false, err
	}
	var header struct {
		Version int                   `json:"version"`
		Kind    userInstallRecordKind `json:"kind"`
	}
	if err := json.Unmarshal(data, &header); err != nil {
		return record, false, err
	}
	if header.Version != userInstallJournalVersion {
		return record, false, errors.New("unsupported pending user installation; preserved")
	}
	record.Kind = header.Kind
	record.Checksum = hash(data)
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	switch header.Kind {
	case userInstallJournalKind:
		if err := decoder.Decode(&record.Journal); err != nil {
			return record, false, err
		}
		if !filepath.IsAbs(record.Journal.OwnerHome) || record.Journal.OwnerHome != directory || !filepath.IsAbs(record.Journal.Coordinator) {
			return record, false, errors.New("pending user journal owner must be absolute; preserved")
		}
	case userInstallReferenceKind:
		if err := decoder.Decode(&record.Reference); err != nil {
			return record, false, err
		}
		if !filepath.IsAbs(record.Reference.OwnerHome) || !filepath.IsAbs(record.Reference.Coordinator) || record.Reference.Coordinator != directory || record.Reference.OwnerHome == directory || len(record.Reference.JournalChecksum) != 64 || record.Reference != userReference(record.Reference.OwnerHome, record.Reference.Coordinator, record.Reference.JournalChecksum) {
			return record, false, errors.New("pending user reference ownership invalid; preserved")
		}
	default:
		return record, false, errors.New("unknown pending user installation record; preserved")
	}
	return record, true, nil
}

// referencedUserJournal follows a small shared reference to private undo data.
// An absent journal is cleanup-only while the owner directory still exists;
// deleting the owner directory is an error, not evidence of a completed install.
func referencedUserJournal(reference userInstallReference) (userInstallJournal, bool, error) {
	var empty userInstallJournal
	info, err := os.Stat(reference.OwnerHome)
	if err != nil {
		return empty, false, fmt.Errorf("pending migration owner unavailable; reference preserved: %w", err)
	}
	if !info.IsDir() {
		return empty, false, errors.New("pending migration owner is not a directory; reference preserved")
	}
	record, exists, err := readPendingUserRecord(reference.OwnerHome)
	if err != nil || !exists {
		return empty, false, err
	}
	if record.Kind != userInstallJournalKind {
		return empty, false, errors.New("pending migration owner is not a journal; reference preserved")
	}
	if record.Journal.OwnerHome != reference.OwnerHome || record.Journal.Coordinator != reference.Coordinator {
		return empty, false, nil
	}
	if record.Checksum != reference.JournalChecksum {
		return empty, false, errors.New("pending migration journal changed; shared reference preserved")
	}
	return record.Journal, true, nil
}

// readUserInstallJournal resolves either record to its private owner's journal.
func readUserInstallJournal(directory string) (userInstallJournal, string, error) {
	record, exists, err := readPendingUserRecord(directory)
	if err != nil {
		return userInstallJournal{}, "", err
	}
	if !exists {
		return userInstallJournal{}, "", errNoPendingUserJournal
	}
	if record.Kind == userInstallJournalKind {
		return record.Journal, record.Journal.OwnerHome, nil
	}
	journal, exists, err := referencedUserJournal(record.Reference)
	if err != nil {
		return journal, "", err
	}
	if !exists {
		return journal, "", errNoPendingUserJournal
	}
	return journal, journal.OwnerHome, nil
}

// pendingUserHomes expands only a matching private transaction. Stale references
// never enroll another migration's private undo journal.

func pendingUserHomes(directory string) ([]string, error) {
	// Validate repository references before they can enroll private lease paths.
	if err := validateUserRecovery(directory); err != nil {
		return nil, err
	}
	record, exists, err := readPendingUserRecord(directory)
	if err != nil || !exists {
		return nil, err
	}
	if record.Kind == userInstallJournalKind {
		return currentUserJournalHomes(record.Journal)
	}
	journal, exists, err := referencedUserJournal(record.Reference)
	if err != nil || !exists {
		return nil, err
	}
	return currentUserJournalHomes(journal)
}

func currentUserJournalHomes(journal userInstallJournal) ([]string, error) {
	homes := userJournalHomes(journal)
	if journal.LegacyRepository != "" {
		current, err := legacyUserHomes(journal.LegacyRepository)
		if err != nil {
			return nil, err
		}
		homes = append(homes, current...)
	}
	return homes, nil
}

func userJournalHomes(journal userInstallJournal) []string {
	homes := append([]string(nil), journal.Homes...)
	if journal.OwnerHome != "" {
		homes = append(homes, journal.OwnerHome)
	}
	for _, change := range journal.Changes {
		switch filepath.Base(change.Path) {
		case "hooks.json", "agenticlens-user-installation.json", "agenticlens-installation.json", "install-codex.json":
			homes = append(homes, filepath.Dir(change.Path))
		}
	}
	return homes
}

func userReference(owner, coordinator, journalChecksum string) userInstallReference {
	return userInstallReference{Version: userInstallJournalVersion, Kind: userInstallReferenceKind, OwnerHome: owner, Coordinator: coordinator, JournalChecksum: journalChecksum, Checksum: hash([]byte(owner + "\x00" + coordinator + "\x00" + journalChecksum))}
}

func removeUserReference(reference userInstallReference) error {
	record, exists, err := readPendingUserRecord(reference.Coordinator)
	if err != nil || !exists {
		return err
	}
	if record.Kind != userInstallReferenceKind || record.Reference != reference {
		return errors.New("pending user reference changed; preserved")
	}
	return os.Remove(filepath.Join(reference.Coordinator, userInstallPendingFile))
}

// Repository references never authorize arbitrary private-file rewrites. Recovery
// uses a matching private journal, its exact checksum and known owned targets.
func validateUserRecovery(directory string) error {
	journal, _, err := readUserInstallJournal(directory)
	if errors.Is(err, errNoPendingUserJournal) {
		return nil
	}
	if err != nil {
		return err
	}
	if journal.OwnerHome != journal.Coordinator {
		owner, exists, err := readPendingUserRecord(journal.OwnerHome)
		if err != nil {
			return err
		}
		if !exists || owner.Kind != userInstallJournalKind {
			return errors.New("pending private user journal changed; preserved")
		}
		shared, exists, err := readPendingUserRecord(journal.Coordinator)
		if err != nil {
			return err
		}
		expected := userReference(journal.OwnerHome, journal.Coordinator, owner.Checksum)
		if !exists || shared.Kind != userInstallReferenceKind || shared.Reference != expected {
			return errors.New("pending migration shared reference changed or missing; private journal preserved")
		}
	}
	return validateUserJournalScope(journal, journal.Coordinator)
}

func validateUserJournalScope(journal userInstallJournal, coordinator string) error {
	allowed := map[string]bool{}
	resources := map[string]bool{journal.OwnerHome: true}
	for _, name := range []string{"hooks.json", "agenticlens-user-installation.json"} {
		allowed[filepath.Join(journal.OwnerHome, name)] = true
	}
	if journal.LegacyRepository == "" {
		if coordinator != journal.OwnerHome {
			return errors.New("pending migration lacks legacy ownership; preserved")
		}
	} else {
		legacyResources, err := legacyUserHomes(journal.LegacyRepository)
		if err != nil {
			return fmt.Errorf("pending legacy migration ownership unavailable; preserved: %w", err)
		}
		legacyHome, err := absolute(legacyResources[0])
		if err != nil {
			return err
		}
		if legacyHome != coordinator {
			return errors.New("pending migration coordinator differs from legacy ownership; preserved")
		}
		stateHome, err := absolute(legacyResources[1])
		if err != nil {
			return err
		}
		state, err := ReadManifest(stateHome, "codex")
		if err != nil {
			return fmt.Errorf("pending legacy state ownership unavailable; preserved: %w", err)
		}
		if state.Protocol != ManifestProtocolVersion || state.Source != "codex" || state.Repository != journal.LegacyRepository || state.ConfigPath != filepath.Join(journal.LegacyRepository, ".codex", "hooks.json") {
			return errors.New("pending legacy state ownership differs; preserved")
		}
		allowed[filepath.Join(legacyHome, "hooks.json")] = true
		allowed[filepath.Join(legacyHome, "agenticlens-installation.json")] = true
		allowed[manifestPath(stateHome, "codex")] = true
		resources[legacyHome] = true
		resources[stateHome] = true
	}
	for _, home := range journal.Homes {
		physical, err := absolute(home)
		if err != nil {
			return err
		}
		if !resources[physical] {
			return fmt.Errorf("pending user resource %s is outside verified ownership; preserved", home)
		}
	}
	seen := map[string]bool{}
	for _, change := range journal.Changes {
		if !filepath.IsAbs(change.Path) {
			return errors.New("pending user change target must be absolute; preserved")
		}
		parent, err := absolute(filepath.Dir(change.Path))
		if err != nil {
			return err
		}
		target := filepath.Join(parent, filepath.Base(change.Path))
		if !allowed[target] || seen[target] {
			return fmt.Errorf("pending user change target %s is outside unique verified ownership; preserved", change.Path)
		}
		seen[target] = true
	}
	return nil
}

// Recovery is called only after the complete fresh resource set is leased. The
// canonical journal is the sole rollback authority; references never carry bytes.
func recoverUserInstallation(directory string) error {
	if err := validateUserRecovery(directory); err != nil {
		return err
	}
	record, exists, err := readPendingUserRecord(directory)
	if err != nil || !exists {
		return err
	}
	journal := record.Journal
	if record.Kind == userInstallReferenceKind {
		journal, exists, err = referencedUserJournal(record.Reference)
		if err != nil {
			return err
		}
		if !exists {
			return removeUserReference(record.Reference)
		}
	}
	var reference userInstallReference
	if journal.OwnerHome != journal.Coordinator {
		owner, exists, err := readPendingUserRecord(journal.OwnerHome)
		if err != nil {
			return err
		}
		if !exists || owner.Kind != userInstallJournalKind {
			return errors.New("pending private user journal changed; preserved")
		}
		reference = userReference(journal.OwnerHome, journal.Coordinator, owner.Checksum)
	}
	if err := rollbackUserJournal(journal); err != nil {
		return fmt.Errorf("user installation recovery incomplete; prior bytes retained: %w", err)
	}
	if err := os.Remove(filepath.Join(journal.OwnerHome, userInstallPendingFile)); err != nil {
		return err
	}
	if reference.OwnerHome != "" {
		return removeUserReference(reference)
	}
	return nil
}

func rollbackUserJournal(journal userInstallJournal) error {
	for i := len(journal.Changes) - 1; i >= 0; i-- {
		change := journal.Changes[i]
		current, err := os.ReadFile(change.Path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
		if (err == nil) == change.Existed && bytes.Equal(current, change.Before) {
			continue
		}
		if change.Remove {
			if !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("transaction target edited; preserved: %s", change.Path)
			}
		} else if err != nil || !bytes.Equal(current, change.Content) {
			return fmt.Errorf("transaction target edited; preserved: %s", change.Path)
		}
		if err := apply(fileChange{path: change.Path, content: change.Before, remove: !change.Existed}); err != nil {
			return err
		}
		if change.Existed && change.Executable {
			if err := os.Chmod(change.Path, 0755); err != nil {
				return err
			}
		}
	}
	return nil
}

func applyUserInstallationWithCheckpoint(home, coordinator, legacyRepository string, changes []fileChange, checkpoint func(userInstallStage, int) error) error {
	if legacyRepository != "" {
		var err error
		legacyRepository, err = absolute(legacyRepository)
		if err != nil {
			return err
		}
	}
	journal := userInstallJournal{Version: userInstallJournalVersion, Kind: userInstallJournalKind, OwnerHome: home, Coordinator: coordinator, LegacyRepository: legacyRepository, Homes: []string{home}}
	if legacyRepository != "" {
		resources, err := legacyUserHomes(legacyRepository)
		if err != nil {
			return err
		}
		journal.Homes = append(journal.Homes, resources...)
	}
	for _, change := range changes {
		journal.Changes = append(journal.Changes, userInstallChange{Path: change.path, Content: change.content, Before: change.before, Remove: change.remove, Existed: change.existed})
	}
	if err := validateUserJournalScope(journal, coordinator); err != nil {
		return err
	}
	body, err := json.Marshal(journal)
	if err != nil {
		return err
	}
	if int64(len(body)) > installJournalLimit {
		return errors.New("pending user journal exceeds size limit; installation preserved")
	}
	for _, directory := range []string{home, coordinator} {
		if _, exists, err := readPendingUserRecord(directory); err != nil {
			return err
		} else if exists {
			return errors.New("pending user installation remains; recover before applying changes")
		}
	}
	var reference userInstallReference
	if home != coordinator {
		reference = userReference(home, coordinator, hash(body))
		encoded, err := json.Marshal(reference)
		if err != nil {
			return err
		}
		if err := apply(fileChange{path: filepath.Join(coordinator, userInstallPendingFile), content: encoded}); err != nil {
			return err
		}
		if checkpoint != nil {
			if err := checkpoint(userInstallReferenced, -1); err != nil {
				return err
			}
		}
	}
	path := filepath.Join(home, userInstallPendingFile)
	if err := apply(fileChange{path: path, content: body}); err != nil {
		if reference.OwnerHome != "" {
			return errors.Join(err, removeUserReference(reference))
		}
		return err
	}
	if checkpoint != nil {
		if err := checkpoint(userInstallJournaled, -1); err != nil {
			return err
		}
	}
	for index, change := range changes {
		if applyErr := apply(change); applyErr != nil {
			return errors.Join(applyErr, recoverUserInstallation(home))
		}
		if checkpoint != nil {
			if err := checkpoint(userInstallApplied, index); err != nil {
				return err
			}
		}
	}
	if checkpoint != nil {
		if err := checkpoint(userInstallBeforeCommit, len(changes)); err != nil {
			return err
		}
	}
	// Removing the private journal commits. The shared reference is cleanup-only
	// metadata and contains no private hook configuration or undo images.
	if err := os.Remove(path); err != nil {
		return err
	}
	if checkpoint != nil {
		if err := checkpoint(userInstallCommitted, len(changes)); err != nil {
			return err
		}
	}
	if reference.OwnerHome != "" {
		return removeUserReference(reference)
	}
	return nil
}
