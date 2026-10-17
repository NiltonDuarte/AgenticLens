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
	Scanned   int                     `json:"scanned"`
	Updated   int                     `json:"updated"`
	Skipped   int                     `json:"skipped"`
	Partial   int                     `json:"partial"`
	Completed int                     `json:"completed"`
	Warnings  []string                `json:"warnings"`
	Results   []TicketMigrationResult `json:"results"`
}
