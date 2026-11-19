package capture

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/domain"
	"golang.org/x/sys/unix"
)

// BindingTicket is the complete user-supplied portion of an inbox response.
// The consumer supplies project and session identity from its private registry.
type BindingTicket struct {
	ID          string `json:"ticket_id"`
	Title       string `json:"ticket_title"`
	Type        string `json:"ticket_type,omitempty"`
	Destination string `json:"publication_destination,omitempty"`
}

type bindingEnvelope struct {
	Version int           `json:"version"`
	Purpose string        `json:"purpose,omitempty"`
	Ticket  BindingTicket `json:"ticket"`
	MAC     string        `json:"mac"`
}

const bindingEnvelopeLimit = 8192

const (
	bindingWithDestination = "binding-with-destination"
	destinationIndication  = "publication-destination"
)

// ErrBindingProofPending identifies source failures after the exact response
// has been authenticated. The worker retries only under a finite allowance.
var ErrBindingProofPending = errors.New("source proof pending")

var errBindingBudget = errors.New("binding inbox byte budget reached")

func (t BindingTicket) validate(cfg domain.TicketConfig) error {
	if !validBindingTicketID(t.ID, cfg) {
		return errors.New("binding requires an exact configured ticket ID")
	}
	if strings.TrimSpace(t.Title) == "" || len(t.Title) > 1024 || !utf8.ValidString(t.Title) || strings.IndexFunc(t.Title, unicode.IsControl) >= 0 {
		return errors.New("binding requires a nonblank single-line ticket title of at most 1024 bytes")
	}
	if typ, ok := domain.NormalizeTicketType(t.Type); !ok || string(typ) != t.Type {
		return errors.New("binding requires an explicit canonical conventional ticket type")
	}
	if t.Destination != "" {
		return validateDestinationHint(t.Destination)
	}
	return nil
}

func (t BindingTicket) validateDestination(cfg domain.TicketConfig) error {
	if !validBindingTicketID(t.ID, cfg) || t.Title != "" || t.Type != "" {
		return errors.New("destination indication requires only the exact bound ticket ID and directory")
	}
	return validateDestinationHint(t.Destination)
}

func validBindingTicketID(value string, cfg domain.TicketConfig) bool {
	id, ok := domain.NormalizeTicketID(value, cfg)
	return ok && id == value && len(value) <= 256
}

func validateDestinationHint(path string) error {
	if strings.TrimSpace(path) == "" || len(path) > 4096 || !utf8.ValidString(path) || strings.IndexFunc(path, unicode.IsControl) >= 0 {
		return errors.New("destination requires a nonblank directory path of at most 4096 bytes")
	}
	// The trusted publisher owns physical resolution and configured scope.
	return nil
}

func (e bindingEnvelope) validate(cfg domain.TicketConfig) error {
	switch e.Purpose {
	case "":
		if e.Version != 1 || e.Ticket.Destination != "" {
			return errors.New("invalid version-1 Codex binding envelope")
		}
		return e.Ticket.validate(cfg)
	case bindingWithDestination:
		if e.Version != 2 || e.Ticket.Destination == "" {
			return errors.New("binding-with-destination requires a version-2 destination")
		}
		return e.Ticket.validate(cfg)
	case destinationIndication:
		if e.Version != 2 {
			return errors.New("destination indication requires a version-2 Codex binding envelope")
		}
		return e.Ticket.validateDestination(cfg)
	default:
		return errors.New("unknown Codex binding envelope purpose")
	}
}

// Each MAC signs an unambiguous tuple of strings in a separate protocol domain.
// A visible slot name or envelope signature cannot authenticate another ticket.
func bindingMAC(key string, fields ...string) string {
	h := hmac.New(sha256.New, []byte(key))
	for _, field := range fields {
		fmt.Fprintf(h, "%d:", len(field))
		io.WriteString(h, field)
	}
	return hex.EncodeToString(h.Sum(nil))
}

func inboxCredential(project string, r domain.CaptureRegistration) string {
	return bindingMAC(r.Request, "codex-binding-credential-v1", project, r.Identity.Store, r.Identity.SessionID, r.InstallationID)
}

func slotName(credential string) string {
	return bindingMAC(credential, "codex-binding-slot-v1") + ".json"
}

func destinationSlotName(credential string) string {
	return bindingMAC(credential, "codex-publication-destination-slot-v1") + ".json"
}

