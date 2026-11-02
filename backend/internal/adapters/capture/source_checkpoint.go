package capture

import (
	"agenticlens/internal/adapters/opencode"
	"agenticlens/internal/domain"
	"context"
	"fmt"
	"os"
)

// sourceFingerprint is an invalidation hint, never an ownership proof. Failed
// probes force extraction. Successful unchanged probes reuse accepted facts.
func sourceFingerprint(ctx context.Context, r domain.CaptureRegistration) (string, error) {
	if r.Identity.Source == domain.SourceCodex {
		info, err := os.Stat(r.TranscriptPath)
		if err != nil {
			return "", err
		}
		return fmt.Sprintf("%d:%d", info.Size(), info.ModTime().UnixNano()), nil
	}
	db, err := opencode.OpenDB(r.Identity.Store)
	if err != nil {
		return "", err
	}
	defer db.Close()
	return opencode.CaptureSourceFingerprint(ctx, db.DB(), r.Identity.SessionID)
}
