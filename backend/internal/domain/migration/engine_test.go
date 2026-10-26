package migration_test

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"

	"agenticlens/internal/domain"
	"agenticlens/internal/domain/migration"
)

func stringPtr(s string) *string {
	return &s
}

func intPtr(i int) *int {
	return &i
}

func int64Ptr(i int64) *int64 {
	return &i
}

func ticketTypePtr(tt domain.ConventionalTicketType) *domain.ConventionalTicketType {
	return &tt
}

func sampleFactLedger() domain.AgenticFactLedger {
	return domain.AgenticFactLedger{
		TicketID:    stringPtr("AL-100"),
		TicketTitle: stringPtr("Sample Ticket"),
		TicketType:  ticketTypePtr(domain.TicketTypeFeat),
		StartedAt:   stringPtr("2026-10-18T10:00:00Z"),
		CompletedAt: stringPtr("2026-10-18T11:00:00Z"),
		UpdatedAt:   "2026-10-18T11:00:00Z",
		Code: domain.CodeFacts{
			LinesAdded:       100,
			LinesDeleted:     20,
			TestLinesAdded:   50,
			TestLinesDeleted: 0,
			ProdLinesAdded:   50,
			ProdLinesDeleted: 20,
			FilesChanged:     5,
			SDLCIterations:   2,
		}, Sessions: sessionFacts([]domain.InvocationFact{
			{
				ID:                    "inv-1",
				ParentID:              nil,
				Role:                  "orchestrator",
				Model:                 "claude-3-opus",
				User:                  "developer",
				StartedAt:             stringPtr("2026-10-18T10:00:00Z"),
				CompletedAt:           stringPtr("2026-10-18T10:30:00Z"),
				ActiveDurationSeconds: 1800,
				Turns:                 10,
				Tokens: domain.InvocationTokens{
					Input:       10000,
					Output:      2000,
					CacheRead:   5000,
					Reasoning:   1000,
					PeakContext: 50000,
				},
				CostUSD:     0.50,
				Compactions: 0,
			},
		},
			[]domain.ToolFact{
				{
					SessionID: "session-fixture",
					Name:      "bash",
					Calls:     15,
					Errors:    1,
				},
			},
			nil),
	}
}

