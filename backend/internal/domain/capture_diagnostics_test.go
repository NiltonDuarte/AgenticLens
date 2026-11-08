package domain

import (
	"strings"
	"testing"
)

func diagnosticOwner(id, at string) SessionFact {
	return SessionFact{ID: id, Source: SourceCodex, Invocations: []InvocationFact{{CompletedAt: &at}}}
}

func TestCaptureAdvisoryRequiresConcreteEvidenceAndExactTail(t *testing.T) {
	id := CaptureIdentity{Source: SourceCodex, Store: "/configured-store", SessionID: "root"}
	published := diagnosticOwner("root", "2026-01-01T10:00:00Z")
	row := CaptureRegistration{Identity: id, TicketID: "TEST-100", State: CapturePending, Activity: "busy", Snapshot: []SessionFact{published}, RefreshedAt: "2026-01-01T10:10:00Z"}
	base := CapturePublicationEvidence{TicketID: "TEST-100", Available: true, Sessions: []SessionFact{published}, PublishedAt: "2026-01-01T10:05:00Z"}
	tests := []struct {
		name string
		edit func(*CaptureAdvisoryEvidence)
		want string
	}{
		{"pending active without caller", func(e *CaptureAdvisoryEvidence) {}, ""},
		{"missing metrics", func(e *CaptureAdvisoryEvidence) { e.Publications[0].Available = false }, "no readable metrics"},
		{"absent owner", func(e *CaptureAdvisoryEvidence) { e.Publications[0].Sessions = nil }, "absent from"},
		{"unbound independent reviewer", func(e *CaptureAdvisoryEvidence) {
			e.Registrations = append(e.Registrations, CaptureRegistration{Identity: CaptureIdentity{SessionID: "reviewer"}, ProjectProof: CaptureProjectVerified})
		}, "unbound"},
		{"failed source", func(e *CaptureAdvisoryEvidence) { e.Registrations[0].State = CaptureFailed; e.Caller = &id }, "source refresh failed"},
		{"tail exact caller", func(e *CaptureAdvisoryEvidence) {
			e.Registrations[0].Snapshot = []SessionFact{diagnosticOwner("root", "2026-01-01T10:06:00Z")}
			e.Caller = &id
		}, ""},
		{"tail unknown caller with newer known data", func(e *CaptureAdvisoryEvidence) {
			e.Registrations[0].Snapshot = []SessionFact{diagnosticOwner("root", "2026-01-01T10:06:00Z")}
		}, "newer than"},
		{"wrong caller store", func(e *CaptureAdvisoryEvidence) {
			e.Registrations[0].Snapshot = []SessionFact{diagnosticOwner("root", "2026-01-01T10:06:00Z")}
			other := id
			other.Store = "/other-store"
			e.Caller = &other
		}, "newer than"},
		{"prepublication omitted data", func(e *CaptureAdvisoryEvidence) {
			e.Registrations[0].Snapshot = []SessionFact{diagnosticOwner("root", "2026-01-01T10:03:00Z")}
			e.Caller = &id
		}, "newer than"},
		{"failed publication", func(e *CaptureAdvisoryEvidence) {
			e.Diagnostics = []CaptureDiagnostic{{Sequence: 1, Phase: "publication", Outcome: "failed", TicketID: "TEST-100"}}
		}, "unresolved operational"},
		{"recovered publication", func(e *CaptureAdvisoryEvidence) {
			e.Diagnostics = []CaptureDiagnostic{{Sequence: 1, Phase: "publication", Outcome: "failed", TicketID: "TEST-100"}, {Sequence: 2, Phase: "publication", Outcome: "committed", TicketID: "TEST-100"}}
		}, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			e := CaptureAdvisoryEvidence{Registrations: []CaptureRegistration{row}, Publications: []CapturePublicationEvidence{base}}
			tt.edit(&e)
			issues := strings.Join(CaptureAdvisoryIssues(e), " ")
			if tt.want == "" && issues != "" {
				t.Fatalf("unexpected warning: %s", issues)
			}
			if tt.want != "" && !strings.Contains(issues, tt.want) {
				t.Fatalf("want %q in %q", tt.want, issues)
			}
		})
	}
}

