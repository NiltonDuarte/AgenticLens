package git

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"strconv"
	"strings"
)

const stagedDiffLimit = 1 << 20

// StagedBlob identifies an exact regular file in the selected commit index.
type StagedBlob struct {
	Name     string
	ObjectID string
}

// HasWorktreeMarker keeps manual capture checks outside Git on their health
// path. Git itself validates discovered metadata when staged inspection runs.
func HasWorktreeMarker(project string) (bool, error) {
	for dir := project; ; dir = filepath.Dir(dir) {
		if _, err := os.Lstat(filepath.Join(dir, ".git")); !errors.Is(err, os.ErrNotExist) {
			return err == nil, err
		}
		if filepath.Dir(dir) == dir {
			return false, nil
		}
	}
}

// StagedRegularBlobs lists the added or changed direct children of a literal
// project-relative directory. Git compares an unborn HEAD against an empty tree.
func StagedRegularBlobs(ctx context.Context, project string, directories []string) ([]StagedBlob, error) {
	args := []string{"diff", "--cached", "--raw", "-z", "--no-abbrev", "--no-ext-diff", "--no-textconv", "--no-renames", "--ita-invisible-in-index", "--diff-filter=AMT", "--relative", "--"}
	selected := make(map[string]bool, len(directories))
	for _, directory := range directories {
		if directory == "." || directory == ".." || strings.HasPrefix(directory, "../") || path.IsAbs(directory) || path.Clean(directory) != directory {
			return nil, errors.New("staged inspection requires a project-relative directory")
		}
		args = append(args, directory+"/")
		selected[directory] = true
	}
	if len(selected) == 0 {
		return nil, errors.New("staged inspection requires a project-relative directory")
	}
	body, err := readStagedGit(ctx, project, stagedDiffLimit, args...)
	if err != nil {
		return nil, fmt.Errorf("read staged Git diff: %w", err)
	}
	var blobs []StagedBlob
	for len(body) > 0 {
		header, rest, hasHeader := bytes.Cut(body, []byte{0})
		nameBytes, next, hasName := bytes.Cut(rest, []byte{0})
		fields := strings.Fields(string(header))
		if !hasHeader || !hasName || len(fields) != 5 || !strings.HasPrefix(fields[0], ":") {
			return blobs, errors.New("invalid staged Git diff record")
		}
		body = next
		name := string(nameBytes)
		if (fields[1] == "100644" || fields[1] == "100755") && selected[path.Dir(name)] {
			if !validObjectID(fields[3]) {
				return blobs, errors.New("invalid staged Git blob ID")
			}
			blobs = append(blobs, StagedBlob{Name: path.Base(name), ObjectID: fields[3]})
		}
	}
	return blobs, nil
}

// ReadStagedBlob reads the object from the selected project's object database.
// Its size is checked before the content command inflates the blob.
func ReadStagedBlob(ctx context.Context, project, id string, limit int64) ([]byte, error) {
	if !validObjectID(id) {
		return nil, errors.New("invalid staged Git blob ID")
	}
	output, err := readStagedGit(ctx, project, 64, "cat-file", "-s", id)
	if err != nil {
		return nil, fmt.Errorf("read staged Git blob size: %w", err)
	}
	size, err := strconv.ParseInt(strings.TrimSpace(string(output)), 10, 64)
	if err != nil || size < 0 || size > limit {
		return nil, errors.New("staged Git blob exceeds inspection bound or has an invalid size")
	}
	body, err := readStagedGit(ctx, project, size, "cat-file", "blob", id)
	if err != nil {
		return nil, fmt.Errorf("read staged Git blob: %w", err)
	}
	if int64(len(body)) != size {
		return nil, errors.New("staged Git blob size differs from its declared size")
	}
	return body, nil
}

func readStagedGit(ctx context.Context, project string, limit int64, args ...string) ([]byte, error) {
	cmd := exec.CommandContext(ctx, "git", append([]string{"--no-pager", "--no-optional-locks", "--no-replace-objects", "--literal-pathspecs", "-c", "core.fsmonitor=false", "-C", project}, args...)...)
	cmd.Env = append(historyGitEnvironment(os.Environ()), "GIT_NO_LAZY_FETCH=1", "GIT_ALLOW_PROTOCOL=")
	// Git may choose a temporary index and an object namespace for the commit
	// being observed. The rest of the environment cannot select another repo.
	for _, key := range []string{"GIT_INDEX_FILE", "GIT_OBJECT_DIRECTORY"} {
		value, set := os.LookupEnv(key)
		if !set {
			continue
		}
		if value != "" && !filepath.IsAbs(value) {
			cwd, err := os.Getwd()
			if err != nil {
				return nil, fmt.Errorf("resolve Git hook %s origin: %w", key, err)
			}
			value = cwd + string(filepath.Separator) + value
		}
		cmd.Env = append(cmd.Env, key+"="+value)
	}
	if alternates, set := os.LookupEnv("GIT_ALTERNATE_OBJECT_DIRECTORIES"); set {
		// Git parses its own alternate path list, including quoted separators.
		cmd.Env = append(cmd.Env, "GIT_ALTERNATE_OBJECT_DIRECTORIES="+alternates)
	}
	return readBoundedGit(cmd, limit)
}
