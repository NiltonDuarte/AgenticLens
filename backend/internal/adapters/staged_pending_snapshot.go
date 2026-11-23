package adapters

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/git"
	"agenticlens/internal/domain"
)

const (
	stagedPendingByteLimit = 16 << 20
	stagedPendingFileLimit = 64
	stagedDiscardTimeout   = 250 * time.Millisecond
)

type StagedPendingSnapshots struct {
	AwaitingTicket      int
	AwaitingDestination int
	Incomplete          bool
}

// InspectStagedPendingSnapshots classifies portable staged blobs first. The
// optional local registry can suppress an exact discard without owning their facts.
func InspectStagedPendingSnapshots(ctx context.Context, project string, directories []string, stateDir string) (result StagedPendingSnapshots) {
	blobs, diffErr := git.StagedRegularBlobs(ctx, project, directories)
	result.Incomplete = diffErr != nil
	type candidate struct {
		source      domain.SourceHarness
		rootSession string
		bound       bool
	}
	var snapshots []candidate
	remaining := int64(stagedPendingByteLimit)
	files := 0
	for _, blob := range blobs {
		id, match := strings.CutSuffix(blob.Name, ".json")
		if !match || len(id) != 2*sha256.Size || strings.ToLower(id) != id {
			continue
		}
		if _, err := hex.DecodeString(id); err != nil {
			continue
		}
		if ctx.Err() != nil || files >= stagedPendingFileLimit || remaining == 0 {
			result.Incomplete = true
			break
		}
		files++
		body, err := git.ReadStagedBlob(ctx, project, blob.ObjectID, remaining)
		if err != nil {
			result.Incomplete = true
			continue
		}
		remaining -= int64(len(body))
		snapshot, err := decodePendingSnapshot(body, blob.Name)
		if err != nil {
			result.Incomplete = true
			continue
		}
		snapshots = append(snapshots, candidate{snapshot.Source, snapshot.RootSessionID, snapshot.TicketID != ""})
	}
	if len(snapshots) == 0 {
		return result
	}

	discardCtx, cancel := context.WithTimeout(ctx, stagedDiscardTimeout)
	defer cancel()
	var store *capture.Store
	if stateDir != "" {
		if opened, err := capture.OpenReadOnlyProjectContext(discardCtx, stateDir, project); err == nil {
			store = opened
			defer func() {
				if err := store.Close(); err != nil {
					result.Incomplete = true
				}
			}()
		}
	}
	for _, snapshot := range snapshots {
		if store != nil && discardCtx.Err() == nil {
			discarded, err := store.ConfirmStagedDiscard(discardCtx, snapshot.source, snapshot.rootSession)
			if err == nil && discarded {
				continue
			}
		}
		if snapshot.bound {
			result.AwaitingDestination++
		} else {
			result.AwaitingTicket++
		}
	}
	return result
}
