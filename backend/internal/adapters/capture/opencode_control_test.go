package capture

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func openCodeControlFixture(t *testing.T) (*Store, *sql.DB, func(string) domain.CaptureRegistration) {
	t.Helper()
	ctx := context.Background()
	project, state := t.TempDir(), t.TempDir()
	path := filepath.Join(t.TempDir(), "native.db")
	shared, err := fixtureSharedYAML(t, "OpenCode capture control", "records")
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.yaml"), shared, 0600); err != nil {
		t.Fatal(err)
	}
	local := fmt.Sprintf("telemetry:\n  default_source: opencode\n  sources:\n    opencode:\n      database_path: %q\n  capture:\n    enabled: true\n    state_dir: %q\n    max_sessions: 16\n    max_source_bytes: 1048576\n", path, state)
	if err := os.WriteFile(filepath.Join(project, ".agenticlens-config.local.yaml"), []byte(local), 0600); err != nil {
		t.Fatal(err)
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	})
	if _, err := db.Exec(`CREATE TABLE session(id TEXT PRIMARY KEY,parent_id TEXT,directory TEXT,agent TEXT,model TEXT,title TEXT,time_created INTEGER,time_updated INTEGER);
CREATE TABLE message(id TEXT PRIMARY KEY,session_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);
CREATE TABLE part(id TEXT PRIMARY KEY,session_id TEXT,message_id TEXT,time_created INTEGER,time_updated INTEGER,data TEXT);`); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(`INSERT INTO session VALUES ('root',NULL,?,'build','fixture','Root',1000,3000),
('child','root',?,'build','fixture','Native child',1100,3000),
('late-child',NULL,?,'build','fixture','Late verified child',1200,3000)`, project, project, project); err != nil {
		t.Fatal(err)
	}
	store, err := OpenProject(state, project)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Error(err)
		}
	})
	return store, db, func(id string) domain.CaptureRegistration {
		t.Helper()
		identity := domain.CaptureIdentity{Source: domain.SourceOpenCode, Store: path, SessionID: id}
		parent, pending, conflict := OpenCodeHandoffParent(ctx, store.configuration, project, path, id)
		if err := errors.Join(pending, conflict); err != nil {
			t.Fatal(err)
		}
		row, err := store.Register(ctx, identity, "", "session.created")
		if err != nil {
			t.Fatal(err)
		}
		row, err = store.InheritOpenCodeParent(ctx, row, parent)
		if err != nil {
			t.Fatal(err)
		}
		return row
	}
}

func TestOpenCodeCaptureControlKeepsBindingCredentialsSeparate(t *testing.T) {
	ctx := context.Background()
	store, _, observe := openCodeControlFixture(t)
	root := observe("root")
	active, err := store.CaptureControlForRoot(ctx, root.Identity)
	if err != nil || active.State != CaptureControlActive || root.Request == "" {
		t.Fatalf("unbound native root did not provide its active capture and binding request: %v", err)
	}
	invitation, err := store.CaptureControlInvitation(ctx, root)
	if err != nil || invitation == nil || invitation.Action != "discard" || invitation.CaptureID != active.CaptureID {
		t.Fatalf("native root did not provide a discard invitation: %v", err)
	}
	// Check both the disclosed binding request and a correctly shaped request
	// scoped with the binding credential. Syntax alone cannot protect the action.
	bindingDerived := strings.Join([]string{controlInboxKey(root.Request, store.project, root, active.Revision),
		strconv.FormatInt(active.Revision, 10),
		controlRequestKey(root.Request, store.project, root, active.CaptureID, "discard", active.Revision)}, ".")
	for _, request := range []string{root.Request, bindingDerived} {
		if _, err := store.ApplyOpenCodeCaptureControl(ctx, root.Identity, active.CaptureID, "discard", request); err == nil || !strings.Contains(err.Error(), "current request") {
			t.Fatalf("binding credential was not rejected at the control authority boundary: %v", err)
		}
		current, err := store.CaptureControlForRoot(ctx, root.Identity)
		if err != nil || current.State != CaptureControlActive || current.Revision != active.Revision {
			t.Fatalf("rejected control changed native disposition: %s at %d, %v", current.State, current.Revision, err)
		}
	}
	discarded, err := store.ApplyOpenCodeCaptureControl(ctx, root.Identity, invitation.CaptureID, invitation.Action, invitation.Request)
	if err != nil || discarded.State != CaptureControlDiscarded || discarded.Revision <= active.Revision || !discarded.CleanupPending {
		t.Fatalf("authenticated native discard did not persist: %s at %d, %v", discarded.State, discarded.Revision, err)
	}
}

