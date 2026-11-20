package adapters

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"syscall"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/domain"
)

// PendingSnapshot is the portable, per-root record of native metrics collected
// before an automatic publication. It carries no destination authority.
type PendingSnapshot struct {
	Version       int                  `json:"version"`
	CaptureID     string               `json:"capture_id"`
	Source        domain.SourceHarness `json:"source"`
	RootSessionID string               `json:"root_session_id"`
	SourceCutoff  string               `json:"source_cutoff"`
	State         string               `json:"state"`
	TicketID      string               `json:"ticket_id,omitempty"`
	TicketTitle   string               `json:"ticket_title,omitempty"`
	TicketType    string               `json:"ticket_type,omitempty"`
	Sessions      []SessionFactDTO     `json:"sessions"`
}

// PendingSnapshotEntry is an exact observation retained for promotion cleanup.
type PendingSnapshotEntry struct {
	Snapshot PendingSnapshot
	digest   [sha256.Size]byte
}

func pendingSnapshotState(ticket string) string {
	if ticket == "" {
		return "awaiting_ticket"
	}
	return "awaiting_destination"
}

func (snapshot PendingSnapshot) validate() error {
	root := domain.CaptureIdentity{Source: snapshot.Source, Store: snapshot.CaptureID, SessionID: snapshot.RootSessionID}
	if err := root.Validate(); err != nil {
		return err
	}
	if snapshot.Version != 1 || snapshot.CaptureID != capture.StableCaptureID(snapshot.Source, snapshot.RootSessionID) || snapshot.State != pendingSnapshotState(snapshot.TicketID) {
		return errors.New("invalid pending snapshot version, identity or state")
	}
	if _, err := time.Parse(time.RFC3339Nano, snapshot.SourceCutoff); err != nil {
		return fmt.Errorf("invalid pending snapshot source cutoff: %w", err)
	}
	if ticket, err := domain.CaptureBinding("", snapshot.TicketID, ""); err != nil {
		return err
	} else if ticket != snapshot.TicketID {
		return errors.New("pending snapshot has a blank ticket identity")
	}
	if snapshot.TicketID == "" && (snapshot.TicketTitle != "" || snapshot.TicketType != "") {
		return errors.New("unassigned pending snapshot has ticket metadata")
	}
	if kind, ok := domain.NormalizeTicketType(snapshot.TicketType); snapshot.TicketType != "" && (!ok || string(kind) != snapshot.TicketType) {
		return errors.New("pending snapshot has an invalid ticket type")
	}
	version := domain.CurrentSchemaVersion
	if err := (AgenticFactLedgerDTO{SchemaVersion: &version, TicketID: &snapshot.CaptureID, Sessions: snapshot.Sessions}).ValidatePersistedEvidence(); err != nil {
		return err
	}
	parents := make(map[string]string, len(snapshot.Sessions))
	for _, session := range snapshot.Sessions {
		if err := (domain.CaptureIdentity{Source: domain.SourceHarness(session.Source), Store: snapshot.CaptureID, SessionID: session.ID}).Validate(); err != nil {
			return err
		}
		if session.Source != string(snapshot.Source) {
			return errors.New("pending snapshot contains facts from a different source")
		}
		parents[session.ID] = ""
		if session.ParentID != nil {
			parents[session.ID] = *session.ParentID
		}
	}
	if _, ok := parents[snapshot.RootSessionID]; !ok {
		return errors.New("pending snapshot does not contain its root")
	}
	if _, internal := parents[parents[snapshot.RootSessionID]]; internal {
		return errors.New("pending snapshot root has a parent inside its own group")
	}
	reachable := map[string]bool{snapshot.RootSessionID: true}
	for id := range parents {
		var chain []string
		for !reachable[id] {
			if _, ok := parents[id]; !ok || len(chain) >= len(parents) {
				return errors.New("pending snapshot contains an unrelated or cyclic owner")
			}
			chain = append(chain, id)
			id = parents[id]
		}
		for _, child := range chain {
			reachable[child] = true
		}
	}
	return nil
}

func (entry PendingSnapshotEntry) Sessions() []domain.SessionFact {
	facts := make([]domain.SessionFact, 0, len(entry.Snapshot.Sessions))
	for _, session := range entry.Snapshot.Sessions {
		facts = append(facts, session.ToDomain())
	}
	return facts
}

