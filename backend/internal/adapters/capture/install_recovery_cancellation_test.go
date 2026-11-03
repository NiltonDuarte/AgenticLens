package capture

import (
	"bytes"
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestInstallerCompletedRecoveryRetainsExpiredAcceptance(t *testing.T) {
	for _, source := range []string{"codex", "opencode"} {
		for _, expiration := range []string{"canceled", "deadline"} {
			t.Run(source+"/"+expiration, func(t *testing.T) {
				// Removal deliberately has no native-process validation that could
				// reject the context before the completed journal acceptance path.
				f := newInstallTransactionFixture(t, source)
				identity, err := currentInstallIdentity(context.Background(), f.options, f.repo, f.state, f.manifest)
				if err != nil {
					t.Fatal(err)
				}
				journal := installJournal{Protocol: 1, Phase: "applying", Identity: identity, Manifest: f.manifest}
				for _, change := range f.changes {
					journal.Changes = append(journal.Changes, installJournalChange{Path: change.path, Before: change.before, After: change.content, Existed: change.existed, Remove: change.remove})
					if err := apply(change); err != nil {
						t.Fatal(err)
					}
				}
				path := installJournalPath(f.state, source)
				if err := saveInstallJournal(path, journal, nil); err != nil {
					t.Fatal(err)
				}
				before, err := os.ReadFile(path)
				if err != nil {
					t.Fatal(err)
				}
				ctx, cancel := context.WithCancel(context.Background())
				cancel()
				cause := context.Canceled
				if expiration == "deadline" {
					ctx, cancel = context.WithDeadline(context.Background(), time.Unix(0, 0))
					defer cancel()
					cause = context.DeadlineExceeded
				}
				result, err := recoverInstallTransaction(ctx, f.options, f.repo, f.state)
				if result != nil || !errors.Is(err, cause) {
					t.Fatalf("completed recovery accepted expired required work: result=%+v err=%v", result, err)
				}
				after, err := os.ReadFile(path)
				if err != nil || !bytes.Equal(before, after) {
					t.Fatalf("expired acceptance changed recovery evidence: %v", err)
				}
				assertInstallFiles(t, f.changes, true)
				result, err = recoverInstallTransaction(context.Background(), f.options, f.repo, f.state)
				if err != nil || result == nil || result.Manifest.Source != source {
					t.Fatalf("uncanceled completed recovery failed: result=%+v err=%v", result, err)
				}
				assertInstallFiles(t, f.changes, true)
				if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
					t.Fatalf("accepted completed recovery retained journal: %v", err)
				}
			})
		}
	}
}
