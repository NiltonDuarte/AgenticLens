package backfill

import "time"

// BackfillOptions specifies parameters and toggles for running the backfill engine.
type BackfillOptions struct {
	// SnapshotCutoff is optional for deterministic callers; zero selects the operation start.
	SnapshotCutoff  time.Time
	WorkspacesDir   string
	TicketID        string
	Source          string
	CodexHome       string
	RecomputeGit    bool
	CodeRef         string
	RecomputeDB     bool
	DryRun          bool
	UpdateSummaries bool
	DBPath          string
	Force           bool
	MigrateSchema   bool
	AllowFallbacks  bool
}

// MigrationOptions specifies parameters for standalone universal schema migration.
type MigrationOptions struct {
	WorkspacesDir   string
	TicketID        string
	TargetVersion   int
	DryRun          bool
	Backup          bool
	UpdateSummaries bool
}