func readPendingSnapshot(path string) (PendingSnapshot, []byte, error) {
	before, err := capture.LstatPublicationFile(path)
	if errors.Is(err, os.ErrNotExist) {
		return PendingSnapshot{}, nil, nil
	}
	if err != nil {
		return PendingSnapshot{}, nil, err
	}
	if !before.Mode().IsRegular() || before.Size() > publicationLimit {
		return PendingSnapshot{}, nil, errors.New("pending snapshot must be a bounded regular file")
	}
	file, err := capture.OpenPublicationFile(path, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return PendingSnapshot{}, nil, err
	}
	defer file.Close()
	opened, err := file.Stat()
	if err != nil {
		return PendingSnapshot{}, nil, err
	}
	if !os.SameFile(before, opened) {
		return PendingSnapshot{}, nil, errors.New("pending snapshot changed during selection")
	}
	body, err := io.ReadAll(io.LimitReader(file, publicationLimit+1))
	if err != nil {
		return PendingSnapshot{}, nil, err
	}
	if len(body) > publicationLimit {
		return PendingSnapshot{}, nil, errors.New("pending snapshot exceeds bounded read")
	}
	var snapshot PendingSnapshot
	decoder := newStrictFactDecoder(body)
	if err := decoder.Decode(&snapshot); err != nil {
		return PendingSnapshot{}, nil, err
	}
	if err := requireFactJSONEOF(decoder, "pending snapshot"); err != nil {
		return PendingSnapshot{}, nil, err
	}
	if err := snapshot.validate(); err != nil {
		return PendingSnapshot{}, nil, err
	}
	if filepath.Base(path) != snapshot.CaptureID+".json" {
		return PendingSnapshot{}, nil, errors.New("pending snapshot filename does not match its identity")
	}
	return snapshot, body, nil
}

type pendingDirectoryAdmission struct {
	owner    *capture.PublicationLock
	borrowed bool
}

func (admission *pendingDirectoryAdmission) Close() error {
	if admission.borrowed {
		return nil
	}
	return admission.owner.Close()
}

func pendingDirectoryLease(ctx context.Context, directory, project string, create bool, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (*pendingDirectoryAdmission, error) {
	rel, err := filepath.Rel(project, directory)
	if err != nil || rel == "." || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return nil, errors.New("pending directory must be within the physical project")
	}
	root, err := os.OpenRoot(project)
	if err != nil {
		return nil, err
	}
	defer root.Close()
	current, err := capture.ProjectDirectory(project)
	if err != nil || current != project {
		return nil, errors.Join(errors.New("pending project moved before directory admission"), err)
	}
	pathInfo, statErr := os.Stat(project)
	openedInfo, rootErr := root.Stat(".")
	if statErr != nil || rootErr != nil || !os.SameFile(pathInfo, openedInfo) {
		return nil, errors.Join(errors.New("pending project changed during directory admission"), statErr, rootErr)
	}
	info, err := root.Lstat(rel)
	if errors.Is(err, os.ErrNotExist) {
		if !create {
			return nil, nil
		}
		if err := root.MkdirAll(rel, 0755); err != nil {
			return nil, err
		}
		info, err = root.Lstat(rel)
	}
	if err != nil {
		return nil, err
	}
	if !info.IsDir() {
		return nil, errors.New("configured pending path is not a regular directory")
	}
	member := filepath.Join(directory, ".pending-snapshot")
	if canonical != nil && canonical.DirectoryPath() == directory {
		if canonical.Project() != project || capture.PublicationOwner(member) != canonical {
			return nil, errors.New("matching canonical publication ownership is no longer held")
		}
		if err := canonical.ValidateLocation(); err != nil {
			return nil, err
		}
		admission := &pendingDirectoryAdmission{owner: canonical, borrowed: true}
		return admitDurablePendingDirectory(ctx, root, rel, directory, admission)
	}
	owner, err := capture.AcquireSelectedPublication(ctx, member, project, info, protection)
	if err != nil {
		return nil, err
	}
	return admitDurablePendingDirectory(ctx, root, rel, directory, &pendingDirectoryAdmission{owner: owner})
}

