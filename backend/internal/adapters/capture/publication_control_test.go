package capture_test

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"agenticlens/internal/adapters/capture"
)

func TestPublicationControlEmptyRegistryAndTicketRotation(t *testing.T) {
	ctx := context.Background()
	s := openStore(t, t.TempDir())
	request, err := s.PublicationRequest(ctx)
	if err != nil || request.Generation != 0 || request.Acknowledged != 0 {
		t.Fatalf("unused registry: %+v %v", request, err)
	}
	if err := s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	before, err := s.PublicationRequest(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.RequestPublication(ctx); err != nil {
		t.Fatal(err)
	}
	if err := s.AcknowledgePublication(ctx, before.Generation); err != nil {
		t.Fatal(err)
	}
	after, err := s.PublicationRequest(ctx)
	if err != nil || after.Generation == after.Acknowledged || after.Acknowledged != before.Generation {
		t.Fatalf("lost inflight request: %+v %v", after, err)
	}
	ordered, err := s.PublicationTickets(ctx, []string{"A", "B"})
	if err != nil || !reflect.DeepEqual(ordered, []string{"A", "B"}) {
		t.Fatalf("initial order %v %v", ordered, err)
	}
	if err := s.AttemptPublicationTicket(ctx, "A"); err != nil {
		t.Fatal(err)
	}
	ordered, err = s.PublicationTickets(ctx, []string{"A", "B"})
	if err != nil || !reflect.DeepEqual(ordered, []string{"B", "A"}) {
		t.Fatalf("starved B: %v %v", ordered, err)
	}
}

func TestAutomaticDeactivationRetainsOtherIntegration(t *testing.T) {
	dir := t.TempDir()
	save := func(source string, active bool) {
		t.Helper()
		b, err := json.Marshal(capture.Manifest{Protocol: capture.ManifestProtocolVersion, Source: source, Active: active})
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(dir, "install-"+source+".json"), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	save("codex", false)
	if enabled, err := capture.AutomaticEnabled(dir); err != nil || enabled {
		t.Fatalf("disabled integration active: %v %v", enabled, err)
	}
	save("opencode", true)
	if enabled, err := capture.AutomaticEnabled(dir); err != nil || !enabled {
		t.Fatalf("other integration stopped: %v %v", enabled, err)
	}
	save("opencode", false)
	if enabled, err := capture.AutomaticEnabled(dir); err != nil || enabled {
		t.Fatalf("disabled integrations active: %v %v", enabled, err)
	}
}