func TestDetectVersion(t *testing.T) {
	engine := migration.NewEngine(nil)

	tests := []struct {
		name        string
		input       string
		expectedVer int
		wantErr     error
	}{
		{
			name:        "empty input",
			input:       "",
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "whitespace only input",
			input:       "   \n\t  \r\n ",
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "malformed json",
			input:       "{not-valid-json",
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "json array root instead of object",
			input:       `[{"schema_version": 2}]`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "json primitive string",
			input:       `"just a string"`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "json primitive number",
			input:       `12345`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "json primitive boolean",
			input:       `true`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "schema_version as string type",
			input:       `{"schema_version": "two", "ticket_id": "AL-10"}`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "schema_version as float type",
			input:       `{"schema_version": 2.5, "ticket_id": "AL-10"}`,
			expectedVer: 0,
			wantErr:     migration.ErrMalformedJSON,
		},
		{
			name:        "unversioned historical json",
			input:       `{"ticket_id": "AL-10", "code": {"lines_added": 10}}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "empty json object",
			input:       `{}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "negative schema_version defaults to 1",
			input:       `{"schema_version": -3, "ticket_id": "AL-10"}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "explicit null schema_version",
			input:       `{"schema_version": null, "ticket_id": "AL-10"}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "schema_version 0",
			input:       `{"schema_version": 0, "ticket_id": "AL-10"}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "schema_version 1",
			input:       `{"schema_version": 1, "ticket_id": "AL-10"}`,
			expectedVer: 1,
			wantErr:     nil,
		},
		{
			name:        "schema_version 2",
			input:       `{"schema_version": 2, "ticket_id": "AL-10"}`,
			expectedVer: 2,
			wantErr:     nil,
		},
		{
			name:        "schema_version 3",
			input:       `{"schema_version": 3, "ticket_id": "AL-10"}`,
			expectedVer: 3,
			wantErr:     nil,
		},
		{
			name:        "large schema_version 999999",
			input:       `{"schema_version": 999999, "ticket_id": "AL-10"}`,
			expectedVer: 999999,
			wantErr:     nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ver, err := engine.DetectVersion([]byte(tt.input))
			if tt.wantErr != nil {
				if err == nil {
					t.Fatalf("expected error %v, got nil", tt.wantErr)
				}
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("expected error wrapping %v, got %v", tt.wantErr, err)
				}
				return
			}
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ver != tt.expectedVer {
				t.Errorf("expected version %d, got %d", tt.expectedVer, ver)
			}
		})
	}
}

// mockStep implements MigrationStep for tests.
type mockStep struct {
	source int
	target int
	desc   string
	apply  func(domain.AgenticFactLedger) (domain.AgenticFactLedger, error)
}

func (m *mockStep) SourceVersion() int  { return m.source }
func (m *mockStep) TargetVersion() int  { return m.target }
func (m *mockStep) Description() string { return m.desc }
func (m *mockStep) Apply(l domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	if m.apply != nil {
		return m.apply(l)
	}
	ver := m.target
	l.SchemaVersion = &ver
	return l, nil
}

func TestRegistry(t *testing.T) {
	t.Run("valid registration and get", func(t *testing.T) {
		reg := migration.NewRegistry()
		step := &mockStep{source: 1, target: 2, desc: "v1 to v2"}
		if err := reg.Register(step); err != nil {
			t.Fatalf("expected successful registration, got %v", err)
		}

		retrieved, ok := reg.GetStep(1)
		if !ok || retrieved != step {
			t.Errorf("expected to retrieve registered step")
		}
	})

	t.Run("nil step rejected", func(t *testing.T) {
		reg := migration.NewRegistry()
		err := reg.Register(nil)
		if !errors.Is(err, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for nil step, got %v", err)
		}
	})

	t.Run("non-contiguous step rejected", func(t *testing.T) {
		reg := migration.NewRegistry()
		// target > source + 1
		stepSkip := &mockStep{source: 1, target: 3, desc: "skipping v2"}
		if err := reg.Register(stepSkip); !errors.Is(err, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for skipping step, got %v", err)
		}

		// target == source
		stepEqual := &mockStep{source: 2, target: 2, desc: "equal target"}
		if err := reg.Register(stepEqual); !errors.Is(err, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for equal step, got %v", err)
		}

		// target < source
		stepDowngrade := &mockStep{source: 3, target: 2, desc: "downgrade step"}
		if err := reg.Register(stepDowngrade); !errors.Is(err, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for downgrade step, got %v", err)
		}
	})

	t.Run("duplicate registration rejected", func(t *testing.T) {
		reg := migration.NewRegistry()
		step1 := &mockStep{source: 1, target: 2, desc: "first"}
		step2 := &mockStep{source: 1, target: 2, desc: "second"}
		if err := reg.Register(step1); err != nil {
			t.Fatal(err)
		}
		err := reg.Register(step2)
		if !errors.Is(err, migration.ErrDuplicateStep) {
			t.Errorf("expected ErrDuplicateStep for duplicate registration, got %v", err)
		}
	})

	t.Run("thread-safe concurrent registration and plan discovery", func(t *testing.T) {
		reg := migration.NewRegistry()
		// Pre-populate steps 1 to 5
		for i := 1; i <= 5; i++ {
			_ = reg.Register(&mockStep{source: i, target: i + 1, desc: fmt.Sprintf("%d to %d", i, i+1)})
		}

		var wg sync.WaitGroup
		concurrency := 40
		wg.Add(concurrency)

		for i := 0; i < concurrency; i++ {
			i := i
			go func() {
				defer wg.Done()
				if i%2 == 0 {
					// Register step for a higher version
					sourceVer := 10 + i
					_ = reg.Register(&mockStep{source: sourceVer, target: sourceVer + 1, desc: fmt.Sprintf("step-%d", sourceVer)})
				} else {
					// Query chain
					_, _ = reg.PlanChain(1, 4)
					_, _ = reg.GetStep(2)
				}
			}()
		}

		wg.Wait()
		chain, err := reg.PlanChain(1, 5)
		if err != nil || len(chain) != 4 {
			t.Errorf("expected valid 4-step chain after concurrent access, got len %d, err %v", len(chain), err)
		}
	})

	t.Run("plan chain scenarios", func(t *testing.T) {
		reg := migration.NewRegistry()
		s1 := &mockStep{source: 1, target: 2, desc: "1 to 2"}
		s2 := &mockStep{source: 2, target: 3, desc: "2 to 3"}
		_ = reg.Register(s1)
		_ = reg.Register(s2)

		// Same version
		chain, err := reg.PlanChain(2, 2)
		if err != nil || len(chain) != 0 {
			t.Errorf("expected empty chain for identical version, got %d steps, err %v", len(chain), err)
		}

		// Downgrade
		_, err = reg.PlanChain(2, 1)
		if !errors.Is(err, migration.ErrDowngradeNotAllowed) {
			t.Errorf("expected ErrDowngradeNotAllowed, got %v", err)
		}

		// Single step
		chain, err = reg.PlanChain(1, 2)
		if err != nil || len(chain) != 1 || chain[0] != s1 {
			t.Errorf("expected 1-step chain with s1, got %v, err %v", chain, err)
		}

		// Multi step
		chain, err = reg.PlanChain(1, 3)
		if err != nil || len(chain) != 2 || chain[0] != s1 || chain[1] != s2 {
			t.Errorf("expected 2-step chain with s1 and s2, got %v, err %v", chain, err)
		}

		// Missing transition in chain
		_, err = reg.PlanChain(1, 4)
		if !errors.Is(err, migration.ErrMissingStep) {
			t.Errorf("expected ErrMissingStep when step missing, got %v", err)
		}

		// Discontinuous chain: registered 1->2 and 3->4 (missing 2->3)
		regDiscontinuous := migration.NewRegistry()
		_ = regDiscontinuous.Register(s1)
		_ = regDiscontinuous.Register(&mockStep{source: 3, target: 4, desc: "3 to 4"})
		_, err = regDiscontinuous.PlanChain(1, 4)
		if !errors.Is(err, migration.ErrMissingStep) {
			t.Errorf("expected ErrMissingStep for gap in chain, got %v", err)
		}
	})
}

func TestStepV1ToV2(t *testing.T) {
	step := migration.NewStepV1ToV2()
	if step.SourceVersion() != 1 || step.TargetVersion() != 2 {
		t.Errorf("unexpected versions: source %d, target %d", step.SourceVersion(), step.TargetVersion())
	}
	if step.Description() == "" {
		t.Errorf("expected non-empty description")
	}

	t.Run("initializes schema version, empty events, and baseline migration info", func(t *testing.T) {
		initial := sampleFactLedger()
		initial.SchemaVersion = nil
		setEventFacts(&initial, nil)
		initial.Migration = nil

		upgraded, err := step.Apply(initial)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if upgraded.SchemaVersion == nil || *upgraded.SchemaVersion != 2 {
			t.Errorf("expected schema version 2, got %v", upgraded.SchemaVersion)
		}
		if upgraded.Sessions == nil {
			t.Errorf("expected non-nil sessions collection")
		}
		if len(upgraded.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got %d", len(upgraded.EventFacts()))
		}
		if upgraded.Migration == nil {
			t.Fatalf("expected non-nil migration info")
		}
		if upgraded.Migration.SchemaVersion != 2 {
			t.Errorf("expected migration schema version 2, got %d", upgraded.Migration.SchemaVersion)
		}
		if upgraded.Migration.LastMigratedBy != "universal-migration-engine" {
			t.Errorf("expected LastMigratedBy 'universal-migration-engine', got %s", upgraded.Migration.LastMigratedBy)
		}
		foundBaseline := false
		for _, s := range upgraded.Migration.StepsCompleted {
			if s == "schema_v2_baseline" {
				foundBaseline = true
				break
			}
		}
		if !foundBaseline {
			t.Errorf("expected 'schema_v2_baseline' in steps completed")
		}
	})

	t.Run("completely empty zero-value ledger initializes schema v2 cleanly", func(t *testing.T) {
		emptyLedger := domain.AgenticFactLedger{}
		upgraded, err := step.Apply(emptyLedger)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upgraded.SchemaVersion == nil || *upgraded.SchemaVersion != 2 {
			t.Errorf("expected schema version 2, got %v", upgraded.SchemaVersion)
		}
		if upgraded.Sessions == nil {
			t.Errorf("expected non-nil sessions collection")
		}
		if len(upgraded.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got %d", len(upgraded.EventFacts()))
		}
		if upgraded.Migration == nil || upgraded.Migration.SchemaVersion != 2 {
			t.Errorf("expected migration info initialized with v2")
		}
	})

	t.Run("idempotent application does not duplicate schema_v2_baseline in steps", func(t *testing.T) {
		initial := sampleFactLedger()
		firstPass, err := step.Apply(initial)
		if err != nil {
			t.Fatal(err)
		}
		secondPass, err := step.Apply(firstPass)
		if err != nil {
			t.Fatal(err)
		}

		baselineCount := 0
		for _, s := range secondPass.Migration.StepsCompleted {
			if s == "schema_v2_baseline" {
				baselineCount++
			}
		}
		if baselineCount != 1 {
			t.Errorf("expected exactly 1 'schema_v2_baseline' in steps completed, got %d", baselineCount)
		}
	})

	t.Run("preserves existing migration steps and events", func(t *testing.T) {
		initial := sampleFactLedger()
		setEventFacts(&initial, []domain.EventFact{
			{
				Timestamp: "2026-10-18T10:05:00Z",
				SessionID: "ses-1",
				Turn:      1,
				Role:      "coder",
				EventType: domain.EventTypeTurnMessage,
				Status:    "completed",
			},
		})
		initial.Migration = &domain.MigrationInfo{
			SchemaVersion:  1,
			Status:         domain.MigrationStatusPartial,
			StepsCompleted: []string{"git_code_facts"},
			StepsPending:   []string{"opencode_db"},
			LastMigratedAt: "2026-10-18T10:00:00Z",
			LastMigratedBy: "test-user",
		}

		upgraded, err := step.Apply(initial)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if len(upgraded.EventFacts()) != 1 {
			t.Errorf("expected 1 preserved event, got %d", len(upgraded.EventFacts()))
		}
		if upgraded.Migration.LastMigratedBy != "test-user" {
			t.Errorf("expected preserved author, got %s", upgraded.Migration.LastMigratedBy)
		}
		hasGit := false
		hasBaseline := false
		for _, s := range upgraded.Migration.StepsCompleted {
			if s == "git_code_facts" {
				hasGit = true
			}
			if s == "schema_v2_baseline" {
				hasBaseline = true
			}
		}
		if !hasGit || !hasBaseline {
			t.Errorf("expected both git_code_facts and schema_v2_baseline in steps completed, got %v", upgraded.Migration.StepsCompleted)
		}
	})
}

func TestInvariantValidator(t *testing.T) {
	validator := migration.NewInvariantValidator()

	base := sampleFactLedger()
	v2 := 2
	base.SchemaVersion = &v2
	setEventFacts(&base, []domain.EventFact{
		{
			Timestamp:     "2026-10-18T10:00:00Z",
			SessionID:     "ses-1",
			Turn:          1,
			Role:          "coder",
			EventType:     domain.EventTypeTurnMessage,
			ToolName:      nil,
			Status:        "completed",
			ErrorMessage:  nil,
			DurationMs:    int64Ptr(100),
			ContextTokens: 5000,
		},
	})

	t.Run("identical pre and post passes", func(t *testing.T) {
		err := validator.Validate(base, base, 2)
		if err != nil {
			t.Errorf("expected clean validation, got %v", err)
		}
	})

	t.Run("schema version nil in post fails", func(t *testing.T) {
		modified := base
		modified.SchemaVersion = nil
		err := validator.Validate(base, modified, 2)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when schema_version nil, got %v", err)
		}
	})

	t.Run("target version mismatch fails", func(t *testing.T) {
		err := validator.Validate(base, base, 3)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation, got %v", err)
		}
	})

	t.Run("ticket id mutations fail", func(t *testing.T) {
		// value changed
		m1 := base
		m1.TicketID = stringPtr("AL-999")
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed ticket_id, got %v", err)
		}

		// non-nil to nil
		m2 := base
		m2.TicketID = nil
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for cleared ticket_id, got %v", err)
		}

		// nil to non-nil
		b3 := base
		b3.TicketID = nil
		if err := validator.Validate(b3, base, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for newly set ticket_id, got %v", err)
		}
	})

	t.Run("ticket title mutations fail", func(t *testing.T) {
		m1 := base
		m1.TicketTitle = stringPtr("Different Title")
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed ticket_title, got %v", err)
		}

		m2 := base
		m2.TicketTitle = nil
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for cleared ticket_title, got %v", err)
		}
	})

	t.Run("ticket type mutations fail", func(t *testing.T) {
		m1 := base
		m1.TicketType = ticketTypePtr(domain.TicketTypeFix)
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed ticket_type, got %v", err)
		}

		m2 := base
		m2.TicketType = nil
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for cleared ticket_type, got %v", err)
		}
	})

	t.Run("started_at mutations fail", func(t *testing.T) {
		m1 := base
		m1.StartedAt = stringPtr("2026-01-01T00:00:00Z")
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed started_at, got %v", err)
		}

		m2 := base
		m2.StartedAt = nil
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for cleared started_at, got %v", err)
		}
	})

	t.Run("completed_at mutations fail", func(t *testing.T) {
		m1 := base
		m1.CompletedAt = stringPtr("2026-01-01T00:00:00Z")
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed completed_at, got %v", err)
		}

		m2 := base
		m2.CompletedAt = nil
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for cleared completed_at, got %v", err)
		}
	})

	t.Run("updated_at cleared fails", func(t *testing.T) {
		modified := base
		modified.UpdatedAt = ""
		err := validator.Validate(base, modified, 2)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation, got %v", err)
		}
	})

	t.Run("code facts individual mutations fail", func(t *testing.T) {
		cases := []struct {
			name   string
			mutate func(*domain.AgenticFactLedger)
		}{
			{"lines_added", func(l *domain.AgenticFactLedger) { l.Code.LinesAdded++ }},
			{"lines_deleted", func(l *domain.AgenticFactLedger) { l.Code.LinesDeleted++ }},
			{"test_lines_added", func(l *domain.AgenticFactLedger) { l.Code.TestLinesAdded++ }},
			{"test_lines_deleted", func(l *domain.AgenticFactLedger) { l.Code.TestLinesDeleted++ }},
			{"prod_lines_added", func(l *domain.AgenticFactLedger) { l.Code.ProdLinesAdded++ }},
			{"prod_lines_deleted", func(l *domain.AgenticFactLedger) { l.Code.ProdLinesDeleted++ }},
			{"files_changed", func(l *domain.AgenticFactLedger) { l.Code.FilesChanged++ }},
			{"sdlc_iterations", func(l *domain.AgenticFactLedger) { l.Code.SDLCIterations++ }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				m := base
				tc.mutate(&m)
				err := validator.Validate(base, m, 2)
				if !errors.Is(err, migration.ErrInvariantViolation) {
					t.Errorf("expected ErrInvariantViolation for mutated %s, got %v", tc.name, err)
				}
			})
		}
	})

	t.Run("invocations count changes fail", func(t *testing.T) {
		// dropped
		m1 := base
		setInvocationFacts(&m1, []domain.InvocationFact{})
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for dropped invocations, got %v", err)
		}

		// added
		m2 := base
		setInvocationFacts(&m2, append(m2.InvocationFacts(), m2.InvocationFacts()[0]))
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for added invocations, got %v", err)
		}
	})

	t.Run("invocation individual field mutations fail", func(t *testing.T) {
		baseWithParent := base
		parented := append([]domain.InvocationFact(nil), base.InvocationFacts()...)
		parented[0].ParentID = stringPtr("parent-1")
		setInvocationFacts(&baseWithParent, parented)

		cases := []struct {
			name   string
			mutate func(*domain.InvocationFact)
		}{
			{"id", func(i *domain.InvocationFact) { i.ID = "inv-diff" }},
			{"parent_id_changed", func(i *domain.InvocationFact) { i.ParentID = stringPtr("parent-2") }},
			{"parent_id_cleared", func(i *domain.InvocationFact) { i.ParentID = nil }},
			{"role", func(i *domain.InvocationFact) { i.Role = "reviewer" }},
			{"model", func(i *domain.InvocationFact) { i.Model = "gpt-4o" }},
			{"user", func(i *domain.InvocationFact) { i.User = "another-user" }},
			{"started_at", func(i *domain.InvocationFact) { i.StartedAt = stringPtr("2026-01-01T00:00:00Z") }},
			{"completed_at", func(i *domain.InvocationFact) { i.CompletedAt = stringPtr("2026-01-01T01:00:00Z") }},
			{"active_duration", func(i *domain.InvocationFact) { i.ActiveDurationSeconds = 9999 }},
			{"turns", func(i *domain.InvocationFact) { i.Turns = 99 }},
			{"cost_usd", func(i *domain.InvocationFact) { i.CostUSD = 9.99 }},
			{"compactions", func(i *domain.InvocationFact) { i.Compactions = 3 }},
			{"tokens_input", func(i *domain.InvocationFact) { i.Tokens.Input = 999999 }},
			{"tokens_output", func(i *domain.InvocationFact) { i.Tokens.Output = 999999 }},
			{"tokens_cache_read", func(i *domain.InvocationFact) { i.Tokens.CacheRead = 999999 }},
			{"tokens_reasoning", func(i *domain.InvocationFact) { i.Tokens.Reasoning = 999999 }},
			{"tokens_peak_context", func(i *domain.InvocationFact) { i.Tokens.PeakContext = 999999 }},
		}

		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				m := baseWithParent
				invs := append([]domain.InvocationFact(nil), baseWithParent.InvocationFacts()...)
				tc.mutate(&invs[0])
				setInvocationFacts(&m, invs)
				err := validator.Validate(baseWithParent, m, 2)
				if !errors.Is(err, migration.ErrInvariantViolation) {
					t.Errorf("expected ErrInvariantViolation for mutated %s, got %v", tc.name, err)
				}
			})
		}
	})

	t.Run("tools count and field mutations fail", func(t *testing.T) {
		// dropped
		m1 := base
		setToolFacts(&m1, []domain.ToolFact{})
		if err := validator.Validate(base, m1, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for dropped tools, got %v", err)
		}

		// added
		m2 := base
		setToolFacts(&m2, append(m2.ToolFacts(), domain.ToolFact{SessionID: "session-fixture", Name: "git", Calls: 1}))
		if err := validator.Validate(base, m2, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for added tools, got %v", err)
		}

		// name changed
		m3 := base
		tools := append([]domain.ToolFact(nil), base.ToolFacts()...)
		tools[0].Name = "zsh"
		setToolFacts(&m3, tools)
		if err := validator.Validate(base, m3, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed tool name, got %v", err)
		}

		// calls changed
		m4 := base
		tools = append([]domain.ToolFact(nil), base.ToolFacts()...)
		tools[0].Calls = 999
		setToolFacts(&m4, tools)
		if err := validator.Validate(base, m4, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed tool calls, got %v", err)
		}

		// errors changed
		m5 := base
		tools = append([]domain.ToolFact(nil), base.ToolFacts()...)
		tools[0].Errors = 999
		setToolFacts(&m5, tools)
		if err := validator.Validate(base, m5, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation for changed tool errors, got %v", err)
		}
	})

	t.Run("event invariant rules", func(t *testing.T) {
		// existing events dropped -> fail
		mDropped := base
		setEventFacts(&mDropped, []domain.EventFact{})
		if err := validator.Validate(base, mDropped, 2); !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when events dropped, got %v", err)
		}

		// new events appended -> allowed and passes!
		mAppended := base
		setEventFacts(&mAppended, append([]domain.EventFact(nil), base.EventFacts()...))
		setEventFacts(&mAppended, append(mAppended.EventFacts(), domain.EventFact{
			Timestamp: "2026-10-18T10:05:00Z",
			SessionID: "ses-1",
			Turn:      2,
			Role:      "coder",
			EventType: domain.EventTypeCompaction,
			Status:    "completed",
		}))
		if err := validator.Validate(base, mAppended, 2); err != nil {
			t.Errorf("expected appending events to pass invariant validation, got %v", err)
		}

		// pre-existing event mutations fail
		baseWithToolEvent := base
		setEventFacts(&baseWithToolEvent, []domain.EventFact{
			{
				Timestamp:     "2026-10-18T10:00:00Z",
				SessionID:     "ses-1",
				Turn:          1,
				Role:          "coder",
				EventType:     domain.EventTypeToolCall,
				ToolName:      stringPtr("bash"),
				Status:        "error",
				ErrorMessage:  stringPtr("command failed"),
				DurationMs:    int64Ptr(250),
				ContextTokens: 5000,
			},
		})

		eventCases := []struct {
			name   string
			mutate func(*domain.EventFact)
		}{
			{"timestamp", func(e *domain.EventFact) { e.Timestamp = "2026-01-01T00:00:00Z" }},
			{"session_id", func(e *domain.EventFact) { e.SessionID = "ses-diff" }},
			{"turn", func(e *domain.EventFact) { e.Turn = 99 }},
			{"role", func(e *domain.EventFact) { e.Role = "reviewer" }},
			{"event_type", func(e *domain.EventFact) { e.EventType = domain.EventTypeTurnMessage }},
			{"tool_name_changed", func(e *domain.EventFact) { e.ToolName = stringPtr("git") }},
			{"tool_name_cleared", func(e *domain.EventFact) { e.ToolName = nil }},
			{"status", func(e *domain.EventFact) { e.Status = "completed" }},
			{"error_message_changed", func(e *domain.EventFact) { e.ErrorMessage = stringPtr("other error") }},
			{"error_message_cleared", func(e *domain.EventFact) { e.ErrorMessage = nil }},
			{"duration_ms_changed", func(e *domain.EventFact) { e.DurationMs = int64Ptr(999) }},
			{"duration_ms_cleared", func(e *domain.EventFact) { e.DurationMs = nil }},
			{"context_tokens", func(e *domain.EventFact) { e.ContextTokens = 999999 }},
		}

		for _, tc := range eventCases {
			t.Run("event_"+tc.name, func(t *testing.T) {
				m := baseWithToolEvent
				events := append([]domain.EventFact(nil), baseWithToolEvent.EventFacts()...)
				tc.mutate(&events[0])
				setEventFacts(&m, events)
				err := validator.Validate(baseWithToolEvent, m, 2)
				if !errors.Is(err, migration.ErrInvariantViolation) {
					t.Errorf("expected ErrInvariantViolation for mutated %s, got %v", tc.name, err)
				}
			})
		}
	})
}

func TestInvariantValidator_SchemaV3Rules(t *testing.T) {
	validator := migration.NewInvariantValidator()

	v2Ver := 2
	v3Ver := 3

	toolNameBash := "bash"
	errMsgRaw := "command failed with exit status 1"
	dur100 := int64(100)

	baseV2 := sampleFactLedger()
	baseV2.SchemaVersion = &v2Ver
	setEventFacts(&baseV2, []domain.EventFact{
		{
			Timestamp:     "2026-10-18T10:00:00Z",
			SessionID:     "ses-1",
			Turn:          1,
			Role:          "orchestrator",
			EventType:     domain.EventTypeTurnMessage,
			Status:        "success",
			ContextTokens: 1000,
		},
		{
			Timestamp:     "2026-10-18T10:00:01Z",
			SessionID:     "ses-1",
			Turn:          2,
			Role:          "coder",
			EventType:     domain.EventTypeToolCall,
			ToolName:      &toolNameBash,
			Status:        "error",
			ErrorMessage:  &errMsgRaw,
			DurationMs:    &dur100,
			ContextTokens: 2000,
		},
		{
			Timestamp:     "2026-10-18T10:00:02Z",
			SessionID:     "ses-1",
			Turn:          3,
			Role:          "orchestrator",
			EventType:     domain.EventTypeCompaction,
			Status:        "success",
			DurationMs:    &dur100,
			ContextTokens: 800,
		},
	})

	validV3 := baseV2
	validV3.SchemaVersion = &v3Ver
	setEventFacts(&validV3, []domain.EventFact{
		{
			Timestamp:     "2026-10-18T10:00:01Z",
			SessionID:     "ses-1",
			Turn:          2,
			Role:          "coder",
			EventType:     domain.EventTypeToolCall,
			ToolName:      &toolNameBash,
			Status:        "error",
			ErrorMessage:  nil,
			DurationMs:    &dur100,
			ContextTokens: 2000,
		},
		{
			Timestamp:     "2026-10-18T10:00:02Z",
			SessionID:     "ses-1",
			Turn:          3,
			Role:          "orchestrator",
			EventType:     domain.EventTypeCompaction,
			Status:        "success",
			DurationMs:    &dur100,
			ContextTokens: 800,
		},
	})

	t.Run("valid v3 pruned events passes validation", func(t *testing.T) {
		err := validator.Validate(baseV2, validV3, 3)
		if err != nil {
			t.Fatalf("expected valid v3 migration to pass, got: %v", err)
		}
	})

	t.Run("v3 migration fails if events count increased", func(t *testing.T) {
		mIncreased := validV3
		setEventFacts(&mIncreased, append([]domain.EventFact(nil), validV3.EventFacts()...))
		setEventFacts(&mIncreased, append(mIncreased.EventFacts(), domain.EventFact{
			Timestamp: "2026-10-18T10:00:03Z",
			SessionID: "ses-1",
			Turn:      4,
			EventType: domain.EventTypeCompaction,
			Status:    "success",
		}))
		setEventFacts(&mIncreased, append(mIncreased.EventFacts(), domain.EventFact{
			Timestamp: "2026-10-18T10:00:04Z",
			SessionID: "ses-1",
			Turn:      5,
			EventType: domain.EventTypeCompaction,
			Status:    "success",
		}))
		// Now len is 4, which is > baseV2 len 3
		err := validator.Validate(baseV2, mIncreased, 3)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when events count increased, got: %v", err)
		}
	})

	t.Run("v3 migration fails if any retained event has non-nil ErrorMessage", func(t *testing.T) {
		mWithErr := validV3
		events := append([]domain.EventFact(nil), validV3.EventFacts()...)
		leakStr := "residual sensitive error message"
		events[0].ErrorMessage = &leakStr
		setEventFacts(&mWithErr, events)

		err := validator.Validate(baseV2, mWithErr, 3)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when event retains non-nil ErrorMessage, got: %v", err)
		}
	})

	t.Run("v3 migration fails if retained event does not correspond to any pre-migration event", func(t *testing.T) {
		mUnmatched := validV3
		events := append([]domain.EventFact(nil), validV3.EventFacts()...)
		inventedTool := "synthetic_tool"
		events[0].ToolName = &inventedTool
		setEventFacts(&mUnmatched, events)

		err := validator.Validate(baseV2, mUnmatched, 3)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when pruned event cannot be mapped, got: %v", err)
		}
	})

	t.Run("v3 migration fails if ComputeTotals aggregate metrics change", func(t *testing.T) {
		mMutatedTotals := validV3
		invs := append([]domain.InvocationFact(nil), validV3.InvocationFacts()...)
		// Mutate cost in invocation
		invs[0].CostUSD = 999.0
		setInvocationFacts(&mMutatedTotals, invs)

		err := validator.Validate(baseV2, mMutatedTotals, 3)
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation when totals mutate, got: %v", err)
		}
	})
}

func TestEngineMigrate(t *testing.T) {
	engine := migration.NewEngine(nil)

	unversionedJSON := `{
		"ticket_id": "AL-100",
		"ticket_title": "Historical Ticket",
		"ticket_type": "chore",
		"started_at": "2026-10-18T10:00:00Z",
		"completed_at": "2026-10-18T11:00:00Z",
		"updated_at": "2026-10-18T11:00:00Z",
		"code": {
			"lines_added": 50,
			"lines_deleted": 10,
			"test_lines_added": 25,
			"test_lines_deleted": 0,
			"prod_lines_added": 25,
			"prod_lines_deleted": 10,
			"files_changed": 2,
			"sdlc_iterations": 1
		},
		"invocations": [
			{
				"id": "inv-1",
				"role": "orchestrator",
				"model": "claude-3-opus",
				"user": "developer",
				"active_duration_seconds": 60,
				"turns": 2,
				"tokens": {
					"input": 100,
					"output": 50
				},
				"cost_usd": 0.05
			}
		],
		"tools": [
			{
				"name": "read",
				"calls": 5,
				"errors": 0
			}
		]
	}`

	t.Run("migrates unversioned JSON to v2", func(t *testing.T) {
		migrated, err := engine.Migrate([]byte(unversionedJSON), 2)
		if err != nil {
			t.Fatalf("unexpected error migrating unversioned json: %v", err)
		}

		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 2 {
			t.Errorf("expected schema_version 2, got %v", migrated.SchemaVersion)
		}
		if migrated.Sessions == nil {
			t.Errorf("expected initialized non-nil sessions collection")
		}
		if len(migrated.EventFacts()) != 0 {
			t.Errorf("expected 0 events, got %d", len(migrated.EventFacts()))
		}
		if migrated.Migration == nil {
			t.Fatalf("expected non-nil migration info")
		}
		if migrated.Migration.SchemaVersion != 2 {
			t.Errorf("expected migration schema_version 2, got %d", migrated.Migration.SchemaVersion)
		}
		if migrated.Code.LinesAdded != 50 {
			t.Errorf("expected lines_added 50, got %d", migrated.Code.LinesAdded)
		}
		if len(migrated.InvocationFacts()) != 1 || migrated.InvocationFacts()[0].ID != "inv-1" {
			t.Errorf("invocations not preserved properly")
		}
		if len(migrated.ToolFacts()) != 0 || migrated.Migration == nil || !strings.Contains(migrated.Migration.Notes, "Removed 1 unscoped legacy tool aggregate") {
			t.Errorf("expected unscoped tool row to be reported and removed, got tools=%+v migration=%+v", migrated.ToolFacts(), migrated.Migration)
		}
	})

	t.Run("idempotency: migrating already v2 ledger returns unmodified data", func(t *testing.T) {
		v2JSON := `{
			"schema_version": 2,
			"ticket_id": "AL-100",
			"updated_at": "2026-10-18T11:00:00Z",
			"code": {"lines_added": 10},
			"invocations": [],
			"tools": [],
			"events": []
		}`

		migrated, err := engine.Migrate([]byte(v2JSON), 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 2 {
			t.Errorf("expected schema_version 2, got %v", migrated.SchemaVersion)
		}
	})

	t.Run("migration halts and returns error when step violates invariants", func(t *testing.T) {
		reg := migration.NewRegistry()
		corruptStep := &mockStep{
			source: 1,
			target: 2,
			desc:   "corrupt step",
			apply: func(l domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
				ver := 2
				l.SchemaVersion = &ver
				// Invariant violation: wipe code facts
				l.Code.LinesAdded = 99999
				return l, nil
			},
		}
		_ = reg.Register(corruptStep)
		faultyEngine := migration.NewEngine(reg)

		_, err := faultyEngine.Migrate([]byte(unversionedJSON), 2)
		if err == nil {
			t.Fatalf("expected invariant violation error, got nil")
		}
		if !errors.Is(err, migration.ErrInvariantViolation) {
			t.Errorf("expected ErrInvariantViolation, got %v", err)
		}
	})

	t.Run("migrates empty json object {} to v2", func(t *testing.T) {
		migrated, err := engine.Migrate([]byte(`{}`), 2)
		if err != nil {
			t.Fatalf("unexpected error migrating empty json: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 2 {
			t.Errorf("expected schema_version 2, got %v", migrated.SchemaVersion)
		}
		if migrated.Sessions == nil || len(migrated.EventFacts()) != 0 {
			t.Errorf("expected empty initialized sessions collection")
		}
	})

	t.Run("empty and whitespace bytes return ErrMalformedJSON", func(t *testing.T) {
		_, errEmpty := engine.Migrate([]byte(""), 2)
		if !errors.Is(errEmpty, migration.ErrMalformedJSON) {
			t.Errorf("expected ErrMalformedJSON on empty bytes, got %v", errEmpty)
		}

		_, errWhitespace := engine.Migrate([]byte("   \n\t  "), 2)
		if !errors.Is(errWhitespace, migration.ErrMalformedJSON) {
			t.Errorf("expected ErrMalformedJSON on whitespace bytes, got %v", errWhitespace)
		}
	})

	t.Run("corrupted json syntax returns ErrMalformedJSON", func(t *testing.T) {
		_, err := engine.Migrate([]byte(`{"ticket_id": "AL-1", "code": {broken`), 2)
		if !errors.Is(err, migration.ErrMalformedJSON) {
			t.Errorf("expected ErrMalformedJSON, got %v", err)
		}
	})

	t.Run("target version less than 1 returns ErrUnsupportedVersion", func(t *testing.T) {
		_, err0 := engine.Migrate([]byte(unversionedJSON), 0)
		if !errors.Is(err0, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for target 0, got %v", err0)
		}

		_, errNeg := engine.Migrate([]byte(unversionedJSON), -1)
		if !errors.Is(errNeg, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for target -1, got %v", errNeg)
		}
	})

	t.Run("downgrade request returns ErrDowngradeNotAllowed", func(t *testing.T) {
		v2Payload := `{"schema_version": 2, "ticket_id": "AL-100", "updated_at": "2026-10-18T10:00:00Z"}`
		_, err := engine.Migrate([]byte(v2Payload), 1)
		if !errors.Is(err, migration.ErrDowngradeNotAllowed) {
			t.Errorf("expected ErrDowngradeNotAllowed, got %v", err)
		}
	})

	t.Run("missing migration step returns ErrMissingStep", func(t *testing.T) {
		// Default engine only has 1->2, 2->3, 3->4. Asking for version 5 should fail with missing step.
		_, err := engine.Migrate([]byte(unversionedJSON), 5)
		if !errors.Is(err, migration.ErrMissingStep) {
			t.Errorf("expected ErrMissingStep when chain has gap, got %v", err)
		}
	})

	t.Run("step apply error halts migration", func(t *testing.T) {
		reg := migration.NewRegistry()
		failingStep := &mockStep{
			source: 1,
			target: 2,
			desc:   "failing step",
			apply: func(l domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
				return domain.AgenticFactLedger{}, errors.New("disk failure during step")
			},
		}
		_ = reg.Register(failingStep)
		failingEngine := migration.NewEngine(reg)

		_, err := failingEngine.Migrate([]byte(unversionedJSON), 2)
		if err == nil {
			t.Fatalf("expected step failure error, got nil")
		}
	})

	t.Run("multi-step sequential migration v1 -> v2 -> v3", func(t *testing.T) {
		reg := migration.NewRegistry()
		_ = reg.Register(migration.NewStepV1ToV2())

		stepV2ToV3 := &mockStep{
			source: 2,
			target: 3,
			desc:   "step v2 to v3 adding note",
			apply: func(l domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
				v3 := 3
				l.SchemaVersion = &v3
				if l.Migration != nil {
					mig := *l.Migration
					mig.SchemaVersion = 3
					mig.StepsCompleted = append(mig.StepsCompleted, "schema_v3_feature")
					l.Migration = &mig
				}
				return l, nil
			},
		}
		_ = reg.Register(stepV2ToV3)

		multiEngine := migration.NewEngine(reg)
		migrated, err := multiEngine.Migrate([]byte(unversionedJSON), 3)
		if err != nil {
			t.Fatalf("multi-step migration failed: %v", err)
		}

		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 3 {
			t.Errorf("expected schema version 3, got %v", migrated.SchemaVersion)
		}
		if migrated.Migration == nil || len(migrated.Migration.StepsCompleted) != 2 {
			t.Errorf("expected 2 migration steps completed, got %+v", migrated.Migration)
		}
	})

	t.Run("migrates v2 ledger with events and error messages to v3 using real StepV2ToV3Sanitized", func(t *testing.T) {
		v2WithEventsJSON := `{
			"schema_version": 2,
			"ticket_id": "AL-32",
			"ticket_title": "Sanitization",
			"ticket_type": "feat",
			"started_at": "2026-10-18T10:00:00Z",
			"completed_at": "2026-10-18T10:30:00Z",
			"updated_at": "2026-10-18T10:30:00Z",
			"code": {
				"lines_added": 20,
				"lines_deleted": 5,
				"test_lines_added": 10,
				"test_lines_deleted": 0,
				"prod_lines_added": 10,
				"prod_lines_deleted": 5,
				"files_changed": 2,
				"sdlc_iterations": 1
			},
			"invocations": [
				{
					"id": "inv-1",
					"role": "orchestrator",
					"model": "claude-3-opus",
					"user": "developer",
					"active_duration_seconds": 60,
					"turns": 2,
					"tokens": {"input": 100, "output": 50},
					"cost_usd": 0.05
				}
			],
			"tools": [
				{"name": "read", "calls": 5, "errors": 0},
				{"name": "bash", "calls": 2, "errors": 1}
			],
			"events": [
				{
					"timestamp": "2026-10-18T10:00:01Z",
					"session_id": "ses-1",
					"turn": 1,
					"role": "orchestrator",
					"event_type": "turn_message",
					"status": "success",
					"context_tokens": 1000
				},
				{
					"timestamp": "2026-10-18T10:00:02Z",
					"session_id": "ses-1",
					"turn": 2,
					"role": "coder",
					"event_type": "tool_call",
					"tool_name": "bash",
					"status": "error",
					"error_message": "secret token leak: 12345",
					"duration_ms": 150,
					"context_tokens": 1200
				}
			]
		}`

		migrated, err := engine.Migrate([]byte(v2WithEventsJSON), 3)
		if err != nil {
			t.Fatalf("unexpected error migrating to v3: %v", err)
		}

		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 3 {
			t.Fatalf("expected schema_version 3, got %v", migrated.SchemaVersion)
		}

		if len(migrated.EventFacts()) != 1 {
			t.Fatalf("expected 1 event retained after pruning, got %d", len(migrated.EventFacts()))
		}

		retained := migrated.EventFacts()[0]
		if retained.Turn != 2 || retained.Status != "error" {
			t.Errorf("expected error tool event retained, got %+v", retained)
		}
		if retained.ToolName == nil || *retained.ToolName != "bash" {
			t.Errorf("expected tool_name bash, got %v", retained.ToolName)
		}
		if retained.ErrorMessage != nil {
			t.Errorf("expected ErrorMessage stripped to nil, got %s", *retained.ErrorMessage)
		}
		if retained.DurationMs == nil || *retained.DurationMs != 150 {
			t.Errorf("expected DurationMs 150, got %v", retained.DurationMs)
		}

		if migrated.Migration == nil {
			t.Fatal("expected non-nil migration info")
		}
		if migrated.Migration.SchemaVersion != 3 {
			t.Errorf("expected migration info schema_version 3, got %d", migrated.Migration.SchemaVersion)
		}
		hasStep := false
		for _, s := range migrated.Migration.StepsCompleted {
			if s == domain.MigrationStepSanitizedEvents {
				hasStep = true
				break
			}
		}
		if !hasStep {
			t.Errorf("expected %s in StepsCompleted: %+v", domain.MigrationStepSanitizedEvents, migrated.Migration.StepsCompleted)
		}
	})

	t.Run("chained migration from v1 unversioned to v3 via default NewEngine(nil)", func(t *testing.T) {
		migrated, err := engine.Migrate([]byte(unversionedJSON), 3)
		if err != nil {
			t.Fatalf("unexpected error on chained v1 -> v3 migration: %v", err)
		}

		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 3 {
			t.Errorf("expected schema version 3, got %v", migrated.SchemaVersion)
		}
		if migrated.Migration == nil || len(migrated.Migration.StepsCompleted) != 2 {
			t.Fatalf("expected 2 migration steps completed (baseline and sanitized), got %+v", migrated.Migration)
		}
		if migrated.Migration.StepsCompleted[0] != "schema_v2_baseline" || migrated.Migration.StepsCompleted[1] != domain.MigrationStepSanitizedEvents {
			t.Errorf("unexpected steps: %v", migrated.Migration.StepsCompleted)
		}
	})

	t.Run("idempotency: migrating already v3 ledger returns unmodified data", func(t *testing.T) {
		v3JSON := `{
			"schema_version": 3,
			"ticket_id": "AL-32",
			"updated_at": "2026-10-18T11:00:00Z",
			"code": {"lines_added": 10},
			"invocations": [],
			"tools": [],
			"events": []
		}`

		migrated, err := engine.Migrate([]byte(v3JSON), 3)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 3 {
			t.Errorf("expected schema_version 3, got %v", migrated.SchemaVersion)
		}
	})

	t.Run("MigrateLedger works on struct directly", func(t *testing.T) {
		ledger := sampleFactLedger()
		upgraded, err := engine.MigrateLedger(ledger, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upgraded.SchemaVersion == nil || *upgraded.SchemaVersion != 2 {
			t.Errorf("expected schema version 2, got %v", upgraded.SchemaVersion)
		}
	})

	t.Run("MigrateLedger edge cases", func(t *testing.T) {
		ledger := sampleFactLedger()

		// targetVersion < 1
		_, err0 := engine.MigrateLedger(ledger, 0)
		if !errors.Is(err0, migration.ErrUnsupportedVersion) {
			t.Errorf("expected ErrUnsupportedVersion for target 0, got %v", err0)
		}

		// already at target version
		v2 := 2
		ledgerV2 := ledger
		ledgerV2.SchemaVersion = &v2
		same, errSame := engine.MigrateLedger(ledgerV2, 2)
		if errSame != nil {
			t.Fatalf("unexpected error: %v", errSame)
		}
		if same.SchemaVersion == nil || *same.SchemaVersion != 2 {
			t.Errorf("expected schema version 2 unchanged")
		}

		// downgrade attempt
		_, errDown := engine.MigrateLedger(ledgerV2, 1)
		if !errors.Is(errDown, migration.ErrDowngradeNotAllowed) {
			t.Errorf("expected ErrDowngradeNotAllowed, got %v", errDown)
		}
	})

	t.Run("migrates v1 unversioned JSON cleanly to v4", func(t *testing.T) {
		v1JSON := `{
			"ticket_id": "AL-100",
			"ticket_title": "V1 to V4 Chain",
			"updated_at": "2026-10-18T11:00:00Z",
			"code": {"lines_added": 20, "sdlc_iterations": 4},
			"invocations": [
				{
					"id": "inv-root",
					"role": "orchestrator",
					"model": "claude-3-opus",
					"user": "developer",
					"turns": 2,
					"tokens": {"input": 100, "output": 50},
					"cost_usd": 0.05
				},
				{
					"id": "inv-child",
					"parent_id": "inv-root",
					"role": "coder",
					"model": "claude-3-opus",
					"user": "developer",
					"turns": 1,
					"tokens": {"input": 50, "output": 25},
					"cost_usd": 0.02
				}
			],
			"tools": []
		}`

		migrated, err := engine.Migrate([]byte(v1JSON), 4)
		if err != nil {
			t.Fatalf("unexpected error migrating v1 to v4: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 4 {
			t.Fatalf("expected schema_version 4, got %v", migrated.SchemaVersion)
		}
		if len(migrated.InvocationFacts()) != 2 {
			t.Fatalf("expected 2 invocations, got %d", len(migrated.InvocationFacts()))
		}
		if migrated.InvocationFacts()[0].UserIterations != 4 {
			t.Errorf("expected primary root invocation to receive 4 iterations from code, got %d", migrated.InvocationFacts()[0].UserIterations)
		}
		if migrated.InvocationFacts()[1].UserIterations != 0 {
			t.Errorf("expected child invocation to receive 0 iterations, got %d", migrated.InvocationFacts()[1].UserIterations)
		}
		if migrated.Migration == nil || migrated.Migration.SchemaVersion != 4 {
			t.Errorf("expected migration info schema_version 4, got %+v", migrated.Migration)
		}
	})

	t.Run("migrates v2 JSON cleanly to v4", func(t *testing.T) {
		v2JSON := `{
			"schema_version": 2,
			"ticket_id": "AL-102",
			"updated_at": "2026-10-18T11:00:00Z",
			"code": {"lines_added": 30, "sdlc_iterations": 7},
			"invocations": [
				{
					"id": "inv-v2-root",
					"role": "orchestrator",
					"model": "gpt-4o",
					"turns": 4,
					"tokens": {"input": 200, "output": 100},
					"cost_usd": 0.10
				}
			],
			"tools": [],
			"events": []
		}`

		migrated, err := engine.Migrate([]byte(v2JSON), 4)
		if err != nil {
			t.Fatalf("unexpected error migrating v2 to v4: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 4 {
			t.Fatalf("expected schema_version 4, got %v", migrated.SchemaVersion)
		}
		if migrated.InvocationFacts()[0].UserIterations != 7 {
			t.Errorf("expected UserIterations 7, got %d", migrated.InvocationFacts()[0].UserIterations)
		}
	})

	t.Run("migrates v3 JSON cleanly to v4", func(t *testing.T) {
		v3JSON := `{
			"schema_version": 3,
			"ticket_id": "AL-103",
			"updated_at": "2026-10-18T11:00:00Z",
			"code": {"lines_added": 10, "sdlc_iterations": 3},
			"invocations": [
				{
					"id": "inv-v3-root",
					"role": "orchestrator",
					"model": "gpt-4o",
					"turns": 2,
					"tokens": {"input": 50, "output": 25},
					"cost_usd": 0.05,
					"user_iterations": 0
				}
			],
			"tools": [],
			"events": []
		}`

		migrated, err := engine.Migrate([]byte(v3JSON), 4)
		if err != nil {
			t.Fatalf("unexpected error migrating v3 to v4: %v", err)
		}
		if migrated.SchemaVersion == nil || *migrated.SchemaVersion != 4 {
			t.Fatalf("expected schema_version 4, got %v", migrated.SchemaVersion)
		}
		if migrated.InvocationFacts()[0].UserIterations != 3 {
			t.Errorf("expected UserIterations 3, got %d", migrated.InvocationFacts()[0].UserIterations)
		}
	})
}

func TestStepV3ToV4UserIterations(t *testing.T) {
	step := migration.NewStepV3ToV4UserIterations()
	if step.SourceVersion() != 3 || step.TargetVersion() != 4 {
		t.Errorf("unexpected versions: source %d, target %d", step.SourceVersion(), step.TargetVersion())
	}
	if step.Description() == "" {
		t.Errorf("expected non-empty description")
	}

	t.Run("upgrades v3 to v4 and assigns Code.SDLCIterations to primary root invocation", func(t *testing.T) {
		initial := sampleFactLedger()
		v3 := 3
		initial.SchemaVersion = &v3
		initial.Code.SDLCIterations = 5
		setInvocationFacts(&initial, []domain.InvocationFact{
			{
				ID:             "root-1",
				ParentID:       nil,
				Role:           "orchestrator",
				Model:          "gpt-4o",
				UserIterations: 0,
			},
			{
				ID:             "root-2",
				ParentID:       nil,
				Role:           "orchestrator",
				Model:          "claude-3-5-sonnet",
				UserIterations: 0,
			},
			{
				ID:             "sub-1",
				ParentID:       stringPtr("root-1"),
				Role:           "coder",
				Model:          "gpt-4o",
				UserIterations: 0,
			},
		})

		upgraded, err := step.Apply(initial)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if upgraded.SchemaVersion == nil || *upgraded.SchemaVersion != 4 {
			t.Errorf("expected schema version 4, got %v", upgraded.SchemaVersion)
		}
		if upgraded.InvocationFacts()[0].UserIterations != 5 {
			t.Errorf("expected primary root invocation to receive 5 iterations, got %d", upgraded.InvocationFacts()[0].UserIterations)
		}
		if upgraded.InvocationFacts()[1].UserIterations != 0 {
			t.Errorf("expected secondary root invocation to have 0 iterations, got %d", upgraded.InvocationFacts()[1].UserIterations)
		}
		if upgraded.InvocationFacts()[2].UserIterations != 0 {
			t.Errorf("expected subagent invocation to have 0 iterations, got %d", upgraded.InvocationFacts()[2].UserIterations)
		}

		if upgraded.Migration == nil {
			t.Fatalf("expected non-nil migration info")
		}
		if upgraded.Migration.SchemaVersion != 4 {
			t.Errorf("expected migration schema version 4, got %d", upgraded.Migration.SchemaVersion)
		}
		foundStep := false
		for _, s := range upgraded.Migration.StepsCompleted {
			if s == domain.MigrationStepUserIterations {
				foundStep = true
				break
			}
		}
		if !foundStep {
			t.Errorf("expected step %s in steps completed", domain.MigrationStepUserIterations)
		}
	})

	t.Run("preserves existing non-zero UserIterations and does not overwrite with Code.SDLCIterations", func(t *testing.T) {
		initial := sampleFactLedger()
		v3 := 3
		initial.SchemaVersion = &v3
		initial.Code.SDLCIterations = 10
		setInvocationFacts(&initial, []domain.InvocationFact{
			{
				ID:             "root-1",
				ParentID:       nil,
				Role:           "orchestrator",
				Model:          "gpt-4o",
				UserIterations: 3,
			},
		})

		upgraded, err := step.Apply(initial)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upgraded.InvocationFacts()[0].UserIterations != 3 {
			t.Errorf("expected UserIterations to remain 3, got %d", upgraded.InvocationFacts()[0].UserIterations)
		}
	})

	t.Run("handles empty invocations slice cleanly", func(t *testing.T) {
		initial := sampleFactLedger()
		v3 := 3
		initial.SchemaVersion = &v3
		initial.Code.SDLCIterations = 4
		setInvocationFacts(&initial, nil)

		upgraded, err := step.Apply(initial)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if upgraded.SchemaVersion == nil || *upgraded.SchemaVersion != 4 {
			t.Errorf("expected schema version 4, got %v", upgraded.SchemaVersion)
		}
	})
}

func TestMigrationDropsUnscopedToolsWithDiagnostic(t *testing.T) {
	engine := migration.NewEngine(nil)
	input := `{"schema_version":3,"tools":[{"name":"bash","calls":4,"errors":1}]}`
	migrated, err := engine.Migrate([]byte(input), 4)
	if err != nil {
		t.Fatalf("expected v3 migration to complete, got %v", err)
	}
	if migrated.Migration == nil || !strings.Contains(migrated.Migration.Notes, "Removed 1 unscoped legacy tool aggregate") || !strings.Contains(migrated.Migration.Notes, "Re-extract from source logs") {
		t.Fatalf("expected explicit dropped-tool diagnostic, got %+v", migrated.Migration)
	}
	if migrated.Migration.Status != domain.MigrationStatusCompleted || len(migrated.Migration.StepsPending) != 0 {
		t.Fatalf("expected completed migration when there are no pending steps, got %+v", migrated.Migration)
	}
	if len(migrated.ToolFacts()) != 0 || len(migrated.Sessions) != 0 {
		t.Fatalf("unscoped tool aggregate should not be assigned to a fabricated session: %+v", migrated.Sessions)
	}
}

func TestMigrationDroppedToolNotePreservesPartialStatusForPendingSteps(t *testing.T) {
	engine := migration.NewEngine(nil)
	input := `{"schema_version":3,"migration":{"schema_version":3,"status":"partial","steps_completed":["git_code_facts"],"steps_pending":["opencode_db"],"last_migrated_at":"2026-10-20T10:00:00Z","last_migrated_by":"tester"},"tools":[{"name":"bash","calls":4,"errors":1}]}`
	migrated, err := engine.Migrate([]byte(input), 4)
	if err != nil {
		t.Fatalf("expected v3 migration to complete with pending external steps, got %v", err)
	}
	if migrated.Migration == nil || migrated.Migration.Status != domain.MigrationStatusPartial || len(migrated.Migration.StepsPending) != 1 || migrated.Migration.StepsPending[0] != "opencode_db" {
		t.Fatalf("expected partial status to remain for a pending step, got %+v", migrated.Migration)
	}
	if !strings.Contains(migrated.Migration.Notes, "Removed 1 unscoped legacy tool aggregate") {
		t.Fatalf("expected explicit dropped-tool note, got %+v", migrated.Migration)
	}
}

func TestMigrationEngineNormalizesInterimV4AndDropsUnscopedTools(t *testing.T) {
	engine := migration.NewEngine(nil)
	input := `{"schema_version":4,"invocations":[{"id":"root-1","parent_id":null,"role":"orchestrator","model":"model-a","user":"dev","tokens":{}}],"tools":[{"name":"bash","calls":4,"errors":1}],"events":[]}`
	migrated, err := engine.Migrate([]byte(input), 4)
	if err != nil {
		t.Fatalf("expected interim-v4 normalization, got %v", err)
	}
	if len(migrated.Sessions) != 1 || migrated.Sessions[0].ID != "root-1" || len(migrated.Sessions[0].Tools) != 0 {
		t.Fatalf("expected attributable facts migrated and unscoped tool omitted: %+v", migrated.Sessions)
	}
	if migrated.Migration == nil || !strings.Contains(migrated.Migration.Notes, "Removed 1 unscoped legacy tool aggregate") {
		t.Fatalf("expected migration diagnostic, got %+v", migrated.Migration)
	}
}
