package main

import (
	"errors"
	"fmt"

	"agenticlens/internal/adapters"
)

// recoverFailedGenerationPublication restores the accepted artifact set after a
// handled replacement failure. The caller still owns the publication lock.
// Process interruptions retain the same journal for recovery on the next run.
func recoverFailedGenerationPublication(ledger string, paths []string, cause error) error {
	if ledger == "" {
		return cause
	}
	recovered, err := adapters.RecoverPublication(ledger, paths)
	if err != nil {
		return errors.Join(cause, fmt.Errorf("publication rollback incomplete; inspect the recovery journal and retry generation: %w", err))
	}
	if !recovered {
		// Commit can remove the journal before its final directory sync fails.
		// Without the prior bytes, restoration cannot be claimed or fabricated.
		return errors.Join(cause, errors.New("no prior-good publication journal remains; replacements may already be committed and restoration could not be confirmed"))
	}
	return fmt.Errorf("%w; restored prior accepted artifact set", cause)
}
