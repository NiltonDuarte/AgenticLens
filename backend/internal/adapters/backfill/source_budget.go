package backfill

import (
	"context"
	"errors"
	"math"
	"time"

	"agenticlens/internal/adapters/claude"
	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/opencode"
)

var errSourceBudget = errors.New("configured source recovery budget reached; retained owners remain incomplete")

type sourceReadBudget struct {
	sessions   int
	bytes      int64
	codexSeen  map[string]bool
	claudeSeen map[string]bool
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

// claudeDiscovery lends the remaining shared allowance to one Claude Code tree
// read. Paths already read in this operation do not consume another session.
func (b *sourceReadBudget) claudeDiscovery(ctx context.Context) (*claude.DiscoveryBudget, int, error) {
	if b == nil || b.bytes < 1 || (b.sessions < 1 && len(b.claudeSeen) == 0) {
		return nil, 0, errSourceBudget
	}
	if b.claudeSeen == nil {
		b.claudeSeen = make(map[string]bool)
	}
	seen := len(b.claudeSeen)
	return &claude.DiscoveryBudget{Context: ctx, MaxSessions: b.sessions + seen, MaxSourceBytes: b.bytes, SessionsRead: seen, SeenPaths: b.claudeSeen}, seen, nil
}

// chargeClaude records the sessions and bytes one Claude Code tree read used,
// including reads that ended in failure.
func (b *sourceReadBudget) chargeClaude(used *claude.DiscoveryBudget, seen int) {
	b.sessions -= used.SessionsRead - seen
	b.bytes -= used.SourceBytesRead
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
