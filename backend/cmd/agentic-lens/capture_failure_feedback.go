package main

import (
	"context"
	"errors"
	"os"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/domain"
)

type captureAgentResponse struct {
	domain.CaptureRegistration
	CaptureFeedback string `json:"capture_feedback,omitempty"`
}

func captureSessionFailureFeedback(ctx context.Context, store *capture.Store, identity domain.CaptureIdentity, policy domain.CaptureConfig, state string) (string, error) {
	if !policy.Enabled {
		return "", nil
	}
	manifest, err := capture.ReadManifest(state, string(identity.Source))
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return "", err
	}
	if err == nil && !manifest.Active {
		return "", nil
	}
	return store.SessionFailureFeedback(ctx, identity)
}
