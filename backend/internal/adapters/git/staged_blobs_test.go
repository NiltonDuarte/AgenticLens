package git_test

import (
	"context"
	"errors"
	"testing"

	"agenticlens/internal/adapters/git"
)

func TestStagedBlobHonorsByteLimitAndCancellation(t *testing.T) {
	project := t.TempDir()
	gitRun(t, project, "init", "-qb", "main")
	const relative = "pending/snapshot.json"
	const body = "staged contents\n"
	writeFile(t, project, relative, []byte(body))
	gitRun(t, project, "add", "--", relative)
	objectID := gitRun(t, project, "rev-parse", ":"+relative)
	writeFile(t, project, relative, []byte("a different working copy\n"))

	ctx := context.Background()
	got, err := git.ReadStagedBlob(ctx, project, objectID, int64(len(body)))
	if err != nil || string(got) != body {
		t.Fatalf("staged read at exact byte limit = %q, %v", got, err)
	}
	got, err = git.ReadStagedBlob(ctx, project, objectID, int64(len(body)-1))
	if err == nil || len(got) != 0 {
		t.Fatalf("oversized blob must return an error without partial bytes: %q, %v", got, err)
	}

	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	got, err = git.ReadStagedBlob(cancelled, project, objectID, int64(len(body)))
	if !errors.Is(err, context.Canceled) || len(got) != 0 {
		t.Fatalf("cancelled blob inspection = %q, %v", got, err)
	}
}
