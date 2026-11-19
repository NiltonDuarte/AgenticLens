package capture

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"syscall"

	"agenticlens/internal/adapters/config"
)

const initialPublicationLimit = 16 << 10

// InitialPublication is private authorization to publish once to a selected
// physical directory. The ledger remains the authority for ticket identity.
type InitialPublication struct {
	Version                int    `json:"version"`
	Ticket                 string `json:"ticket"`
	ID                     string `json:"id"`
	Path                   string `json:"path"`
	Directory              string `json:"directory"`
	CommittedTransaction   string `json:"committed_transaction"`
	CommittedArtifacts     string `json:"committed_artifacts"`
	ExistingLedgerObserved bool   `json:"existing_ledger_observed,omitempty"`
}

// InitialPublicationLink binds a transaction journal to the exact private
// authorization from which it began.
type InitialPublicationLink struct {
	Ticket string `json:"ticket"`
	ID     string `json:"id"`
}

func (selection *InitialPublication) Closed() bool {
	return selection.CommittedTransaction != "" || selection.ExistingLedgerObserved
}

func (selection *InitialPublication) Link() InitialPublicationLink {
	return InitialPublicationLink{Ticket: selection.Ticket, ID: selection.ID}
}

// InitialPublicationLease serializes first destination choices and first
// writes for one ticket, including writers targeting different directories.
type InitialPublicationLease struct {
	file       *os.File
	project    string
	state      string
	ticket     string
	protection *ProtectedPublication
}

func AcquireInitialPublication(ctx context.Context, project, ticket string, protection *ProtectedPublication) (*InitialPublicationLease, error) {
	if ticket == "" {
		return nil, errors.New("first-publication coordination requires the exact ticket ID")
	}
	project, err := config.ProjectRoot(project)
	if err != nil {
		return nil, err
	}
	state, err := publicationState(project, protection)
	if err != nil {
		return nil, err
	}
	record := initialPublicationPath(state, project, ticket)
	lock := record + ".lock"
	if err := protection.checkPrivatePath(lock); err != nil {
		return nil, err
	}
	if err := MkdirPublicationStateDirectories(filepath.Dir(lock), 0700); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(lock, os.O_CREATE|os.O_RDWR|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0600)
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Join(errors.New("first-publication lock must be a regular file"), file.Close())
	}
	if err := waitPublicationLock(ctx, file); err != nil {
		return nil, errors.Join(err, file.Close())
	}
	return &InitialPublicationLease{file: file, project: project, state: state, ticket: ticket, protection: protection}, nil
}

func (lease *InitialPublicationLease) Close() error {
	return lease.file.Close()
}

func (lease *InitialPublicationLease) Read() (*InitialPublication, error) {
	return readInitialPublication(lease.state, lease.project, lease.ticket, lease.protection)
}

func (lease *InitialPublicationLease) Prepare(owner *PublicationLock) error {
	if lease.state != owner.state || lease.project != owner.project {
		return errors.New("first-publication state changed during selection; retry")
	}
	if err := owner.ValidateLocation(); err != nil {
		return err
	}
	prior, err := lease.Read()
	if err != nil {
		return err
	}
	path := filepath.Join(owner.original, "agentic_metrics.json")
	if prior != nil {
		if prior.Closed() {
			return errors.New("the initial selection for this ticket is no longer available")
		}
		if prior.Path != path || prior.Directory != owner.identity {
			return errors.New("a different first-publication destination is already selected for this ticket")
		}
		return nil
	}
	token := make([]byte, 16)
	if _, err := rand.Read(token); err != nil {
		return err
	}
	record := InitialPublication{Version: 1, Ticket: lease.ticket, ID: hex.EncodeToString(token), Path: path, Directory: owner.identity}
	return saveInitialPublication(owner, record)
}

// ObserveExisting permanently closes a staged selection when the lifecycle
// worker has admitted a committed ledger for the same ticket.
func (lease *InitialPublicationLease) ObserveExisting(owner *PublicationLock) error {
	if lease.state != owner.state || lease.project != owner.project {
		return errors.New("first-publication state changed during existing ledger admission; retry")
	}
	record, err := lease.Read()
	if err != nil || record == nil || record.CommittedTransaction != "" {
		return err
	}
	if record.ExistingLedgerObserved {
		return syncInitialPublicationState(owner, record.Ticket)
	}
	record.ExistingLedgerObserved = true
	return saveInitialPublication(owner, *record)
}

func initialPublicationPath(state, project, ticket string) string {
	scope := sha256.Sum256([]byte(project + "\x00" + ticket))
	return filepath.Join(state, "publication", "initial", hex.EncodeToString(scope[:])+".json")
}