func (t BindingTicket) signature(credential, project string, id domain.CaptureIdentity) string {
	return bindingMAC(credential, "codex-binding-envelope-v1", project, string(id.Source), id.Store, id.SessionID, t.ID, t.Title, t.Type)
}

func (e bindingEnvelope) signature(credential, project string, id domain.CaptureIdentity) string {
	if e.Version == 1 {
		return e.Ticket.signature(credential, project, id)
	}
	t := e.Ticket
	return bindingMAC(credential, "codex-binding-envelope-v2", e.Purpose, project, string(id.Source), id.Store, id.SessionID, t.ID, t.Title, t.Type, t.Destination)
}

// CheckInstalledInboxIsolation is evaluated before trusted capture opens state.
// One global hook may later serve a different repository or a changed local
// configuration; the install-time check alone cannot protect those projects.
func CheckInstalledInboxIsolation(project, home string) error {
	if home == "" {
		return nil
	}
	// The registry is shared by both harnesses. Prove path separation even
	// when the Codex inbox is missing. Its availability matters only when
	// offering or processing a Codex binding.
	home, err := CanonicalStore(home)
	if err != nil {
		return err
	}
	installed, receiptErr := codexInboxInReceipt(home)
	if receiptErr == nil && !installed {
		return nil
	}
	// If the receipt cannot be read, conservatively check the potential grant.
	path, err := absolute(filepath.Join(home, codexInboxName))
	if err != nil {
		return err
	}
	return CheckCodexInboxSeparation(project, home, path)
}

// The sandbox can replace inbox entries. Pin the directory and use single
// component, internally derived names for every filesystem operation.
func openBindingInbox(home string) (int, error) {
	path, err := CodexBindingInbox(home)
	if err != nil {
		return -1, err
	}
	return unix.Open(path, unix.O_RDONLY|unix.O_CLOEXEC|unix.O_NOFOLLOW|unix.O_DIRECTORY, 0)
}

// SubmitCodexBinding writes a response atomically and never opens registry,
// worker, journal, or publication state. Success means queued for verification.
func SubmitCodexBinding(project, home, sessionID, credential string, ticket BindingTicket, cfg domain.TicketConfig) error {
	envelope := bindingEnvelope{Version: 1, Ticket: ticket}
	if ticket.Destination != "" {
		envelope.Version = 2
		envelope.Purpose = bindingWithDestination
	}
	return submitCodexResponse(project, home, sessionID, credential, envelope, cfg)
}

// SubmitCodexDestination queues a response to a separate, bound-ticket request.
// A captured root cannot use its initial binding receipt to change the hint.
func SubmitCodexDestination(project, home, sessionID, credential string, ticket BindingTicket, cfg domain.TicketConfig) error {
	return submitCodexResponse(project, home, sessionID, credential,
		bindingEnvelope{Version: 2, Purpose: destinationIndication, Ticket: ticket}, cfg)
}

