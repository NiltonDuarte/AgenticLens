package backfill

// BackfillOptions specifies parameters and toggles for running the backfill engine.
type BackfillOptions struct {
	WorkspacesDir   string
	TicketID        string
	RecomputeGit    bool
	RecomputeDB     bool
	DryRun          bool
	UpdateSummaries bool
	DBPath          string
	Force           bool
	MigrateSchema   bool
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
