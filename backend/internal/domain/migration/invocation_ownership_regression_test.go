package migration_test

import (
	"encoding/json"
	"fmt"
	"reflect"
	"testing"

	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

func TestMigrationPreservesRepeatedRowsWithinLegacySessionOwners(t *testing.T) {
	rootID := "opaque-root"
	start, end := "2026-10-28T10:00:00Z", "2026-10-28T10:01:00Z"
	row := domain.InvocationFact{
		Role: "coder", Model: "model-a", StartedAt: &start, CompletedAt: &end,
		Turns: 3, UserIterations: 2, Tokens: domain.InvocationTokens{Input: 20, Output: 5}, CostUSD: 0.125,
	}
	otherModel := row
	otherModel.Model = "model-b"
	want := []domain.SessionFact{
		{ID: rootID, Source: domain.SourceCodex, User: "alice", Invocations: []domain.InvocationFact{row, row, otherModel}},
		{ID: "opaque-child", ParentID: &rootID, Source: domain.SourceCodex, User: "bob", Invocations: []domain.InvocationFact{row}},
	}

	type legacyRow struct {
		domain.InvocationFact
		SessionID string  `json:"id"`
		ParentID  *string `json:"parent_id"`
		User      string  `json:"user"`
	}
	for _, shape := range []string{"flat", "sessions"} {
		for version := 1; version <= domain.CurrentSchemaVersion; version++ {
			t.Run(fmt.Sprintf("%s/v%d", shape, version), func(t *testing.T) {
				var input []byte
				var err error
				if shape == "flat" {
					// Interleave two owners and repeat an indistinguishable row.
					input, err = json.Marshal(struct {
						SchemaVersion int                  `json:"schema_version"`
						Source        domain.SourceHarness `json:"source"`
						Invocations   []legacyRow          `json:"invocations"`
					}{version, domain.SourceCodex, []legacyRow{
						{row, rootID, nil, "alice"},
						{row, "opaque-child", &rootID, "bob"},
						{row, rootID, nil, "alice"},
						{otherModel, rootID, nil, "alice"},
					}})
				} else {
					input, err = json.Marshal(domain.AgenticFactLedger{SchemaVersion: &version, Sessions: want})
				}
				if err != nil {
					t.Fatal(err)
				}
				migrated, err := migration.NewEngine(nil).Migrate(input, domain.CurrentSchemaVersion)
				if err != nil {
					t.Fatal(err)
				}
				if migrated.SchemaVersion == nil || *migrated.SchemaVersion != domain.CurrentSchemaVersion || len(migrated.Sessions) != len(want) {
					t.Fatalf("migration lost the schema or owners: %+v", migrated)
				}
				for i, owner := range want {
					got := migrated.Sessions[i]
					if got.ID != owner.ID || !reflect.DeepEqual(got.ParentID, owner.ParentID) || got.Source != owner.Source || got.User != owner.User || !reflect.DeepEqual(got.Invocations, owner.Invocations) {
						t.Fatalf("owner %d metadata or repeated metric rows changed: got %+v, want %+v", i, got, owner)
					}
				}
				serialized, err := json.Marshal(migrated)
				if err != nil {
					t.Fatal(err)
				}
				var persisted struct {
					Sessions []struct {
						Invocations []map[string]json.RawMessage `json:"invocations"`
					} `json:"sessions"`
				}
				if err := json.Unmarshal(serialized, &persisted); err != nil {
					t.Fatal(err)
				}
				for si, session := range persisted.Sessions {
					for ii, invocation := range session.Invocations {
						if _, exists := invocation["id"]; exists {
							t.Fatalf("migrated owner %d row %d retained the legacy key", si, ii)
						}
					}
				}
				if got := domain.ComputeTotals(migrated.InvocationFacts()); got.Invocations != 4 || got.InputTokens != 80 || got.CostUSD != 0.5 {
					t.Fatalf("migration collapsed or duplicated metric contributions: %+v", got)
				}
			})
		}
	}
}
