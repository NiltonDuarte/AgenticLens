package backfill

// TicketMigrationResult represents the migration result for an individual ticket workspace.
type TicketMigrationResult struct {
	TicketID       string   `json:"ticket_id"`
	WorkspacePath  string   `json:"workspace_path"`
	PreviousStatus string   `json:"previous_status"`
	NewStatus      string   `json:"new_status"`
	GitUpdated     bool     `json:"git_updated"`
	DBUpdated      bool     `json:"db_updated"`
	DBFound        bool     `json:"db_found"`
	Skipped        bool     `json:"skipped"`
	PendingSteps   []string `json:"pending_steps,omitempty"`
	Notes          string   `json:"notes,omitempty"`
	Error          error    `json:"error,omitempty"`
}

// BackfillReport summarizes the aggregate outcome of a backfill run.
type BackfillReport struct {
	Scanned         int                     `json:"scanned"`
	Updated         int                     `json:"updated"`
	Skipped         int                     `json:"skipped"`
	Partial         int                     `json:"partial"`
	Completed       int                     `json:"completed"`
	Warnings        []string                `json:"warnings"`
	Results         []TicketMigrationResult `json:"results"`
	SchemaMigration *SchemaMigrationReport  `json:"schema_migration,omitempty"`
}

// TicketSchemaMigrationResult represents the schema migration result for an individual ticket workspace.
type TicketSchemaMigrationResult struct {
	TicketID      string `json:"ticket_id"`
	WorkspacePath string `json:"workspace_path"`
	SourceVersion int    `json:"source_version"`
	TargetVersion int    `json:"target_version"`
	Status        string `json:"status"`
	BackupCreated bool   `json:"backup_created"`
	DiffSummary   string `json:"diff_summary"`
	Error         error  `json:"error,omitempty"`
}

// SchemaMigrationReport summarizes the aggregate outcome of a schema migration run.
type SchemaMigrationReport struct {
	Scanned  int                           `json:"scanned"`
	Migrated int                           `json:"migrated"`
	UpToDate int                           `json:"up_to_date"`
	Skipped  int                           `json:"skipped"`
	Errors   []string                      `json:"errors"`
	Results  []TicketSchemaMigrationResult `json:"results"`
}
