package domain

import (
	"reflect"
	"testing"
)

func TestEvaluateMigrationStatus(t *testing.T) {
	tests := []struct {
		name     string
		input    MigrationEvaluationInput
		expected MigrationInfo
	}{
		{
			name: "initial full migration succeeded on machine A",
			input: MigrationEvaluationInput{
				ExistingMigration:  nil,
				RecomputeGit:       true,
				GitSucceeded:       true,
				RecomputeDB:        true,
				DBAttempted:        true,
				DBSessionFound:     true,
				MigratedBy:         "user-a",
				MigratedAt:         "2026-10-17T12:00:00Z",
				HasExistingDBFacts: false,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusCompleted,
				StepsCompleted: []string{"git_code_facts", "opencode_db"},
				StepsPending:   nil,
				LastMigratedAt: "2026-10-17T12:00:00Z",
				LastMigratedBy: "user-a",
				Notes:          "",
			},
		},
		{
			name: "machine A runs backfill for ticket authored on machine B (partial)",
			input: MigrationEvaluationInput{
				ExistingMigration:  nil,
				RecomputeGit:       true,
				GitSucceeded:       true,
				RecomputeDB:        true,
				DBAttempted:        true,
				DBSessionFound:     false,
				MigratedBy:         "user-a",
				MigratedAt:         "2026-10-17T12:00:00Z",
				HasExistingDBFacts: true,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusPartial,
				StepsCompleted: []string{"git_code_facts"},
				StepsPending:   []string{"opencode_db"},
				LastMigratedAt: "2026-10-17T12:00:00Z",
				LastMigratedBy: "user-a",
				Notes:          "Local SQLite session not found; preserved existing invocations and telemetry for subsequent host migration",
			},
		},
		{
			name: "machine B resumes partial migration and completes it",
			input: MigrationEvaluationInput{
				ExistingMigration: &MigrationInfo{
					SchemaVersion:  2,
					Status:         MigrationStatusPartial,
					StepsCompleted: []string{"git_code_facts"},
					StepsPending:   []string{"opencode_db"},
					LastMigratedAt: "2026-10-17T12:00:00Z",
					LastMigratedBy: "user-a",
					Notes:          "Local SQLite session not found; preserved existing invocations and telemetry for subsequent host migration",
				},
				RecomputeGit:       false,
				GitSucceeded:       false,
				RecomputeDB:        true,
				DBAttempted:        true,
				DBSessionFound:     true,
				MigratedBy:         "user-b",
				MigratedAt:         "2026-10-17T14:00:00Z",
				HasExistingDBFacts: true,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusCompleted,
				StepsCompleted: []string{"git_code_facts", "opencode_db"},
				StepsPending:   nil,
				LastMigratedAt: "2026-10-17T14:00:00Z",
				LastMigratedBy: "user-b",
				Notes:          "",
			},
		},
		{
			name: "machine B resumes partial migration but session still not found (remains partial)",
			input: MigrationEvaluationInput{
				ExistingMigration: &MigrationInfo{
					SchemaVersion:  2,
					Status:         MigrationStatusPartial,
					StepsCompleted: []string{"git_code_facts"},
					StepsPending:   []string{"opencode_db"},
					LastMigratedAt: "2026-10-17T12:00:00Z",
					LastMigratedBy: "user-a",
				},
				RecomputeGit:       false,
				GitSucceeded:       false,
				RecomputeDB:        true,
				DBAttempted:        true,
				DBSessionFound:     false,
				MigratedBy:         "user-b",
				MigratedAt:         "2026-10-17T14:00:00Z",
				HasExistingDBFacts: true,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusPartial,
				StepsCompleted: []string{"git_code_facts"},
				StepsPending:   []string{"opencode_db"},
				LastMigratedAt: "2026-10-17T14:00:00Z",
				LastMigratedBy: "user-b",
				Notes:          "Local SQLite session not found; preserved existing invocations and telemetry for subsequent host migration",
			},
		},
		{
			name: "git only recompute on unmigrated file",
			input: MigrationEvaluationInput{
				ExistingMigration:  nil,
				RecomputeGit:       true,
				GitSucceeded:       true,
				RecomputeDB:        false,
				DBAttempted:        false,
				DBSessionFound:     false,
				MigratedBy:         "user-a",
				MigratedAt:         "2026-10-17T15:00:00Z",
				HasExistingDBFacts: true,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusCompleted,
				StepsCompleted: []string{"git_code_facts"},
				StepsPending:   nil,
				LastMigratedAt: "2026-10-17T15:00:00Z",
				LastMigratedBy: "user-a",
				Notes:          "",
			},
		},
		{
			name: "empty inputs and unknown author fallback",
			input: MigrationEvaluationInput{
				ExistingMigration:  nil,
				RecomputeGit:       false,
				GitSucceeded:       false,
				RecomputeDB:        true,
				DBAttempted:        true,
				DBSessionFound:     false,
				MigratedBy:         "",
				MigratedAt:         "",
				HasExistingDBFacts: false,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusPending,
				StepsCompleted: nil,
				StepsPending:   []string{"opencode_db"},
				LastMigratedAt: "",
				LastMigratedBy: "unknown",
				Notes:          "Local SQLite session not found; pending database resolution on author host",
			},
		},
		{
			name: "git recompute requested but failed records pending git_code_facts",
			input: MigrationEvaluationInput{
				ExistingMigration:  nil,
				RecomputeGit:       true,
				GitSucceeded:       false,
				RecomputeDB:        false,
				DBAttempted:        false,
				DBSessionFound:     false,
				MigratedBy:         "tester",
				MigratedAt:         "2026-10-17T16:00:00Z",
				HasExistingDBFacts: false,
			},
			expected: MigrationInfo{
				SchemaVersion:  2,
				Status:         MigrationStatusPending,
				StepsCompleted: nil,
				StepsPending:   []string{"git_code_facts"},
				LastMigratedAt: "2026-10-17T16:00:00Z",
				LastMigratedBy: "tester",
				Notes:          "",
			},
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := EvaluateMigrationStatus(tc.input)
			if actual.SchemaVersion != tc.expected.SchemaVersion {
				t.Errorf("expected SchemaVersion %d, got %d", tc.expected.SchemaVersion, actual.SchemaVersion)
			}
			if actual.Status != tc.expected.Status {
				t.Errorf("expected Status %s, got %s", tc.expected.Status, actual.Status)
			}
			if !reflect.DeepEqual(actual.StepsCompleted, tc.expected.StepsCompleted) {
				t.Errorf("expected StepsCompleted %v, got %v", tc.expected.StepsCompleted, actual.StepsCompleted)
			}
			if len(actual.StepsPending) == 0 && len(tc.expected.StepsPending) == 0 {
				// match
			} else if !reflect.DeepEqual(actual.StepsPending, tc.expected.StepsPending) {
				t.Errorf("expected StepsPending %v, got %v", tc.expected.StepsPending, actual.StepsPending)
			}
			if actual.LastMigratedAt != tc.expected.LastMigratedAt {
				t.Errorf("expected LastMigratedAt %s, got %s", tc.expected.LastMigratedAt, actual.LastMigratedAt)
			}
			if actual.LastMigratedBy != tc.expected.LastMigratedBy {
				t.Errorf("expected LastMigratedBy %s, got %s", tc.expected.LastMigratedBy, actual.LastMigratedBy)
			}
			if actual.Notes != tc.expected.Notes {
				t.Errorf("expected Notes %q, got %q", tc.expected.Notes, actual.Notes)
			}
		})
	}
}