func TestOpenCodeCaptureControlRequiresCurrentIndependentRoot(t *testing.T) {
	ctx := context.Background()
	store, native, observe := openCodeControlFixture(t)
	parent, child := observe("root"), observe("child")
	if child.ParentID != parent.Identity.SessionID || child.TicketID != "" || parent.TicketID != "" {
		t.Fatalf("native handoff did not retain the exact unbound parent edge: %q", child.ParentID)
	}
	if offer, err := store.CaptureControlInvitation(ctx, child); err != nil || offer != nil {
		t.Fatalf("native child received a separate control invitation: %v", err)
	}

	late := observe("late-child")
	discard, err := store.CaptureControlInvitation(ctx, late)
	if err != nil || discard == nil || discard.Action != "discard" {
		t.Fatalf("native root without a known parent did not receive its initial invitation: %v", err)
	}
	discarded, err := store.ApplyOpenCodeCaptureControl(ctx, late.Identity, discard.CaptureID, discard.Action, discard.Request)
	if err != nil || discarded.State != CaptureControlDiscarded {
		t.Fatalf("initial discard failed: %v", err)
	}
	restore, err := store.CaptureControlInvitation(ctx, late)
	if err != nil || restore == nil || restore.Action != "restore" {
		t.Fatalf("discarded root did not receive a restore invitation: %v", err)
	}
	if _, err := native.Exec(`UPDATE session SET parent_id=? WHERE id=?`, parent.Identity.SessionID, late.Identity.SessionID); err != nil {
		t.Fatal(err)
	}
	if _, err := store.ApplyOpenCodeCaptureControl(ctx, late.Identity, restore.CaptureID, restore.Action, restore.Request); err == nil || !strings.Contains(err.Error(), "parent changed") {
		t.Fatalf("native source parent did not supersede the old control request: %v", err)
	}
	late = observe(late.Identity.SessionID)
	if late.ParentID != parent.Identity.SessionID || late.TicketID != "" {
		t.Fatalf("handoff failed to persist the newly verified unbound edge: %q", late.ParentID)
	}
	if offer, err := store.CaptureControlInvitation(ctx, late); err != nil || offer != nil {
		t.Fatalf("recorded child retained a separate invitation for its earlier disposition: %v", err)
	}
	if _, err := store.ApplyOpenCodeCaptureControl(ctx, late.Identity, restore.CaptureID, restore.Action, restore.Request); err == nil || !strings.Contains(err.Error(), "current request") {
		t.Fatalf("recorded child could reuse its earlier independent restore invitation: %v", err)
	}
	retained, err := store.CaptureControl(ctx, discard.CaptureID)
	if err != nil || retained.State != CaptureControlDiscarded || retained.Revision != discarded.Revision {
		t.Fatalf("rejected restore changed the durable control: %s at %d, %v", retained.State, retained.Revision, err)
	}

	if _, err := store.BindNativeOpenCode(ctx, parent.Identity, BindingTicket{ID: "T-100", Title: "Native parent", Type: "fix"}, parent.Request); err != nil {
		t.Fatal(err)
	}
	independent, err := store.BindNativeOpenCode(ctx, child.Identity, BindingTicket{ID: "T-200", Title: "Independent native child", Type: "fix"}, child.Request)
	if err != nil {
		t.Fatal(err)
	}
	offer, err := store.CaptureControlInvitation(ctx, independent)
	if err != nil || offer == nil || offer.Action != "discard" || offer.CaptureID != StableCaptureID(child.Identity.Source, child.Identity.SessionID) {
		t.Fatalf("independently bound native child did not receive its own invitation: %v", err)
	}
	controlled, err := store.ApplyOpenCodeCaptureControl(ctx, child.Identity, offer.CaptureID, offer.Action, offer.Request)
	if err != nil || controlled.State != CaptureControlDiscarded || controlled.TicketID != "T-200" {
		t.Fatalf("verified independent control did not preserve the child's binding: %v", err)
	}
	active, err := store.CaptureControlForRoot(ctx, parent.Identity)
	if err != nil || active.State != CaptureControlActive || active.TicketID != "T-100" {
		t.Fatalf("independent discard changed the native parent's disposition: %v", err)
	}
}
