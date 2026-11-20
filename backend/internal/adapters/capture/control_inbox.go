package capture

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"agenticlens/internal/domain"
	"golang.org/x/sys/unix"
)

// NativeCaptureControl is an invitation scoped to an exact verified root and
// the current operation revision. It uses a separate purpose from binding.
type NativeCaptureControl struct {
	CaptureID string `json:"capture_id"`
	Action    string `json:"action"`
	Request   string `json:"request"`
}

type captureControlEnvelope struct {
	Version   int    `json:"version"`
	CaptureID string `json:"capture_id"`
	Action    string `json:"action"`
	Revision  int64  `json:"revision"`
	MAC       string `json:"mac"`
}

func controlInboxEligible(row domain.CaptureRegistration) bool {
	return row.ProjectProof == domain.CaptureProjectVerified && row.CallbackObservedAt != ""
}

func controlInboxKey(seed, project string, row domain.CaptureRegistration, revision int64) string {
	return bindingMAC(seed, "capture-control-slot-v1", project, string(row.Identity.Source), row.Identity.Store,
		row.Identity.SessionID, row.InstallationID, strconv.FormatInt(revision, 10))
}

func controlRequestKey(seed, project string, row domain.CaptureRegistration, captureID, action string, revision int64) string {
	return bindingMAC(seed, "capture-control-request-v1", project, string(row.Identity.Source), row.Identity.Store,
		row.Identity.SessionID, row.InstallationID, captureID, action, strconv.FormatInt(revision, 10))
}

func (e captureControlEnvelope) signature(key, project string, identity domain.CaptureIdentity) string {
	return bindingMAC(key, "capture-control-response-v1", project, string(identity.Source), identity.Store,
		identity.SessionID, e.CaptureID, e.Action, strconv.FormatInt(e.Revision, 10))
}

func controlAction(state CaptureControlState) string {
	switch state {
	case CaptureControlActive, CaptureControlRestoring:
		return "discard"
	case CaptureControlDiscarded:
		return "restore"
	default:
		return ""
	}
}

func validControlDigest(value string) bool {
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size && hex.EncodeToString(decoded) == value
}

func parseControlRequest(request string) (slot, auth string, revision int64, err error) {
	parts := strings.SplitN(request, ".", 3)
	if len(parts) != 3 || !validControlDigest(parts[0]) || !validControlDigest(parts[2]) {
		return "", "", 0, errors.New("invalid native capture control request")
	}
	revision, err = strconv.ParseInt(parts[1], 10, 64)
	if err != nil || revision < 0 || strconv.FormatInt(revision, 10) != parts[1] {
		return "", "", 0, errors.New("invalid native capture control revision")
	}
	return parts[0], parts[2], revision, nil
}

func controlInboxName(slot string) string { return "control-" + slot + ".json" }

// CaptureControlInvitation authorizes the next state transition for this
// native session only when it owns the verified capture root.
func (s *Store) CaptureControlInvitation(ctx context.Context, row domain.CaptureRegistration) (*NativeCaptureControl, error) {
	return s.captureControlInvitation(ctx, row, true)
}

func (s *Store) captureControlInvitation(ctx context.Context, row domain.CaptureRegistration, issue bool) (*NativeCaptureControl, error) {
	if !controlInboxEligible(row) || row.ProjectDirectory != s.project {
		return nil, nil
	}
	if row.Identity.Source == domain.SourceOpenCode {
		root, err := s.openCodeControlRoot(ctx, row)
		if err != nil || !root {
			return nil, err
		}
	}
	status, err := s.CaptureControlForRoot(ctx, row.Identity)
	if err != nil {
		return nil, err
	}
	action := controlAction(status.State)
	if action == "" || status.Root != row.Identity {
		return nil, nil
	}
	seed, err := s.captureControlSeed(ctx, row.Identity, issue)
	if err != nil || seed == "" {
		return nil, err
	}
	request := strings.Join([]string{controlInboxKey(seed, s.project, row, status.Revision), strconv.FormatInt(status.Revision, 10),
		controlRequestKey(seed, s.project, row, status.CaptureID, action, status.Revision)}, ".")
	return &NativeCaptureControl{CaptureID: status.CaptureID, Action: action, Request: request}, nil
}

