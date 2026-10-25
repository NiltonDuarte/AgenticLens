package codex

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

const maxDiscoveryLineBytes = 8 * 1024 * 1024

// DiscoveryOptions selects a Codex thread from local rollout files. Matching
// prioritizes a literal nonce, then a thread ID, then ticket/cwd discovery.
type DiscoveryOptions struct {
	CodexHome  string
	TicketID   string
	Nonce      string
	SessionID  string
	WorkingDir string
}

type rolloutCandidate struct {
	path        string
	threadID    string
	parentID    string
	workingDir  string
	updatedAt   time.Time
	nonceMatch  bool
	ticketMatch bool
}

// DiscoverRolloutPaths finds the selected root thread and its known child
// threads under CODEX_HOME/sessions and CODEX_HOME/archived_sessions. It reads
// rollout records only to inspect session metadata and compare the requested
// nonce/ticket text; it does not retain or return transcript content.
func DiscoverRolloutPaths(options DiscoveryOptions) ([]string, error) {
	home := strings.TrimSpace(options.CodexHome)
	if home == "" {
		return nil, errors.New("Codex home is empty")
	}
	workingDir := cleanPath(options.WorkingDir)
	if workingDir == "" {
		workingDir, _ = os.Getwd()
		workingDir = cleanPath(workingDir)
	}

	paths, err := listRolloutPaths(home)
	if err != nil {
		return nil, err
	}
	candidates := make([]rolloutCandidate, 0, len(paths))
	searchNonce := options.Nonce
	searchTicket := options.TicketID
	if searchNonce != "" || options.SessionID != "" {
		searchTicket = ""
	}
	for _, path := range paths {
		candidate, inspectErr := inspectRollout(path, searchNonce, searchTicket)
		if inspectErr != nil {
			// A partially written or future-format rollout should not hide other
			// usable sessions. Explicit selection reports a not-found error below.
			continue
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no readable Codex rollout JSONL files found under %s", home)
	}

	selectedID := strings.TrimSpace(options.SessionID)
	switch {
	case strings.TrimSpace(options.Nonce) != "":
		selectedID = newestMatchingRoot(candidates, func(candidate rolloutCandidate) bool { return candidate.nonceMatch })
		if selectedID == "" {
			return nil, fmt.Errorf("no Codex rollout contains the supplied nonce")
		}
	case selectedID != "":
		found := false
		for _, candidate := range candidates {
			if candidate.threadID == selectedID {
				found = true
				break
			}
		}
		if !found {
			return nil, fmt.Errorf("Codex session %q was not found under %s", selectedID, home)
		}
		selectedID = rootThreadID(candidates, selectedID)
	case strings.TrimSpace(options.TicketID) != "":
		selectedID = newestMatchingRootInWorkingDir(candidates, workingDir)
		if selectedID == "" {
			selectedID = newestMatchingRoot(candidates, func(candidate rolloutCandidate) bool { return candidate.ticketMatch })
		}
		if selectedID == "" {
			// A unique session in the current working directory is a safe fallback
			// when the ticket ID is not present in persisted rollout text.
			selectedID = uniqueWorkingDirectoryRoot(candidates, workingDir)
		}
		if selectedID == "" {
			return nil, fmt.Errorf("could not uniquely resolve a Codex session for ticket %q; use --nonce or --session-id", options.TicketID)
		}
	default:
		return nil, errors.New("Codex session discovery requires --nonce, --session-id, or --ticket-id")
	}

	selected := make([]rolloutCandidate, 0)
	for _, candidate := range candidates {
		if candidate.threadID == selectedID || isDescendant(candidates, candidate.threadID, selectedID) {
			selected = append(selected, candidate)
		}
	}
	if len(selected) == 0 {
		return nil, fmt.Errorf("Codex root session %q has no readable rollout", selectedID)
	}
	sort.Slice(selected, func(i, j int) bool {
		if selected[i].threadID == selectedID {
			return true
		}
		if selected[j].threadID == selectedID {
			return false
		}
		return selected[i].threadID < selected[j].threadID
	})
	result := make([]string, 0, len(selected))
	for _, candidate := range selected {
		result = append(result, candidate.path)
	}
	return result, nil
}

// ParseRolloutPath opens and parses an uncompressed or gzip-compressed Codex rollout.
func ParseRolloutPath(path string) (Transcript, error) {
	file, err := os.Open(path)
	if err != nil {
		return Transcript{}, err
	}
	defer file.Close()
	var reader io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		compressed, gzipErr := gzip.NewReader(file)
		if gzipErr != nil {
			return Transcript{}, fmt.Errorf("opening compressed Codex rollout: %w", gzipErr)
		}
		defer compressed.Close()
		reader = compressed
	}
	transcript, err := ParseRollout(reader)
	if err != nil {
		return Transcript{}, fmt.Errorf("parsing Codex rollout %s: %w", path, err)
	}
	return transcript, nil
}

