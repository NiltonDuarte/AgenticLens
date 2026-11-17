package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

const publicationLimit = 128 << 20

var ErrPublicationPending = errors.New("incomplete artifact publication")

type priorArtifact struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
	Data    []byte `json:"data"`
}

// validatePublicationObject requires exact, unique field names. JSON's normal
// case-insensitive and last-key-wins struct decoding is too permissive here.
func validatePublicationObject(body []byte, fields []string) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != json.Delim('{') {
		return errors.New("publication metadata must be an object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("invalid publication metadata key")
		}
		allowed := false
		for _, field := range fields {
			if field == key {
				allowed = true
			}
		}
		if !allowed || seen[key] {
			return fmt.Errorf("unknown or duplicate publication metadata field %q", key)
		}
		seen[key] = true
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if len(seen) != len(fields) {
		return errors.New("incomplete publication metadata")
	}
	return nil
}

// UnmarshalJSON rejects omitted presence/data fields rather than interpreting
// malformed metadata as authority to remove an artifact.
func (a *priorArtifact) UnmarshalJSON(body []byte) error {
	if err := validatePublicationObject(body, []string{"path", "present", "data"}); err != nil {
		return err
	}
	type artifactWire struct {
		Path    string          `json:"path"`
		Present *bool           `json:"present"`
		Data    json.RawMessage `json:"data"`
	}
	var wire artifactWire
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	if wire.Path == "" || wire.Present == nil || len(wire.Data) == 0 {
		return errors.New("incomplete publication artifact metadata")
	}
	var data []byte
	if err := json.Unmarshal(wire.Data, &data); err != nil {
		return err
	}
	*a = priorArtifact{Path: wire.Path, Present: *wire.Present, Data: data}
	return nil
}

type publicationJournal struct {
	Version     int             `json:"version"`
	Ledger      string          `json:"ledger"`
	Directory   string          `json:"directory"`
	Project     string          `json:"project"`
	MemberCount int             `json:"member_count"`
	Artifacts   []priorArtifact `json:"artifacts"`
}
type pendingPublication struct {
	Version     int    `json:"version"`
	Transaction string `json:"transaction"`
}