// SubmitCodexCaptureControl writes only to the installed shared inbox. The
// trusted worker resolves root ownership and applies the revision fence.
func SubmitCodexCaptureControl(project, home, sessionID, captureID, action, request string) error {
	if !validControlDigest(captureID) || (action != "discard" && action != "restore") {
		return errors.New("capture control requires an exact capture ID and discard or restore")
	}
	slot, auth, revision, err := parseControlRequest(request)
	if err != nil {
		return err
	}
	id := domain.CaptureIdentity{Source: domain.SourceCodex, Store: home, SessionID: sessionID}
	if err := id.Validate(); err != nil {
		return err
	}
	dir, project, id, err := openCodexSubmission(project, id)
	if err != nil {
		return err
	}
	defer unix.Close(dir)
	envelope := captureControlEnvelope{Version: 1, CaptureID: captureID, Action: action, Revision: revision}
	envelope.MAC = envelope.signature(auth, project, id)
	body, err := json.Marshal(envelope)
	if err != nil {
		return err
	}
	return writeCodexInboxEntry(dir, controlInboxName(slot), body)
}

func (s *Store) consumeCodexCaptureControl(ctx context.Context, dir int, row domain.CaptureRegistration, remaining *int64) (bool, error) {
	if !controlInboxEligible(row) || row.ProjectDirectory != s.project {
		return false, nil
	}
	seed, err := s.captureControlSeed(ctx, row.Identity, false)
	if err != nil || seed == "" {
		return false, err
	}
	var revision int64
	captureID := StableCaptureID(row.Identity.Source, row.Identity.SessionID)
	err = s.db.QueryRowContext(ctx, `SELECT revision FROM capture_controls WHERE worktree_id=? AND capture_id=?`, s.worktreeID, captureID).Scan(&revision)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return false, err
	}
	// Each operation revision owns a different slot. A delayed submission
	// cannot replace a newer response waiting for the worker.
	body, err := readCodexInboxEntry(dir, controlInboxName(controlInboxKey(seed, s.project, row, revision)), remaining)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	var response captureControlEnvelope
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&response); err != nil || !json.Valid(body) || response.Version != 1 ||
		response.CaptureID != captureID || response.Revision != revision || (response.Action != "discard" && response.Action != "restore") {
		return false, errors.New("invalid Codex capture control envelope")
	}
	key := controlRequestKey(seed, s.project, row, response.CaptureID, response.Action, response.Revision)
	if !hmac.Equal([]byte(response.signature(key, s.project, row.Identity)), []byte(response.MAC)) {
		return false, errors.New("Codex capture control authentication failed")
	}
	status, err := s.CaptureControlForRoot(ctx, row.Identity)
	if err != nil {
		return false, err
	}
	// Retained envelopes authorized at an older revision are normal inbox
	// history. Never unlink a path another submitting process can replace.
	if status.Root != row.Identity || status.CaptureID != response.CaptureID ||
		status.Revision != response.Revision || controlAction(status.State) != response.Action {
		return false, nil
	}
	_, err = s.SetCaptureDiscarded(ctx, response.CaptureID, response.Action == "discard", &response.Revision)
	return err == nil, err
}

// ApplyOpenCodeCaptureControl accepts a trusted, source-checked plugin handoff.
// The exact root, action and revision are authenticated independently of bind.
func (s *Store) ApplyOpenCodeCaptureControl(ctx context.Context, id domain.CaptureIdentity, captureID, action, request string) (CaptureControl, error) {
	if id.Source != domain.SourceOpenCode || !validControlDigest(captureID) || (action != "discard" && action != "restore") {
		return CaptureControl{}, errors.New("capture control requires an exact native OpenCode root, capture ID and operation")
	}
	id, err := canonicalIdentity(id)
	if err != nil {
		return CaptureControl{}, err
	}
	parent, pending, conflict := OpenCodeHandoffParent(ctx, s.configuration, s.project, id.Store, id.SessionID)
	if err := errors.Join(pending, conflict); err != nil {
		return CaptureControl{}, fmt.Errorf("verify native OpenCode capture control scope: %w", err)
	}
	var body []byte
	if err := s.db.QueryRowContext(ctx, `SELECT record FROM registrations WHERE worktree_id=? AND identity=?`, s.worktreeID, IdentityKey(id)).Scan(&body); err != nil {
		return CaptureControl{}, err
	}
	row, err := decode(body)
	if err != nil {
		return CaptureControl{}, err
	}
	if row.ParentID != parent {
		return CaptureControl{}, errors.New("native OpenCode parent changed; obtain a current capture control request")
	}
	offer, err := s.captureControlInvitation(ctx, row, false)
	if err != nil {
		return CaptureControl{}, err
	}
	if offer == nil || offer.CaptureID != captureID || offer.Action != action || !hmac.Equal([]byte(offer.Request), []byte(request)) {
		return CaptureControl{}, errors.New("capture control requires the current request for the verified OpenCode root and operation")
	}
	_, _, revision, err := parseControlRequest(request)
	if err != nil {
		return CaptureControl{}, err
	}
	return s.SetCaptureDiscarded(ctx, captureID, action == "discard", &revision)
}