func submitCodexResponse(project, home, sessionID, credential string, envelope bindingEnvelope, cfg domain.TicketConfig) error {
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: sessionID}
	if err := id.Validate(); err != nil {
		return err
	}
	if err := envelope.validate(cfg); err != nil {
		return err
	}
	if b, err := hex.DecodeString(credential); err != nil || len(b) != sha256.Size {
		return errors.New("invalid Codex binding request")
	}
	var err error
	project, err = ProjectDirectory(project)
	if err != nil {
		return err
	}
	id, err = canonicalIdentity(id)
	if err != nil {
		return err
	}
	installed, err := CodexInboxInstalled(id.Store)
	if err != nil {
		return err
	}
	if !installed {
		return errors.New("Codex binding inbox is not installed")
	}
	if err := CheckInstalledInboxIsolation(project, id.Store); err != nil {
		return err
	}
	dir, err := openBindingInbox(id.Store)
	if err != nil {
		return fmt.Errorf("open Codex binding inbox; verify the installed integration: %w", err)
	}
	defer unix.Close(dir)
	envelope.MAC = envelope.signature(credential, project, id)
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	if len(body) > bindingEnvelopeLimit {
		return errors.New("Codex binding envelope exceeds its size limit")
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	name := ".submit-" + hex.EncodeToString(nonce[:])
	fd, err := unix.Openat(dir, name, unix.O_WRONLY|unix.O_CREAT|unix.O_EXCL|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	// A failed submission can leave a random temporary entry. The directory
	// is shared with concurrent submitters; never unlink a replaced pathname.
	f := os.NewFile(uintptr(fd), name)
	_, err = f.Write(body)
	if err == nil {
		err = f.Sync()
	}
	if err = errors.Join(err, f.Close()); err != nil {
		return err
	}
	slot := slotName(credential)
	if envelope.Purpose == destinationIndication {
		slot = destinationSlotName(credential)
	}
	if err := unix.Renameat(dir, name, dir, slot); err != nil {
		return err
	}
	return unix.Fsync(dir)
}

func readBindingEnvelope(dir int, name string, remaining *int64) (bindingEnvelope, error) {
	var env bindingEnvelope
	var before unix.Stat_t
	if err := unix.Fstatat(dir, name, &before, unix.AT_SYMLINK_NOFOLLOW); err != nil {
		return env, err
	}
	if before.Mode&unix.S_IFMT != unix.S_IFREG || before.Nlink != 1 {
		return env, errors.New("binding inbox entry must be a single-link bounded regular file")
	}
	fd, err := unix.Openat(dir, name, unix.O_RDONLY|unix.O_NOFOLLOW|unix.O_NONBLOCK|unix.O_CLOEXEC, 0)
	if err != nil {
		return env, err
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	var st unix.Stat_t
	if err := unix.Fstat(fd, &st); err != nil {
		return env, err
	}
	if st.Dev != before.Dev || st.Ino != before.Ino || st.Mode&unix.S_IFMT != unix.S_IFREG || st.Nlink != 1 || st.Size <= 0 || st.Size > bindingEnvelopeLimit {
		return env, errors.New("binding inbox entry must be a single-link bounded regular file")
	}
	if st.Size > *remaining {
		return env, errBindingBudget
	}
	body, err := io.ReadAll(io.LimitReader(f, min(*remaining, bindingEnvelopeLimit)))
	*remaining -= int64(len(body))
	if err != nil {
		return env, err
	}
	var after unix.Stat_t
	if err := unix.Fstat(fd, &after); err != nil {
		return env, err
	}
	if int64(len(body)) != st.Size || after.Size != st.Size {
		return env, errors.New("binding inbox entry changed while reading")
	}
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&env); err != nil {
		return env, errors.New("invalid Codex binding envelope")
	}
	if (env.Version != 1 && env.Version != 2) || !json.Valid(body) {
		return env, errors.New("invalid Codex binding envelope version or content")
	}
	return env, nil
}

func needsExplicitBinding(r domain.CaptureRegistration) bool {
	return r.TicketID == "" || r.TicketTitle == "" || r.TicketType == "" ||
		(r.InstallationID != "" && r.BindingInstallationID != r.InstallationID)
}

// NativePublicationDestinationRequest authorizes a single indication for an
// already verified binding. Each accepted indication retires that request.
func NativePublicationDestinationRequest(r domain.CaptureRegistration) string {
	if r.Request == "" || r.ProjectDirectory == "" || r.ProjectProof != domain.CaptureProjectVerified ||
		r.Binding != "explicit_agent_response" || r.BindingAcceptedAt == "" || r.CallbackObservedAt == "" || needsExplicitBinding(r) {
		return ""
	}
	return bindingMAC(r.Request, "native-publication-destination-request-v1", r.ProjectDirectory,
		string(r.Identity.Source), r.Identity.Store, r.Identity.SessionID, r.InstallationID,
		r.BindingAcceptedAt, r.TicketID, strconv.FormatUint(r.DestinationRevision, 10))
}

// BindNativeOpenCode persists the exact ticket metadata and optional advisory
// destination together after the native tool context has been verified.
func (s *Store) BindNativeOpenCode(ctx context.Context, id domain.CaptureIdentity, ticket BindingTicket, request string) (domain.CaptureRegistration, error) {
	if id.Source != domain.SourceOpenCode || s.project == "" {
		return domain.CaptureRegistration{}, errors.New("OpenCode binding requires the native project and store")
	}
	if err := ticket.validate(s.configuration.Ticket); err != nil {
		return domain.CaptureRegistration{}, err
	}
	return s.change(ctx, id, true, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		if request == "" || !hmac.Equal([]byte(r.Request), []byte(request)) {
			return r, errors.New("binding requires the exact native OpenCode request")
		}
		if !needsExplicitBinding(r) {
			return r, errors.New("binding request is no longer pending")
		}
		if ticket.Destination != "" && r.BindingAcceptedAt != "" && r.BindingInstallationID != r.InstallationID {
			return r, errors.New("renew the OpenCode ticket binding before requesting a new publication destination")
		}
		return s.bindTicketResponse(ctx, r, ticket, nil)
	})
}