func (j *publicationJournal) UnmarshalJSON(body []byte) error {
	if err := validatePublicationObject(body, []string{"version", "ledger", "directory", "project", "member_count", "artifacts"}); err != nil {
		return err
	}
	type wire publicationJournal
	var value wire
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&value); err != nil {
		return err
	}
	*j = publicationJournal(value)
	return nil
}
func canonicalPublicationPath(path string) (string, error) {
	if capture.PublicationOwner(path) != nil {
		return capture.PublicationPath(path), nil
	}
	return capture.ProjectDirectory(path)
}
func journalPath(ledger string) string {
	return capture.PublicationPath(ledger) + ".publication-pending"
}
func CheckPublication(ledger string) error {
	if _, err := capture.LstatPublicationFile(journalPath(ledger)); err == nil {
		return fmt.Errorf("%w at %s; recovery requires the originating operational state home", ErrPublicationPending, ledger)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}
func publicationDestinations(ledger string, paths []string) (string, []string, error) {
	owner := capture.PublicationOwner(ledger)
	if owner == nil {
		return "", nil, fmt.Errorf("publication requires ownership of the resolved ledger directory")
	}
	if err := owner.ValidateLocation(); err != nil {
		return "", nil, err
	}
	binding := filepath.Base(ledger)
	targets := []string{}
	physicalTargets := []string{}
	infos := []os.FileInfo{}
	hasLedger := false
	for _, path := range paths {
		if path == "" {
			continue
		}
		bound := capture.PublicationPath(path)
		rel, err := filepath.Rel(owner.DirectoryPath(), bound)
		if err != nil || rel == "." || rel == ".." || filepath.IsAbs(rel) || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
			return "", nil, fmt.Errorf("publication member is outside selected ledger directory: %s", path)
		}
		// Existing ancestors must not escape the held directory through a symlink.
		actual, err := owner.PhysicalMemberPath(bound)
		if err != nil {
			return "", nil, err
		}
		home, err := owner.PhysicalMemberPath(owner.DirectoryPath())
		if err != nil {
			return "", nil, err
		}
		if !insideWorkspace(home, actual) {
			return "", nil, fmt.Errorf("publication member escapes selected directory")
		}
		info, err := capture.StatPublicationFile(bound)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return "", nil, err
		}
		if info != nil && !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("publication member must be regular")
		}
		for i, prior := range targets {
			if actual == physicalTargets[i] || rel == prior || strings.HasPrefix(rel, prior+string(filepath.Separator)) || strings.HasPrefix(prior, rel+string(filepath.Separator)) || (info != nil && infos[i] != nil && os.SameFile(info, infos[i])) {
				return "", nil, fmt.Errorf("overlapping publication members")
			}
		}
		if rel == binding {
			hasLedger = true
		}
		markerActual, e := owner.PhysicalMemberPath(journalPath(ledger))
		if e != nil {
			return "", nil, e
		}
		if actual == markerActual || rel == binding+".publication-pending" {
			return "", nil, fmt.Errorf("artifact overlaps pending indicator")
		}
		targets = append(targets, rel)
		physicalTargets = append(physicalTargets, actual)
		infos = append(infos, info)
	}
	if !hasLedger || len(targets) < 1 || len(targets) > 3 {
		return "", nil, fmt.Errorf("publication requires ledger and one to three authorized members")
	}
	return binding, targets, nil
}
func syncDirectory(path string) error {
	f, err := capture.OpenPublicationFile(path, os.O_RDONLY, 0)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
func durableReplace(path string, data []byte) error {
	path = capture.PublicationPath(path)
	f, err := capture.CreatePublicationTemp(filepath.Dir(path), ".publication-stage-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() {
		if e := capture.RemovePublicationFile(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "publication staging cleanup:", e)
		}
	}()
	if err = f.Chmod(0600); err == nil {
		_, err = f.Write(data)
	}
	if err = errors.Join(err, f.Sync(), f.Close()); err != nil {
		return err
	}
	if err = capture.RenamePublicationFile(name, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
func privateJournalPath(owner *capture.PublicationLock, transaction string) string {
	digest := sha256.Sum256([]byte(owner.Project()))
	return filepath.Join(owner.StateHome(), "publication", "journals", hex.EncodeToString(digest[:]), transaction+".json")
}
func loadPending(ledger string) (pendingPublication, bool, error) {
	var pending pendingPublication
	file, err := capture.OpenPublicationFile(journalPath(ledger), os.O_RDONLY, 0)
	if errors.Is(err, os.ErrNotExist) {
		return pending, false, nil
	}
	if err != nil {
		return pending, true, err
	}
	body, readErr := io.ReadAll(io.LimitReader(file, 4097))
	err = errors.Join(readErr, file.Close())
	if err != nil {
		return pending, true, err
	}
	if len(body) > 4096 {
		return pending, true, fmt.Errorf("pending-publication indicator exceeds bound")
	}
	if err = validatePublicationObject(body, []string{"version", "transaction"}); err != nil {
		return pending, true, fmt.Errorf("unsupported/retired publication state; recover with matching tool before upgrade: %w", err)
	}
	if err = json.Unmarshal(body, &pending); err != nil {
		return pending, true, err
	}
	token, err := hex.DecodeString(pending.Transaction)
	if err != nil || len(token) != 16 || pending.Version != 2 {
		return pending, true, fmt.Errorf("unsupported pending publication format; explicit recovery required")
	}
	return pending, true, nil
}
func BeginPublication(ledger string, paths []string) error {
	if err := CheckPublication(ledger); err != nil {
		return err
	}
	binding, targets, err := publicationDestinations(ledger, paths)
	if err != nil {
		return err
	}
	owner := capture.PublicationOwner(ledger)
	j := publicationJournal{Version: 2, Ledger: binding, Directory: owner.DirectoryIdentity(), Project: owner.Project(), MemberCount: len(targets), Artifacts: []priorArtifact{}}
	total := 0
	for _, rel := range targets {
		member := priorArtifact{Path: rel}
		f, err := capture.OpenPublicationFile(filepath.Join(owner.DirectoryPath(), rel), os.O_RDONLY, 0)
		if err == nil {
			member.Present = true
			member.Data, err = io.ReadAll(io.LimitReader(f, int64(publicationLimit-total+1)))
			err = errors.Join(err, f.Close())
		} else if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		if err != nil {
			return err
		}
		total += len(member.Data)
		if total > publicationLimit {
			return fmt.Errorf("publication backup exceeds 128 MiB recovery bound")
		}
		j.Artifacts = append(j.Artifacts, member)
	}
	token := make([]byte, 16)
	if _, err = rand.Read(token); err != nil {
		return err
	}
	pending := pendingPublication{Version: 2, Transaction: hex.EncodeToString(token)}
	private := privateJournalPath(owner, pending.Transaction)
	if err = capture.MkdirPublicationStateDirectories(filepath.Dir(private), 0700); err != nil {
		return err
	}
	body, err := json.Marshal(j)
	if err != nil {
		return err
	}
	if err = durableReplace(private, body); err != nil {
		return err
	}
	indicator, err := json.Marshal(pending)
	if err != nil {
		return err
	}
	return durableReplace(journalPath(ledger), indicator)
}
func loadRecovery(ledger string, paths []string) (publicationJournal, pendingPublication, bool, error) {
	pending, present, err := loadPending(ledger)
	var j publicationJournal
	if err != nil || !present {
		return j, pending, present, err
	}
	owner := capture.PublicationOwner(ledger)
	if owner == nil {
		return j, pending, true, fmt.Errorf("recovery requires publication ownership and the originating operational state home")
	}
	binding, targets, err := publicationDestinations(ledger, paths)
	if err != nil {
		return j, pending, true, err
	}
	f, err := capture.OpenPublicationFile(privateJournalPath(owner, pending.Transaction), os.O_RDONLY, 0)
	if err != nil {
		return j, pending, true, fmt.Errorf("private publication recovery unavailable; pending indicator/artifacts retained: %w", err)
	}
	body, readErr := io.ReadAll(io.LimitReader(f, 2*publicationLimit+1))
	err = errors.Join(readErr, f.Close())
	if err != nil {
		return j, pending, true, err
	}
	if len(body) > 2*publicationLimit {
		return j, pending, true, fmt.Errorf("private journal exceeds bound")
	}
	if err = json.Unmarshal(body, &j); err != nil {
		return j, pending, true, err
	}
	if j.Version != 2 || j.Ledger != binding || j.Directory != owner.DirectoryIdentity() || j.Project != owner.Project() || j.MemberCount != len(targets) || len(j.Artifacts) != len(targets) {
		return j, pending, true, fmt.Errorf("private journal does not match the complete authorized publication set; recovery incomplete")
	}
	seen := map[string]bool{}
	total := 0
	for _, a := range j.Artifacts {
		allowed := false
		for _, target := range targets {
			if a.Path == target {
				allowed = true
			}
		}
		if !allowed || seen[a.Path] || (!a.Present && len(a.Data) > 0) {
			return j, pending, true, fmt.Errorf("invalid or conflicting authorized recovery member")
		}
		seen[a.Path] = true
		total += len(a.Data)
	}
	if total > publicationLimit {
		return j, pending, true, fmt.Errorf("private recovery data exceeds bound")
	}
	return j, pending, true, nil
}
func RecoverPublication(ledger string, authorizedPaths []string) (bool, error) {
	j, pending, present, err := loadRecovery(ledger, authorizedPaths)
	if err != nil || !present {
		return present, err
	}
	owner := capture.PublicationOwner(ledger)
	for _, a := range j.Artifacts {
		path := filepath.Join(owner.DirectoryPath(), a.Path)
		if a.Present {
			if err = durableReplace(path, a.Data); err != nil {
				return true, err
			}
		} else {
			if err = capture.RemovePublicationFile(path); err != nil && !errors.Is(err, os.ErrNotExist) {
				return true, err
			}
			if err = syncDirectory(filepath.Dir(path)); err != nil && !errors.Is(err, os.ErrNotExist) {
				return true, err
			}
		}
	}
	return true, finishPublication(ledger, pending, owner)
}
func finishPublication(ledger string, pending pendingPublication, owner *capture.PublicationLock) error {
	if err := capture.RemovePublicationFile(journalPath(ledger)); err != nil {
		return err
	}
	if err := syncDirectory(owner.DirectoryPath()); err != nil {
		return err
	}
	path := privateJournalPath(owner, pending.Transaction)
	if err := capture.RemovePublicationFile(path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}
func CommitPublication(ledger string, paths []string) error {
	_, pending, present, err := loadRecovery(ledger, paths)
	if err != nil {
		return err
	}
	if !present {
		return fmt.Errorf("no pending publication to commit")
	}
	owner := capture.PublicationOwner(ledger)
	for _, path := range paths {
		f, err := capture.OpenPublicationFile(path, os.O_RDONLY, 0)
		if err != nil {
			return err
		}
		if err = errors.Join(f.Sync(), f.Close()); err != nil {
			return err
		}
		// Nested member directories must be durable before acknowledging the
		// complete set, including every newly created ancestor below the root.
		for parent := filepath.Dir(capture.PublicationPath(path)); parent != owner.DirectoryPath(); parent = filepath.Dir(parent) {
			if err = syncDirectory(parent); err != nil {
				return err
			}
		}
	}
	if err = syncDirectory(owner.DirectoryPath()); err != nil {
		return err
	}
	return finishPublication(ledger, pending, owner)
}
func readPublishedLedger(path string) ([]byte, error) {
	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	directory, err := capture.ReadPublicationDirectory(path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if err := directory.Close(); err != nil {
			fmt.Fprintln(os.Stderr, "publication read close:", err)
		}
	}()
	return readDirectoryLedger(directory, path)
}

func readDirectoryLedger(directory *capture.ReadDirectory, path string) ([]byte, error) {
	pending := filepath.Base(path) + ".publication-pending"
	check := func() error {
		if _, err := directory.Lstat(pending); err == nil {
			return fmt.Errorf("%w at %s; originating operational state is required for recovery", ErrPublicationPending, path)
		} else if !errors.Is(err, os.ErrNotExist) {
			return err
		}
		return nil
	}
	if err := check(); err != nil {
		return nil, err
	}
	body, err := directory.ReadFile(filepath.Base(path))
	var pathError *os.PathError
	if errors.As(err, &pathError) {
		return nil, fmt.Errorf("ledger %s unavailable: %w", path, pathError.Err)
	}
	if err != nil {
		return nil, err
	}
	if err := check(); err != nil {
		return nil, err
	}
	return body, nil
}

func readOwnedLedger(path string) ([]byte, error) {
	path = capture.PublicationPath(path)
	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	f, err := capture.OpenPublicationFile(path, os.O_RDONLY, 0)
	if err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(io.LimitReader(f, publicationLimit+1))
	err = errors.Join(readErr, f.Close())
	if err != nil {
		return nil, err
	}
	if len(body) > publicationLimit {
		return nil, fmt.Errorf("ledger exceeds 128 MiB read bound")
	}
	if err = CheckPublication(path); err != nil {
		return nil, err
	}
	return body, nil
}

// LoadObservedFactLedger performs a bounded advisory observation without taking
// publication ownership or creating lock files. It reuses canonical ledger
// validation and detects a journal or replacement racing the read. Callers must
// report unavailable observations; this is not authority to refresh or publish.
func LoadObservedFactLedger(path string) (*ExistingLedgerFacts, error) {
	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	before, err := os.Stat(path)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > publicationLimit {
		return nil, errors.New("metrics publication is not a bounded regular file")
	}
	facts, err := loadExistingFactLedger(path, true)
	if err != nil {
		return nil, err
	}
	after, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !os.SameFile(before, after) || before.Size() != after.Size() || !before.ModTime().Equal(after.ModTime()) {
		return nil, errors.New("metrics publication changed during read-only inspection; retry capture check")
	}
	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	return facts, nil
}
