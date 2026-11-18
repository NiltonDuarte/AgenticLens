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
	ID    string `json:"ticket_id"`
	Title string `json:"ticket_title"`
	Type  string `json:"ticket_type,omitempty"`
}

type bindingEnvelope struct {
	Version int           `json:"version"`
	Ticket  BindingTicket `json:"ticket"`
	MAC     string        `json:"mac"`
}

const bindingEnvelopeLimit = 8192

var errBindingBudget = errors.New("binding inbox byte budget reached")

func (t BindingTicket) validate(cfg domain.TicketConfig) error {
	id, ok := domain.NormalizeTicketID(t.ID, cfg)
	if !ok || id != t.ID || len(t.ID) > 256 {
		return errors.New("binding requires an exact configured ticket ID")
	}
	if strings.TrimSpace(t.Title) == "" || len(t.Title) > 1024 || !utf8.ValidString(t.Title) || strings.IndexFunc(t.Title, unicode.IsControl) >= 0 {
		return errors.New("binding requires a nonblank single-line ticket title of at most 1024 bytes")
	}
	if t.Type != "" {
		if typ, ok := domain.NormalizeTicketType(t.Type); !ok || string(typ) != t.Type {
			return errors.New("binding requires a canonical conventional ticket type")
		}
	}
	return nil
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

func (t BindingTicket) signature(credential, project string, id domain.CaptureIdentity) string {
	return bindingMAC(credential, "codex-binding-envelope-v1", project, string(id.Source), id.Store, id.SessionID, t.ID, t.Title, t.Type)
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
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: sessionID}
	if err := id.Validate(); err != nil {
		return err
	}
	if err := ticket.validate(cfg); err != nil {
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
	body, err := json.Marshal(bindingEnvelope{Version: 1, Ticket: ticket, MAC: ticket.signature(credential, project, id)})
	if err != nil {
		return err
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
	if err := unix.Renameat(dir, name, dir, slotName(credential)); err != nil {
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
	if env.Version != 1 || !json.Valid(body) {
		return env, errors.New("invalid Codex binding envelope version or content")
	}
	return env, nil
}

func needsExplicitBinding(r domain.CaptureRegistration) bool {
	return r.TicketID == "" || r.TicketTitle == "" || (r.InstallationID != "" && r.BindingInstallationID != r.InstallationID)
}

func eligibleCodexBinding(home string, r domain.CaptureRegistration) bool {
	return r.Identity.Source == domain.SourceCodex && r.Identity.Store == home && r.CallbackObservedAt != "" && needsExplicitBinding(r)
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
		scan.Generation = generation
		scan.Remaining = len(eligible)
	}
	scan.Remaining = min(scan.Remaining, len(eligible))
	if len(eligible) == 0 {
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
		sharedAllowance := budget.SourceBytesRead > 0 || budget.SessionsRead > 0
		remaining := maxBytes - budget.SourceBytesRead
		env, err := readBindingEnvelope(dir, name, &remaining)
		budget.SourceBytesRead = maxBytes - remaining
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err == nil {
			want := env.Ticket.signature(credential, s.project, r.Identity)
			if !hmac.Equal([]byte(want), []byte(env.MAC)) {
				err = errors.New("binding authentication failed")
			}
		}
		if err == nil {
			err = env.Ticket.validate(s.configuration.Ticket)
		}
		if err == nil {
			_, err = s.change(ctx, r.Identity, true, func(current domain.CaptureRegistration) (domain.CaptureRegistration, error) {
				if !needsExplicitBinding(current) || current.CallbackObservedAt == "" || inboxCredential(s.project, current) != credential {
					return current, errors.New("binding request is no longer pending")
				}
				bound, e := s.bindVerified(ctx, current, env.Ticket.ID, &budget)
				if e != nil {
					return current, e
				}
				bound.TicketTitle = env.Ticket.Title
				if env.Ticket.Type != "" {
					bound.TicketType = env.Ticket.Type
				}
				return bound, nil
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
		if err != nil && firstErr == nil {
			firstErr = fmt.Errorf("Codex binding for session %s: %w", r.Identity.SessionID, err)
		}
	}
	return accepted, visited, firstErr
}