// IndicateNativeOpenCodeDestination consumes a request for a single bound
// session. The normal publisher independently validates the advisory path.
func (s *Store) IndicateNativeOpenCodeDestination(ctx context.Context, id domain.CaptureIdentity, ticket BindingTicket, request string) (domain.CaptureRegistration, error) {
	if id.Source != domain.SourceOpenCode || s.project == "" {
		return domain.CaptureRegistration{}, errors.New("OpenCode destination indication requires the native project and store")
	}
	if err := ticket.validateDestination(s.configuration.Ticket); err != nil {
		return domain.CaptureRegistration{}, err
	}
	return s.change(ctx, id, true, func(r domain.CaptureRegistration) (domain.CaptureRegistration, error) {
		return s.applyVerifiedDestination(ctx, r, ticket, request, nil)
	})
}

func (s *Store) bindTicketResponse(ctx context.Context, r domain.CaptureRegistration, ticket BindingTicket, budget *codex.DiscoveryBudget) (domain.CaptureRegistration, error) {
	bound, err := s.bindVerified(ctx, r, ticket.ID, budget)
	if err != nil {
		return r, err
	}
	bound.TicketTitle = ticket.Title
	bound.TicketType = ticket.Type
	if ticket.Destination != "" {
		bound.PublicationDestination = ticket.Destination
	}
	return bound, nil
}

func (s *Store) applyVerifiedDestination(ctx context.Context, r domain.CaptureRegistration, ticket BindingTicket, request string, budget *codex.DiscoveryBudget) (domain.CaptureRegistration, error) {
	want := NativePublicationDestinationRequest(r)
	if want == "" || r.ProjectDirectory != s.project || r.TicketID != ticket.ID || !hmac.Equal([]byte(want), []byte(request)) {
		return r, errors.New("destination indication requires the current native request for the exact bound ticket")
	}
	// Recheck native source ownership without changing the existing binding.
	verified, err := s.bindVerified(ctx, r, ticket.ID, budget)
	if err != nil {
		return r, err
	}
	r.TranscriptPath = verified.TranscriptPath
	r.PublicationDestination = ticket.Destination
	r.DestinationRevision++
	return r, nil
}

func eligibleCodexBinding(home string, r domain.CaptureRegistration) bool {
	return r.Identity.Source == domain.SourceCodex && r.Identity.Store == home && r.CallbackObservedAt != "" &&
		(needsExplicitBinding(r) || NativePublicationDestinationRequest(r) != "")
}

// ConsumeCodexBindings probes only challenge-derived entries belonging to this
// project's observed sessions. A rolling cursor bounds probes without allowing
// old unassigned registrations or malformed submissions to starve later ones.
func (s *Store) ConsumeCodexBindings(ctx context.Context, home string, rows []domain.CaptureRegistration, cursor *int, maxSessions int, maxBytes int64) (int, error) {
	accepted, _, err := s.consumeCodexBindings(ctx, home, rows, cursor, maxSessions, maxBytes)
	return accepted, err
}

// BindingScan persists the last registration inspected and the amount of work
// left in a sweep for the most recent publication request. A bounded successor
// can finish a wake even when one process lifetime is shorter than a full sweep.
type BindingScan struct {
	Generation int64  `json:"generation"`
	Last       string `json:"last,omitempty"`
	Remaining  int    `json:"remaining"`
	RetryUntil int64  `json:"retry_until,omitempty"`
}

