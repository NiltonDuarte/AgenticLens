package migration

import (
	"errors"

	"agenticlens/internal/domain"
)

// MigrationStep defines the pure transformation contract for a discrete schema version upgrade.
type MigrationStep interface {
	SourceVersion() int
	TargetVersion() int
	Description() string
	Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error)
}

// Domain sentinel errors for schema migration.
var (
	ErrUnsupportedVersion  = errors.New("unsupported schema version")
	ErrDowngradeNotAllowed = errors.New("schema downgrade is not allowed")
	ErrMissingStep         = errors.New("missing migration step for schema transition")
	ErrDuplicateStep       = errors.New("migration step already registered for source version")
	ErrInvariantViolation  = errors.New("migration invariant violation")
	ErrMalformedJSON       = errors.New("malformed JSON payload")
)