func TestCaptureVerificationRequiresSameCurrentInstallationSession(t *testing.T) {
	id := CaptureIdentity{Source: SourceCodex, Store: "/store", SessionID: "root"}
	row := CaptureRegistration{Identity: id, TicketID: "TEST-100", InstallationID: "install-current", BindingInstallationID: "install-current", SourceInstallationID: "install-current", CallbackObservedAt: "2026-01-01T10:00:00Z", Binding: "explicit_agent_response", RefreshedAt: "2026-01-01T10:02:00Z", Snapshot: []SessionFact{diagnosticOwner("root", "2026-01-01T10:01:00Z")}}
	committed := CaptureDiagnostic{Sequence: 1, TicketID: "TEST-100", Phase: "publication", Outcome: "committed", Owners: []CapturePublishedOwner{{Source: id.Source, Store: id.Store, SessionID: id.SessionID, InstallationID: row.InstallationID}}}
	for _, tt := range []struct {
		name string
		edit func(*CaptureRegistration, *CaptureDiagnostic)
		want string
	}{
		{"complete exchange", func(r *CaptureRegistration, d *CaptureDiagnostic) {}, "verified"},
		{"native exchange without installation receipt", func(r *CaptureRegistration, d *CaptureDiagnostic) {
			r.InstallationID = ""
			r.BindingInstallationID = ""
			r.SourceInstallationID = ""
			r.BindingAcceptedAt = "2026-01-01T10:01:00Z"
			d.Owners = []CapturePublishedOwner{{Source: id.Source, Store: id.Store, SessionID: id.SessionID}}
		}, "verified"},
		{"native binding missing accepted response", func(r *CaptureRegistration, d *CaptureDiagnostic) {
			r.InstallationID = ""
			r.BindingInstallationID = ""
			r.SourceInstallationID = ""
			d.Owners = []CapturePublishedOwner{{Source: id.Source, Store: id.Store, SessionID: id.SessionID}}
		}, "verification_pending"},
		{"callback absent", func(r *CaptureRegistration, d *CaptureDiagnostic) { r.CallbackObservedAt = "" }, "verification_pending"},
		{"previous installation binding", func(r *CaptureRegistration, d *CaptureDiagnostic) { r.BindingInstallationID = "install-old" }, "verification_pending"},
		{"previous installation source", func(r *CaptureRegistration, d *CaptureDiagnostic) { r.SourceInstallationID = "install-old" }, "verification_pending"},
		{"different store publication", func(r *CaptureRegistration, d *CaptureDiagnostic) {
			d.Owners = []CapturePublishedOwner{{Source: id.Source, Store: "/other", SessionID: id.SessionID, InstallationID: r.InstallationID}}
		}, "verification_pending"},
		{"unfinished publication", func(r *CaptureRegistration, d *CaptureDiagnostic) { d.Outcome = "staging" }, "degraded"},
	} {
		t.Run(tt.name, func(t *testing.T) {
			r, d := row, committed
			tt.edit(&r, &d)
			h := CaptureVerification([]CaptureRegistration{r}, []CaptureDiagnostic{d})
			if h.State != tt.want {
				t.Fatalf("health=%+v", h)
			}
		})
	}
	first, second := row, row
	first.Binding = ""
	first.Snapshot = nil
	second.CallbackObservedAt = ""
	if h := CaptureVerification([]CaptureRegistration{first, second}, []CaptureDiagnostic{committed}); h.State == "verified" {
		t.Fatalf("joined unrelated exchange evidence: %+v", h)
	}
}