// Repeat the complete directory barrier on retries. A prior attempt may have
// renamed or removed a file before syncing the leaf or its new ancestors.
func admitDurablePendingDirectory(ctx context.Context, root *os.Root, relative, directory string, admission *pendingDirectoryAdmission) (*pendingDirectoryAdmission, error) {
	for parent := relative; ; parent = filepath.Dir(parent) {
		if err := ctx.Err(); err != nil {
			return nil, errors.Join(err, admission.Close())
		}
		if err := checkPendingDirectory(admission.owner, filepath.Join(directory, ".pending-snapshot")); err != nil {
			return nil, errors.Join(err, admission.Close())
		}
		f, err := root.Open(parent)
		if err == nil {
			err = errors.Join(f.Sync(), f.Close())
		}
		if err != nil {
			return nil, errors.Join(fmt.Errorf("syncing pending directory path %s: %w", parent, err), admission.Close())
		}
		if parent == "." {
			return admission, nil
		}
	}
}

func verifyPendingConfiguration(project, directory string) error {
	cfg, err := config.Load(project)
	if err != nil {
		return err
	}
	configured, err := config.ResolvePendingDirectory(cfg, project)
	if err != nil {
		return err
	}
	if configured != directory {
		return errors.New("configured pending directory changed during capture; retry publication")
	}
	return nil
}

func checkPendingDirectory(owner *capture.PublicationLock, path string) error {
	actualDirectory, err := owner.PhysicalMemberPath(filepath.Join(filepath.Dir(path), "."))
	if err != nil {
		return err
	}
	if actualDirectory != filepath.Dir(path) {
		return errors.New("pending directory moved during capture; retry using the configured directory")
	}
	return nil
}

type pendingOwnerGroup struct {
	root  domain.CaptureRegistration
	facts map[string]domain.SessionFact
}

func pendingOwnerGroups(project string, result capture.ReconcileResult) (map[string]*pendingOwnerGroup, map[string]domain.CaptureRegistration, map[string]string) {
	rows := make(map[string]domain.CaptureRegistration, len(result.ProjectRegistrations))
	for _, row := range result.ProjectRegistrations {
		if row.ProjectDirectory == project && row.ProjectProof == domain.CaptureProjectVerified {
			rows[capture.IdentityKey(row.Identity)] = row
		}
	}
	groups := map[string]*pendingOwnerGroup{}
	assignments := map[string]string{}
	for _, row := range result.Registrations {
		key := capture.IdentityKey(row.Identity)
		if _, verified := rows[key]; !verified {
			continue
		}
		root := row
		for depth := 0; root.ParentID != ""; depth++ {
			parentID := root.Identity
			parentID.SessionID = root.ParentID
			parent, ok := rows[capture.IdentityKey(parentID)]
			if ok && parent.TicketID == result.TicketID && depth < len(rows) {
				root = parent
				continue
			}
			// A separately bound native child may own a different ticket
			// while recording its original parent, which remains outside its group.
			if root.TicketID != "" && root.Binding == "explicit_agent_response" && (!ok || parent.TicketID != result.TicketID) {
				break
			}
			root = domain.CaptureRegistration{}
			break
		}
		if root.Identity.SessionID == "" {
			continue
		}
		id := capture.StableCaptureID(root.Identity.Source, root.Identity.SessionID)
		assignments[key] = id
		group := groups[id]
		if group == nil {
			group = &pendingOwnerGroup{root: root, facts: map[string]domain.SessionFact{}}
			groups[id] = group
		}
		for _, fact := range row.Snapshot {
			if fact.ID == row.Identity.SessionID && fact.Source == row.Identity.Source {
				group.facts[fact.ID] = fact
			}
		}
	}
	return groups, rows, assignments
}

