package main

import (
	"context"
	"errors"
	"fmt"
	"time"

	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
)

// captureWorkerActive shares the same recorded activity and bounded exact child
// completion evidence for periodic work and release takeover. Unknown activity
// does not fabricate completion; terminal/idle roots never force a successor.
func captureWorkerActive(project string, rows []domain.CaptureRegistration, budget int64) (bool, string) {
	for _, row := range rows {
		if row.ParentID == "" && domain.CaptureNeedsPeriodicRefresh(row) {
			return true, ""
		}
	}
	detail := ""
	for _, row := range rows {
		if !domain.CaptureNeedsPeriodicRefresh(row) {
			continue
		}
		if row.ParentID != "" && row.Identity.Source == domain.SourceOpenCode && row.ProjectDirectory == project && budget >= 4*1024*1024 {
			budget -= 4 * 1024 * 1024
			db, err := opencode.OpenDB(row.Identity.Store)
			if err == nil {
				ctx, cancel := context.WithTimeout(context.Background(), time.Second)
				finished, proofErr := opencode.CaptureChildFinished(ctx, db.DB(), row.ParentID, row.Identity.SessionID)
				cancel()
				closeErr := db.Close()
				if proofErr == nil && closeErr == nil && finished {
					continue
				}
				if proofErr != nil || closeErr != nil {
					detail = fmt.Sprintf("child activity proof unavailable: %v", errors.Join(proofErr, closeErr))
				}
			} else {
				detail = fmt.Sprintf("child activity source unavailable: %v", err)
			}
		}
		return true, detail
	}
	return false, detail
}