func readInitialPublication(state, project, ticket string, protection *ProtectedPublication) (*InitialPublication, error) {
	path := initialPublicationPath(state, project, ticket)
	if err := protection.checkPrivatePath(path); err != nil {
		return nil, err
	}
	file, err := os.OpenFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	info, err := file.Stat()
	if err != nil {
		return nil, errors.Join(err, file.Close())
	}
	if !info.Mode().IsRegular() {
		return nil, errors.Join(errors.New("first-publication selection must be a regular file"), file.Close())
	}
	data, readErr := io.ReadAll(io.LimitReader(file, initialPublicationLimit+1))
	if err := errors.Join(readErr, file.Close()); err != nil {
		return nil, err
	}
	if len(data) > initialPublicationLimit {
		return nil, errors.New("first-publication selection exceeds size limit")
	}
	var record InitialPublication
	if err := json.Unmarshal(data, &record); err != nil {
		return nil, fmt.Errorf("invalid first-publication selection: %w", err)
	}
	if record.Version != 1 || record.Ticket != ticket || !validInitialPublicationToken(record.ID) || record.Directory == "" || !filepath.IsAbs(record.Path) || filepath.Base(record.Path) != "agentic_metrics.json" ||
		!validInitialPublicationCommit(record) {
		return nil, errors.New("first-publication selection does not match its authorized scope")
	}
	return &record, nil
}

func validInitialPublicationToken(value string) bool {
	body, err := hex.DecodeString(value)
	return err == nil && len(body) == 16
}

func validInitialPublicationCommit(record InitialPublication) bool {
	if record.CommittedTransaction == "" {
		return record.CommittedArtifacts == ""
	}
	digest, err := hex.DecodeString(record.CommittedArtifacts)
	return !record.ExistingLedgerObserved && validInitialPublicationToken(record.CommittedTransaction) && err == nil && len(digest) == sha256.Size
}

// InitialPublicationForTransaction verifies a journal's authorization under
// the directory lease. A consumed authorization can only finish its own set.
func (owner *PublicationLock) InitialPublicationForTransaction(link InitialPublicationLink, transaction string) (*InitialPublication, error) {
	record, err := readInitialPublication(owner.state, owner.project, link.Ticket, owner.protection)
	if err != nil {
		return nil, err
	}
	if record == nil || record.ID != link.ID || record.Directory != owner.identity {
		return nil, errors.New("publication transaction does not match its initial selection")
	}
	if record.ExistingLedgerObserved {
		return nil, errors.New("first-publication selection retired after observing an existing ledger")
	}
	if record.CommittedTransaction != "" && record.CommittedTransaction != transaction {
		return nil, errors.New("first-publication selection already consumed by a different transaction")
	}
	return record, nil
}

func (owner *PublicationLock) ConsumeInitialPublication(link InitialPublicationLink, transaction, digest string) error {
	record, err := owner.InitialPublicationForTransaction(link, transaction)
	if err != nil {
		return err
	}
	if record.CommittedTransaction != "" {
		if record.CommittedArtifacts != digest {
			return errors.New("initial publication artifacts changed since commit")
		}
		return syncInitialPublicationState(owner, record.Ticket)
	}
	record.CommittedTransaction = transaction
	record.CommittedArtifacts = digest
	return saveInitialPublication(owner, *record)
}

func saveInitialPublication(owner *PublicationLock, record InitialPublication) error {
	path := initialPublicationPath(owner.state, owner.project, record.Ticket)
	if err := owner.CheckPrivatePath(path); err != nil {
		return err
	}
	body, err := json.Marshal(record)
	if err != nil {
		return err
	}
	if len(body) > initialPublicationLimit {
		return errors.New("first-publication selection exceeds size limit")
	}
	file, err := os.CreateTemp(filepath.Dir(path), ".initial-publication-")
	if err != nil {
		return err
	}
	defer os.Remove(file.Name())
	_, err = file.Write(body)
	if err := errors.Join(err, file.Sync(), file.Close()); err != nil {
		return err
	}
	if err := os.Rename(file.Name(), path); err != nil {
		return err
	}
	return syncInitialPublicationState(owner, record.Ticket)
}

func syncInitialPublicationState(owner *PublicationLock, ticket string) error {
	path := initialPublicationPath(owner.state, owner.project, ticket)
	if err := owner.CheckPrivatePath(path); err != nil {
		return err
	}
	return syncPublicationStateDirectory(filepath.Dir(path))
}
