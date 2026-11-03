package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

const MaxCommittedLedgerBytes = 32 * 1024 * 1024

type CommittedLedger struct {
	Commit string
	Blob   string
	Data   []byte
}

// ResolveHistoryCommit pins publication history independently of code-metrics refs.
func ResolveHistoryCommit(ctx context.Context, project string) (string, error) {
	b, err := boundedGit(ctx, project, 128, "rev-parse", "--verify", "HEAD^{commit}")
	if err != nil {
		return "", fmt.Errorf("pin committed ledger HEAD: %w", err)
	}
	return strings.TrimSpace(string(b)), nil
}

// ResolveHistoryCommonDirectory uses the same explicit-project environment as
// committed object reads. Hook-local repository settings cannot redirect proof.
func ResolveHistoryCommonDirectory(ctx context.Context, project string) (string, error) {
	body, err := boundedGit(ctx, project, 4096, "rev-parse", "--path-format=absolute", "--git-common-dir")
	if err != nil {
		return "", fmt.Errorf("resolve history common Git directory: %w", err)
	}
	return strings.TrimSpace(string(body)), nil
}

// ResolveCommittedLedger reads one exact regular blob from a pinned commit tree.
// Neither the index nor working-tree contents grant historical authority.
func ResolveCommittedLedger(ctx context.Context, project, commit, relativePath string) (CommittedLedger, error) {
	result := CommittedLedger{Commit: commit}
	if !validObjectID(commit) {
		return result, errors.New("history requires a pinned commit object ID")
	}
	if relativePath == "" || relativePath == "." || path.Clean(relativePath) != relativePath || strings.HasPrefix(relativePath, "/") || relativePath == ".." || strings.HasPrefix(relativePath, "../") || strings.ContainsAny(relativePath, "\\\x00") {
		return result, errors.New("committed ledger path must be an exact repository-relative path")
	}
	entry, err := boundedGit(ctx, project, 8192, "ls-tree", "-z", commit, "--", ":(literal)"+relativePath)
	if err != nil {
		return result, fmt.Errorf("read committed ledger tree: %w", err)
	}
	entries := bytes.Split(entry, []byte{0})
	if len(entries) != 2 || len(entries[1]) != 0 {
		return result, errors.New("committed ledger path is missing or ambiguous")
	}
	header, name, ok := strings.Cut(string(entries[0]), "\t")
	fields := strings.Fields(header)
	if !ok || name != relativePath || len(fields) != 3 || (fields[0] != "100644" && fields[0] != "100755") || fields[1] != "blob" || !validObjectID(fields[2]) {
		return result, errors.New("committed ledger must be a regular file blob")
	}
	result.Blob = fields[2]
	sizeText, err := boundedGit(ctx, project, 64, "cat-file", "-s", result.Blob)
	if err != nil {
		return result, err
	}
	size, err := strconv.ParseInt(strings.TrimSpace(string(sizeText)), 10, 64)
	if err != nil || size < 0 || size > MaxCommittedLedgerBytes {
		return result, errors.New("committed ledger exceeds supported size bound")
	}
	result.Data, err = boundedGit(ctx, project, MaxCommittedLedgerBytes, "cat-file", "blob", result.Blob)
	return result, err
}

func validObjectID(s string) bool {
	if len(s) != 40 && len(s) != 64 {
		return false
	}
	for _, c := range s {
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

func boundedGit(ctx context.Context, project string, limit int64, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"--no-replace-objects", "-C", project}, args...)...)
	cmd.Env = historyGitEnvironment(os.Environ())
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}
	if err = cmd.Start(); err != nil {
		return nil, err
	}
	body, readErr := io.ReadAll(io.LimitReader(stdout, limit+1))
	if int64(len(body)) > limit || readErr != nil {
		killErr := cmd.Process.Kill()
		waitErr := cmd.Wait()
		return nil, errors.Join(errors.New("Git object output exceeds bound or could not be read"), readErr, killErr, waitErr)
	}
	if err = cmd.Wait(); err != nil {
		return nil, err
	}
	return body, nil
}

// Git lifecycle hooks export repository, worktree, object, index and config
// overrides. History authority belongs to the explicit project, so none of
// those inherited Git settings may select its repository or object namespace.
// This policy applies only to committed-history/proof subprocesses.
func historyGitEnvironment(inherited []string) []string {
	clean := make([]string, 0, len(inherited)+4)
	for _, entry := range inherited {
		name, _, _ := strings.Cut(entry, "=")
		if !strings.HasPrefix(strings.ToUpper(name), "GIT_") {
			clean = append(clean, entry)
		}
	}
	return append(clean, "GIT_CONFIG_NOSYSTEM=1", "GIT_CONFIG_SYSTEM="+os.DevNull, "GIT_CONFIG_GLOBAL="+os.DevNull, "GIT_TERMINAL_PROMPT=0")
}
