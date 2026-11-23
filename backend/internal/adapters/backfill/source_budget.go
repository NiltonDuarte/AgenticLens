package backfill

import (
	"context"
	"errors"
	"math"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/opencode"
)

var errSourceBudget = errors.New("configured source recovery budget reached; retained owners remain incomplete")

type sourceReadBudget struct {
	sessions  int
	bytes     int64
	codexSeen map[string]bool
}

func (b *sourceReadBudget) beginOwner() error {
	if b == nil || b.sessions < 1 || b.bytes < 1 {
		return errSourceBudget
	}
	b.sessions--
	return nil
}

func (b *sourceReadBudget) parseCodex(ctx context.Context, path string, cutoff time.Time) (codex.Transcript, error) {
	if err := ctx.Err(); err != nil {
		return codex.Transcript{}, err
	}
	if b == nil || b.bytes < 1 {
		return codex.Transcript{}, errSourceBudget
	}
	if !b.codexSeen[path] {
		if err := b.beginOwner(); err != nil {
			return codex.Transcript{}, err
		}
		if b.codexSeen == nil {
			b.codexSeen = make(map[string]bool)
		}
		b.codexSeen[path] = true
	}
	transcript, consumed, err := codex.ParseRolloutPathWithinBudgetContext(ctx, path, cutoff, b.bytes)
	b.bytes -= consumed
	if ctxErr := ctx.Err(); ctxErr != nil {
		return codex.Transcript{}, errors.Join(ctxErr, err)
	}
	if errors.Is(err, codex.ErrDiscoveryBudget) {
		return codex.Transcript{}, errSourceBudget
	}
	return transcript, err
}

// Include previously visited IDs in the bounded metadata page, along with one
// extra row to detect that the remaining snapshot budget cannot cover the tree.
func (b *sourceReadBudget) discoveryLimit(visited int) int {
	if b.sessions >= math.MaxInt-visited {
		return math.MaxInt
	}
	return b.sessions + visited + 1
}

func isSourceBudgetError(err error) bool {
	return errors.Is(err, errSourceBudget) || errors.Is(err, codex.ErrDiscoveryBudget) || errors.Is(err, opencode.ErrCaptureBudget)
}