func (s *Store) ScanCodexBindings(ctx context.Context, home string, rows []domain.CaptureRegistration, scan *BindingScan, generation int64, maxSessions int, maxBytes int64) (int, error) {
	if home == "" {
		scan.Remaining = 0
		return 0, nil
	}
	home, err := CanonicalStore(home)
	if err != nil {
		return 0, err
	}
	var eligible []domain.CaptureRegistration
	for _, r := range rows { // Store.List orders these by IdentityKey.
		if eligibleCodexBinding(home, r) {
			eligible = append(eligible, r)
		}
	}
	if generation != scan.Generation {
		scan.RetryUntil = 0
		scan.Generation = generation
		scan.Remaining = len(eligible)
	}
	scan.Remaining = min(scan.Remaining, len(eligible))
	if len(eligible) == 0 {
		scan.RetryUntil = 0
		return 0, nil
	}
	cursor := sort.Search(len(eligible), func(i int) bool { return IdentityKey(eligible[i].Identity) > scan.Last }) % len(eligible)
	accepted, visited, err := s.consumeCodexBindings(ctx, home, eligible, &cursor, maxSessions, maxBytes)
	if visited > 0 {
		scan.Last = IdentityKey(eligible[(cursor+len(eligible)-1)%len(eligible)].Identity)
		scan.Remaining = max(0, scan.Remaining-visited)
	}
	return accepted, err
}

func (s *Store) consumeCodexBindings(ctx context.Context, home string, rows []domain.CaptureRegistration, cursor *int, maxSessions int, maxBytes int64) (int, int, error) {
	if home == "" || len(rows) == 0 {
		return 0, 0, nil
	}
	home, err := CanonicalStore(home)
	if err != nil {
		return 0, 0, err
	}
	dir := -1
	accepted := 0
	var firstErr error
	visited := 0
	budget := codex.DiscoveryBudget{Context: ctx, MaxSessions: maxSessions, MaxSourceBytes: maxBytes}
	for visited < min(len(rows), maxSessions) && budget.SourceBytesRead < maxBytes {
		if err := ctx.Err(); err != nil {
			return accepted, visited, errors.Join(firstErr, err)
		}
		visited++
		index := *cursor % len(rows)
		r := rows[index]
		*cursor = (*cursor + 1) % len(rows)
		if !eligibleCodexBinding(home, r) {
			continue
		}
		if dir < 0 {
			installed, err := CodexInboxInstalled(home)
			if err != nil || !installed {
				return accepted, visited, err
			}
			dir, err = openBindingInbox(home)
			if err != nil {
				return accepted, visited, err
			}
			defer unix.Close(dir)
		}
		credential := inboxCredential(s.project, r)
		name := slotName(credential)
		destination := !needsExplicitBinding(r)
		if destination {
			credential = NativePublicationDestinationRequest(r)
			name = destinationSlotName(credential)
		}
		sharedAllowance := budget.SourceBytesRead > 0 || budget.SessionsRead > 0
		remaining := maxBytes - budget.SourceBytesRead
		env, err := readBindingEnvelope(dir, name, &remaining)
		budget.SourceBytesRead = maxBytes - remaining
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err == nil {
			want := env.signature(credential, s.project, r.Identity)
			if destination != (env.Purpose == destinationIndication) || !hmac.Equal([]byte(want), []byte(env.MAC)) {
				err = errors.New("binding authentication failed")
			}
		}
		if err == nil {
			err = env.validate(s.configuration.Ticket)
		}
		if err == nil {
			_, err = s.change(ctx, r.Identity, true, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				if destination {
					return s.applyVerifiedDestination(ctx, current, env.Ticket, credential, &budget)
				}
				if !needsExplicitBinding(current) || current.CallbackObservedAt == "" || inboxCredential(s.project, current) != credential {
					return current, errors.New("binding request is no longer pending")
				}
				return s.bindTicketResponse(ctx, current, env.Ticket, &budget)
			})
			if err == nil {
				accepted++
				// A completed registration rejects replays. Leave the slot in place:
				// a replacement after our read belongs to its writer.
			}
		}
		// A response that only exceeded the remainder gets first use of the
		// next pass. One that exceeds a whole pass cannot starve later owners.
		if sharedAllowance && (errors.Is(err, errBindingBudget) || errors.Is(err, codex.ErrDiscoveryBudget)) {
			*cursor = index
			return accepted, visited - 1, errors.Join(firstErr, err)
		}
		if err != nil && (firstErr == nil || (errors.Is(err, ErrBindingProofPending) && !errors.Is(firstErr, ErrBindingProofPending))) {
			firstErr = errors.Join(firstErr, fmt.Errorf("Codex binding for session %s: %w", r.Identity.SessionID, err))
		}
	}
	return accepted, visited, firstErr
}
