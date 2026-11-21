package backfill

import (
	"context"
	"errors"
	"time"

	"agenticlens/internal/adapters/codex"
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

func (b *sourceReadBudget) discoveryLimit() int {
	limit := b.sessions
	if limit < int(^uint(0)>>1) {
		limit++
	}
	return limit
}