func TestCaptureFailureFeedbackRequiresRelevantTerminalEvidence(t *testing.T) {
	identity := CaptureIdentity{Source: SourceCodex, Store: "/private/source", SessionID: "root"}
	row := CaptureRegistration{Identity: identity, TicketID: "TEST-100", State: CapturePending, Activity: "idle"}
	failed := CaptureDiagnostic{Sequence: 2, TicketID: "TEST-100", Phase: "publication", Outcome: "failed", Detail: "secret private-path other-ticket"}
	sourceFailed := CaptureDiagnostic{Sequence: 2, TicketID: "TEST-100", Phase: "reconciliation", Outcome: "failed", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: identity, State: CaptureFailed}}}}
	for _, scenario := range []struct {
		name        string
		diagnostics []CaptureDiagnostic
		want        string
	}{
		{"pending alone", []CaptureDiagnostic{{Sequence: 1, TicketID: "TEST-100", Phase: "publication", Outcome: "staging"}}, ""},
		{"failed publication", []CaptureDiagnostic{failed}, "publication failed"},
		{"inflight retry does not acknowledge failure", []CaptureDiagnostic{failed, {Sequence: 3, TicketID: "TEST-100", Phase: "publication", Outcome: "preparing"}}, "publication failed"},
		{"new committed publication supersedes history", []CaptureDiagnostic{failed, {Sequence: 3, TicketID: "TEST-100", Phase: "publication", Outcome: "committed"}}, ""},
		{"other ticket", []CaptureDiagnostic{{Sequence: 3, TicketID: "TEST-200", Phase: "publication", Outcome: "failed"}}, ""},
		{"exact source failure", []CaptureDiagnostic{sourceFailed}, "source reconciliation failed"},
		{"pending retry not accepted source success", []CaptureDiagnostic{sourceFailed, {Sequence: 3, TicketID: "TEST-100", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: identity, State: CapturePending}}}}}, "source reconciliation failed"},
		{"new exact successful source read", []CaptureDiagnostic{sourceFailed, {Sequence: 3, TicketID: "TEST-100", Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: identity, State: CaptureCurrent, RefreshedAt: "2026-01-01T10:00:00Z"}}}}}, ""},
		{"other exact source identity", []CaptureDiagnostic{{Sequence: 3, TicketID: "TEST-100", Phase: "reconciliation", Outcome: "failed", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: CaptureIdentity{Source: SourceCodex, Store: "/other-store", SessionID: "root"}, State: CaptureFailed}}}}}, ""},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			feedback := CaptureFailureFeedback(row, scenario.diagnostics)
			if scenario.want == "" && feedback != "" {
				t.Fatalf("unexpected feedback: %s", feedback)
			}
			if scenario.want != "" && !strings.Contains(feedback, scenario.want) {
				t.Fatalf("missing confirmed failure: %s", feedback)
			}
			for _, private := range []string{identity.Store, identity.SessionID, "secret", "other-ticket"} {
				if strings.Contains(feedback, private) {
					t.Fatalf("private diagnostic evidence leaked: %s", feedback)
				}
			}
		})
	}
	unbound := row
	unbound.TicketID = ""
	if feedback := CaptureFailureFeedback(unbound, []CaptureDiagnostic{failed}); feedback != "" {
		t.Fatalf("unbound session inherited collection failure: %s", feedback)
	}
}

func TestCaptureHealthRetainsExactSourceFailureThroughPendingRetry(t *testing.T) {
	id := CaptureIdentity{Source: SourceCodex, Store: "/store", SessionID: "root"}
	row := CaptureRegistration{Identity: id, TicketID: "TEST-100", State: CapturePending, InstallationID: "current", BindingInstallationID: "current", SourceInstallationID: "current", CallbackObservedAt: "2026-01-01T10:00:00Z", Binding: "explicit_agent_response", RefreshedAt: "2026-01-01T10:02:00Z", Snapshot: []SessionFact{diagnosticOwner("root", "2026-01-01T10:01:00Z")}}
	committed := CaptureDiagnostic{Sequence: 1, TicketID: row.TicketID, Phase: "publication", Outcome: "committed", Owners: []CapturePublishedOwner{{Source: id.Source, Store: id.Store, SessionID: id.SessionID, InstallationID: row.InstallationID}}}
	failure := CaptureDiagnostic{Sequence: 2, ObservedAt: "2026-01-01T10:03:00Z", TicketID: row.TicketID, Phase: "reconciliation", Outcome: "partial", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: id, State: CaptureFailed}}}}
	pending := CaptureDiagnostic{Sequence: 3, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: id, State: CapturePending}}}}
	publication := CapturePublicationEvidence{TicketID: row.TicketID, Available: true, Sessions: row.Snapshot, PublishedAt: "2026-01-01T10:02:00Z"}
	otherID := id
	otherID.Store = "/other-store"
	otherSuccess := CaptureDiagnostic{Sequence: 4, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: otherID, State: CaptureCurrent, RefreshedAt: "2026-01-01T10:04:00Z"}}}}
	otherFailure := CaptureDiagnostic{Sequence: 5, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "failed", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: otherID, State: CaptureFailed}}}}

	for _, scenario := range []struct {
		name        string
		diagnostics []CaptureDiagnostic
		failed      bool
	}{
		{"ordinary pending", []CaptureDiagnostic{committed, pending}, false},
		{"retained partial failure", []CaptureDiagnostic{committed, failure, pending}, true},
		{"another owner success does not recover", []CaptureDiagnostic{committed, failure, pending, otherSuccess}, true},
		{"another owner failure is scoped", []CaptureDiagnostic{committed, pending, otherFailure}, false},
		{"exact successful read", []CaptureDiagnostic{committed, failure, pending, {Sequence: 4, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: id, State: CaptureCurrent, RefreshedAt: "2026-01-01T10:04:00Z"}}}}}, false},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			h := CaptureVerification([]CaptureRegistration{row}, scenario.diagnostics)
			if !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
				t.Fatalf("historical exchange facts lost: %+v", h)
			}
			if (h.State == "degraded") != scenario.failed {
				t.Fatalf("terminal health=%+v", h)
			}
			issues := CaptureAdvisoryIssues(CaptureAdvisoryEvidence{Registrations: []CaptureRegistration{row}, Diagnostics: scenario.diagnostics, Publications: []CapturePublicationEvidence{publication}, Caller: &id})
			if strings.Contains(strings.Join(issues, " "), "source refresh failed") != scenario.failed {
				t.Fatalf("advisory=%v", issues)
			}
		})
	}
}