func listRolloutPaths(home string) ([]string, error) {
	var paths []string
	for _, directory := range []string{"sessions", "archived_sessions"} {
		root := filepath.Join(home, directory)
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("accessing Codex %s directory: %w", directory, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() {
				return nil
			}
			if strings.HasSuffix(entry.Name(), ".jsonl") || strings.HasSuffix(entry.Name(), ".jsonl.gz") {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scanning Codex %s directory: %w", directory, err)
		}
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no rollout files found under %s/sessions or %s/archived_sessions", home, home)
	}
	sort.Strings(paths)
	return paths, nil
}

func inspectRollout(path, nonce, ticketID string) (rolloutCandidate, error) {
	file, err := os.Open(path)
	if err != nil {
		return rolloutCandidate{}, err
	}
	defer file.Close()
	var reader io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		compressed, gzipErr := gzip.NewReader(file)
		if gzipErr != nil {
			return rolloutCandidate{}, gzipErr
		}
		defer compressed.Close()
		reader = compressed
	}
	info, err := file.Stat()
	if err != nil {
		return rolloutCandidate{}, err
	}
	candidate := rolloutCandidate{path: path, updatedAt: info.ModTime()}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), maxDiscoveryLineBytes)
	for scanner.Scan() {
		line := scanner.Text()
		if nonce != "" && strings.Contains(line, nonce) {
			candidate.nonceMatch = true
		}
		if ticketID != "" && strings.Contains(strings.ToLower(line), strings.ToLower(ticketID)) {
			candidate.ticketMatch = true
		}
		if candidate.threadID != "" {
			continue
		}
		var record rolloutRecord
		if json.Unmarshal([]byte(line), &record) != nil || record.Type != "session_meta" {
			continue
		}
		var metadata sessionMetadata
		if json.Unmarshal(record.Payload, &metadata) != nil {
			continue
		}
		candidate.threadID = strings.TrimSpace(metadata.ID)
		candidate.parentID = strings.TrimSpace(metadata.ParentID)
		candidate.workingDir = cleanPath(metadata.CWD)
		if nonce == "" && ticketID == "" {
			break
		}
	}
	if err := scanner.Err(); err != nil {
		return rolloutCandidate{}, err
	}
	if candidate.threadID == "" {
		return rolloutCandidate{}, errors.New("rollout has no readable session metadata")
	}
	return candidate, nil
}

func newestMatchingRoot(candidates []rolloutCandidate, match func(rolloutCandidate) bool) string {
	updated := make(map[string]time.Time)
	for _, candidate := range candidates {
		if !match(candidate) {
			continue
		}
		rootID := rootThreadID(candidates, candidate.threadID)
		if candidate.updatedAt.After(updated[rootID]) {
			updated[rootID] = candidate.updatedAt
		}
	}
	var selected string
	var latest time.Time
	for rootID, timestamp := range updated {
		if selected == "" || timestamp.After(latest) {
			selected, latest = rootID, timestamp
		}
	}
	return selected
}

func newestMatchingRootInWorkingDir(candidates []rolloutCandidate, workingDir string) string {
	return newestMatchingRoot(candidates, func(candidate rolloutCandidate) bool {
		return candidate.ticketMatch && candidate.workingDir == workingDir
	})
}

func uniqueWorkingDirectoryRoot(candidates []rolloutCandidate, workingDir string) string {
	roots := make(map[string]struct{})
	for _, candidate := range candidates {
		if candidate.workingDir == workingDir {
			roots[rootThreadID(candidates, candidate.threadID)] = struct{}{}
		}
	}
	if len(roots) != 1 {
		return ""
	}
	for rootID := range roots {
		return rootID
	}
	return ""
}

func rootThreadID(candidates []rolloutCandidate, threadID string) string {
	parents := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		parents[candidate.threadID] = candidate.parentID
	}
	seen := make(map[string]struct{})
	for {
		if _, exists := seen[threadID]; exists {
			return threadID
		}
		seen[threadID] = struct{}{}
		parentID := parents[threadID]
		if parentID == "" || parents[parentID] == "" {
			return parentIDOrSelf(parentID, threadID, parents)
		}
		threadID = parentID
	}
}

func parentIDOrSelf(parentID, threadID string, parents map[string]string) string {
	if parentID != "" {
		if _, exists := parents[parentID]; exists {
			return parentID
		}
	}
	return threadID
}

func isDescendant(candidates []rolloutCandidate, threadID, rootID string) bool {
	seen := make(map[string]struct{})
	current := threadID
	parents := make(map[string]string, len(candidates))
	for _, candidate := range candidates {
		parents[candidate.threadID] = candidate.parentID
	}
	for current != "" {
		if current == rootID {
			return true
		}
		if _, exists := seen[current]; exists {
			return false
		}
		seen[current] = struct{}{}
		current = parents[current]
	}
	return false
}

func cleanPath(path string) string {
	if strings.TrimSpace(path) == "" {
		return ""
	}
	clean, err := filepath.Abs(filepath.Clean(path))
	if err != nil {
		return filepath.Clean(path)
	}
	return clean
}
