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
}