// SyncPendingSnapshots writes exact project-verified roots and their native
// descendants. Accepted unavailable owners survive refresh within their group.
func SyncPendingSnapshots(ctx context.Context, store *capture.Store, directory, project string, result capture.ReconcileResult, create bool, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (entries []PendingSnapshotEntry, returnErr error) {
	groups, rows, assignments := pendingOwnerGroups(project, result)
	if len(groups) == 0 {
		return nil, nil
	}
	admission, err := pendingDirectoryLease(ctx, directory, project, create, protection, canonical)
	if err != nil || admission == nil {
		return nil, err
	}
	defer func() { returnErr = errors.Join(returnErr, admission.Close()) }()
	if err := verifyPendingConfiguration(project, directory); err != nil {
		return nil, err
	}
	returnErr = store.WithCaptureLifecycle(ctx, result.LifecycleRevision, func() error {
		var err error
		entries, err = syncPendingOwnerGroups(ctx, directory, admission.owner, result, groups, rows, assignments, create)
		return err
	})
	return entries, returnErr
}

func syncPendingOwnerGroups(ctx context.Context, directory string, owner *capture.PublicationLock, result capture.ReconcileResult, groups map[string]*pendingOwnerGroup, rows map[string]domain.CaptureRegistration, assignments map[string]string, create bool) ([]PendingSnapshotEntry, error) {
	entries := make([]PendingSnapshotEntry, 0, len(groups))
	title, kind, _ := domain.CaptureTicketMetadata(result.Registrations, result.TicketID)
	refreshed := map[string]bool{}
	for _, session := range result.RefreshedSessions {
		refreshed[session.ID] = true
	}
	ids := make([]string, 0, len(groups))
	for id := range groups {
		ids = append(ids, id)
	}
	sort.Strings(ids)
	for _, id := range ids {
		if err := ctx.Err(); err != nil {
			return entries, err
		}
		group := groups[id]
		path := filepath.Join(directory, id+".json")
		if err := checkPendingDirectory(owner, path); err != nil {
			return entries, err
		}
		prior, previous, err := readPendingSnapshot(path)
		if err != nil {
			return entries, fmt.Errorf("reading pending snapshot %s: %w", id, err)
		}
		if previous == nil && !create {
			continue
		}
		if prior.TicketID != "" && prior.TicketID != result.TicketID {
			if result.TicketID == "" {
				// Binding may arrive during a pre-binding reconciliation.
				continue
			}
			return entries, fmt.Errorf("pending snapshot %s has conflicting ticket metadata", id)
		}
		facts := group.facts
		for _, oldFact := range prior.Sessions {
			identity := group.root.Identity
			identity.SessionID = oldFact.ID
			key := capture.IdentityKey(identity)
			if row, known := rows[key]; known && (row.TicketID != result.TicketID || assignments[key] != id) {
				continue
			}
			if _, present := facts[oldFact.ID]; !present || !refreshed[oldFact.ID] {
				facts[oldFact.ID] = oldFact.ToDomain()
			}
		}
		sessions := pendingConnectedFacts(group.root.Identity.SessionID, facts)
		if len(sessions) == 0 {
			continue
		}
		snapshot := PendingSnapshot{
			Version: 1, CaptureID: id, Source: group.root.Identity.Source,
			RootSessionID: group.root.Identity.SessionID, SourceCutoff: result.StartedAt,
			State: pendingSnapshotState(result.TicketID), TicketID: result.TicketID, TicketTitle: title, TicketType: kind, Sessions: sessions,
		}
		if group.root.TicketTitle != "" {
			snapshot.TicketTitle = group.root.TicketTitle
		}
		if group.root.TicketType != "" {
			snapshot.TicketType = group.root.TicketType
		}
		// An unchanged payload keeps its original cutoff and bytes.
		if previous != nil && samePendingFacts(prior, snapshot) {
			entries = append(entries, PendingSnapshotEntry{Snapshot: prior, digest: sha256.Sum256(previous)})
			continue
		}
		if err := snapshot.validate(); err != nil {
			return entries, fmt.Errorf("validating pending snapshot %s: %w", id, err)
		}
		body, err := json.MarshalIndent(snapshot, "", "  ")
		if err != nil {
			return entries, err
		}
		body = append(body, '\n')
		if len(body) > publicationLimit {
			return entries, errors.New("pending snapshot exceeds publication bound")
		}
		if !bytes.Equal(previous, body) {
			if err := durableReplace(path, body); err != nil {
				return entries, fmt.Errorf("publishing pending snapshot %s: %w", id, err)
			}
		}
		entries = append(entries, PendingSnapshotEntry{Snapshot: snapshot, digest: sha256.Sum256(body)})
	}
	return entries, nil
}

func samePendingFacts(left, right PendingSnapshot) bool {
	left.SourceCutoff, right.SourceCutoff = "", ""
	leftJSON, leftErr := json.Marshal(left)
	rightJSON, rightErr := json.Marshal(right)
	return leftErr == nil && rightErr == nil && bytes.Equal(leftJSON, rightJSON)
}

func pendingConnectedFacts(root string, facts map[string]domain.SessionFact) []SessionFactDTO {
	if _, ok := facts[root]; !ok {
		return nil
	}
	children := map[string][]string{}
	for id, fact := range facts {
		if id != root && fact.ParentID != nil {
			children[*fact.ParentID] = append(children[*fact.ParentID], id)
		}
	}
	connected := []string{root}
	for i := 0; i < len(connected); i++ {
		connected = append(connected, children[connected[i]]...)
	}
	sort.Strings(connected)
	sessions := make([]SessionFactDTO, 0, len(connected))
	for _, id := range connected {
		sessions = append(sessions, SessionFactDTOFromDomain(facts[id]))
	}
	return sessions
}

// RemovePromotedPendingSnapshots runs only after the normal canonical transaction
// commits. An independently replaced snapshot survives until the next promotion.
func RemovePromotedPendingSnapshots(ctx context.Context, store *capture.Store, revision int64, directory, project string, entries []PendingSnapshotEntry, protection *capture.ProtectedPublication, canonical *capture.PublicationLock) (returnErr error) {
	if len(entries) == 0 {
		return nil
	}
	admission, err := pendingDirectoryLease(ctx, directory, project, false, protection, canonical)
	if err != nil || admission == nil {
		return err
	}
	defer func() { returnErr = errors.Join(returnErr, admission.Close()) }()
	if err := verifyPendingConfiguration(project, directory); err != nil {
		return err
	}
	return store.WithCaptureLifecycle(ctx, revision, func() error {
		return removePromotedPendingEntries(ctx, directory, admission.owner, entries)
	})
}

func removePromotedPendingEntries(ctx context.Context, directory string, owner *capture.PublicationLock, entries []PendingSnapshotEntry) error {
	for _, entry := range entries {
		if err := ctx.Err(); err != nil {
			return err
		}
		path := filepath.Join(directory, entry.Snapshot.CaptureID+".json")
		if err := checkPendingDirectory(owner, path); err != nil {
			return err
		}
		_, body, err := readPendingSnapshot(path)
		if err != nil {
			return err
		}
		if body == nil {
			continue
		}
		if sha256.Sum256(body) != entry.digest {
			return errors.New("pending snapshot changed during promotion; retry without deleting newer facts")
		}
		if err := capture.RemovePublicationFile(path); err != nil {
			return err
		}
		if err := syncDirectory(directory); err != nil {
			return err
		}
	}
	return nil
}

// RemoveDiscardedPendingSnapshot acknowledges cleanup only for the current
// discard revision and the validated artifact for its native root.
func RemoveDiscardedPendingSnapshot(ctx context.Context, store *capture.Store, directory, project string, discard capture.CaptureControl, protection *capture.ProtectedPublication) (returnErr error) {
	if err := discard.Root.Validate(); err != nil {
		return err
	}
	if discard.CaptureID != capture.StableCaptureID(discard.Root.Source, discard.Root.SessionID) {
		return errors.New("discarded snapshot does not match the verified capture identity")
	}
	admission, err := pendingDirectoryLease(ctx, directory, project, false, protection, nil)
	if err != nil {
		return err
	}
	if admission != nil {
		defer func() { returnErr = errors.Join(returnErr, admission.Close()) }()
	}
	if err := verifyPendingConfiguration(project, directory); err != nil {
		return err
	}
	return store.WithDiscardedCapture(ctx, discard.CaptureID, discard.Revision, func() error {
		if admission == nil {
			return nil
		}
		path := filepath.Join(directory, discard.CaptureID+".json")
		if err := checkPendingDirectory(admission.owner, path); err != nil {
			return err
		}
		snapshot, body, err := readPendingSnapshot(path)
		if err != nil || body == nil {
			return err
		}
		if snapshot.Source != discard.Root.Source || snapshot.RootSessionID != discard.Root.SessionID {
			return errors.New("pending snapshot belongs to a different capture")
		}
		if err := capture.RemovePublicationFile(path); err != nil {
			return err
		}
		return syncDirectory(directory)
	})
}