func TestCaptureCollectionFailureRequiresCompletedWholeAttempt(t *testing.T) {
	id := CaptureIdentity{Source: SourceCodex, Store: "/store", SessionID: "root"}
	row := CaptureRegistration{Identity: id, TicketID: "TEST-100", State: CaptureCurrent, InstallationID: "current", BindingInstallationID: "current", SourceInstallationID: "current", CallbackObservedAt: "2026-01-01T10:00:00Z", Binding: "explicit_agent_response", RefreshedAt: "2026-01-01T10:05:00Z", Snapshot: []SessionFact{diagnosticOwner("root", "2026-01-01T10:01:00Z")}}
	committed := CaptureDiagnostic{Sequence: 1, TicketID: row.TicketID, Phase: "publication", Outcome: "committed", Owners: []CapturePublishedOwner{{Source: id.Source, Store: id.Store, SessionID: id.SessionID, InstallationID: row.InstallationID}}}
	fatal := CaptureDiagnostic{Sequence: 2, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "failed", ObservedAt: "2026-01-01T10:03:00Z", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptFailed, Owners: []CaptureRegistration{row}}}
	partial := CaptureDiagnostic{Sequence: 3, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{row}}}
	complete := partial
	complete.Sequence = 4
	complete.Reconciliation = &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptCompleted, Owners: []CaptureRegistration{row}}
	legacy := fatal
	legacy.Reconciliation = &CaptureReconciliationDiagnostic{Owners: []CaptureRegistration{row}}
	otherComplete := complete
	otherComplete.TicketID = "TEST-200"
	otherID := id
	otherID.Store = "/other-store"
	otherFailure := partial
	otherFailure.Outcome = "failed"
	otherFailure.Reconciliation = &CaptureReconciliationDiagnostic{Attempt: CaptureAttemptPartial, Owners: []CaptureRegistration{{Identity: otherID, State: CaptureFailed}}}
	for _, scenario := range []struct {
		name                string
		diagnostics         []CaptureDiagnostic
		state               string
		collection, unknown bool
	}{
		{"fatal despite successful reads", []CaptureDiagnostic{committed, fatal}, "degraded", true, false},
		{"partial fresh read cannot clear fatal", []CaptureDiagnostic{committed, fatal, partial}, "degraded", true, false},
		{"other ticket completion cannot clear fatal", []CaptureDiagnostic{committed, fatal, otherComplete}, "degraded", true, false},
		{"whole completion recovers", []CaptureDiagnostic{committed, fatal, partial, complete}, "verified", false, false},
		{"unrelated owner failure remains isolated", []CaptureDiagnostic{committed, otherFailure}, "verified", false, false},
		{"legacy attempt unknown", []CaptureDiagnostic{committed, legacy}, "verification_pending", false, true},
		{"legacy cannot clear typed fatal", []CaptureDiagnostic{committed, fatal, {Sequence: 3, TicketID: row.TicketID, Phase: "reconciliation", Outcome: "accepted", Reconciliation: &CaptureReconciliationDiagnostic{Owners: []CaptureRegistration{row}}}}, "degraded", true, true},
	} {
		t.Run(scenario.name, func(t *testing.T) {
			h := CaptureVerification([]CaptureRegistration{row}, scenario.diagnostics)
			if h.State != scenario.state || !h.CallbackObserved || !h.BindingAccepted || !h.SourceAccepted || !h.PublicationAccepted {
				t.Fatalf("whole attempt health=%+v", h)
			}
			f := CaptureCurrentFailures(row, scenario.diagnostics)
			if f.Collection != scenario.collection || f.AttemptUnknown != scenario.unknown || f.Source {
				t.Fatalf("owner/collection conflated: %+v", f)
			}
			feedback := CaptureFailureFeedback(row, scenario.diagnostics)
			if strings.Contains(feedback, "whole-collection reconciliation failed") != scenario.collection {
				t.Fatalf("whole feedback=%q", feedback)
			}
			if scenario.unknown && !scenario.collection && !strings.Contains(feedback, "incomplete") {
				t.Fatalf("legacy false certainty: %q", feedback)
			}
		})
	}
}
