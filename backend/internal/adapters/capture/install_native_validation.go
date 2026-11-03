package capture

import (
	"context"
	"errors"
	"fmt"
	"strings"
)

// validateInstallNativeOutput compares output only after a successful execution.
// A canceled or failed probe is not evidence of a different installed release.
func validateInstallNativeOutput(ctx context.Context, binary, expected, label string, args ...string) error {
	output, err := installCommandOutput(ctx, binary, args...)
	if err != nil {
		if cause := ctx.Err(); cause != nil {
			if errors.Is(cause, context.DeadlineExceeded) {
				return fmt.Errorf("%s timed out: %w", label, errors.Join(cause, err))
			}
			return fmt.Errorf("%s canceled: %w", label, errors.Join(cause, err))
		}
		return fmt.Errorf("%s execution failed: %w", label, err)
	}
	if actual := strings.TrimSpace(string(output)); actual != expected {
		return fmt.Errorf("%s mismatch: observed %q, expected %q", label, actual, expected)
	}
	return nil
}
