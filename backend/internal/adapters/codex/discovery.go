package codex

import (
	"agenticlens/internal/adapters/filesystem"
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

// DiscoveryOptions selects a Codex thread from local rollout files using an
// explicit nonce or thread ID. Ticket text and working-directory heuristics
// are intentionally excluded because they can identify the wrong session.
type DiscoveryOptions struct {
	CodexHome string
	Nonce     string
	SessionID string
	Budget    *DiscoveryBudget
}

type rolloutCandidate struct {
	path       string
	threadID   string
	parentID   string
	nonceMatch bool
}

// DiscoverRolloutPaths finds the selected root thread and its known child
// threads under the explicitly selected home's sessions and archived_sessions. It reads
// rollout records only to inspect session metadata and compare the requested
// nonce/ticket text; it does not retain or return transcript content.
func DiscoverRolloutPaths(options DiscoveryOptions) ([]string, error) {
	home := options.CodexHome
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply an explicit source store")
	}
	if err := options.Budget.validate(); err != nil {
		return nil, err
	}
	paths, err := listRolloutPathsWithBudget(home, options.Budget)
	if err != nil {
		return nil, err
	}
	candidates := make([]rolloutCandidate, 0, len(paths))
	for _, path := range paths {
		candidate, inspectErr := inspectRolloutWithBudget(path, options.Nonce, options.Budget)
		if inspectErr != nil {
			if errors.Is(inspectErr, ErrDiscoveryBudget) {
				return nil, inspectErr
			}
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
		matches := matchingRoots(candidates, func(candidate rolloutCandidate) bool { return candidate.nonceMatch })
		if len(matches) == 0 {
			return nil, fmt.Errorf("no Codex rollout contains the supplied nonce")
		}
		if len(matches) > 1 {
			return nil, fmt.Errorf("nonce matches multiple Codex root sessions (%s); use --session-id", strings.Join(matches, ", "))
		}
		selectedID = matches[0]
		if explicitID := strings.TrimSpace(options.SessionID); explicitID != "" {
			if !containsThread(candidates, explicitID) {
				return nil, fmt.Errorf("Codex session %q was not found under %s", explicitID, home)
			}
			explicitRoot := rootThreadID(candidates, explicitID)
			if explicitRoot != selectedID {
				return nil, fmt.Errorf("supplied nonce and session ID select different Codex root sessions")
			}
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
	default:
		return nil, errors.New("Codex session discovery requires --nonce or --session-id")
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

func containsThread(candidates []rolloutCandidate, threadID string) bool {
	for _, candidate := range candidates {
		if candidate.threadID == threadID {
			return true
		}
	}
	return false
}

// ParseRolloutPath opens and parses an uncompressed or gzip-compressed Codex rollout.
func ParseRolloutPath(path string) (Transcript, error) {
	return ParseRolloutPathAt(path, time.Now().UTC())
}

func ParseRolloutPathAt(path string, cutoff time.Time) (transcript Transcript, err error) {
	file, err := os.Open(path)
	if err != nil {
		return Transcript{}, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	var reader io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		compressed, gzipErr := gzip.NewReader(file)
		if gzipErr != nil {
			return Transcript{}, fmt.Errorf("opening compressed Codex rollout: %w", gzipErr)
		}
		defer func() { err = errors.Join(err, compressed.Close()) }()
		reader = compressed
	}
	transcript, err = ParseRolloutAt(reader, cutoff)
	if err != nil {
		return Transcript{}, fmt.Errorf("parsing Codex rollout %s: %w", path, err)
	}
	return transcript, nil
}

func listRolloutPathsWithBudget(home string, budget *DiscoveryBudget) ([]string, error) {
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply --codex-home or telemetry.sources.codex.home in the selected local YAML")
	}
	if err := budget.validate(); err != nil {
		return nil, err
	}
	var err error
	home, err = physicalRolloutPath(home)
	if err != nil {
		return nil, err
	}

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
				if err := budget.claim(path); err != nil {
					return err
				}
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

func inspectRolloutWithBudget(path, nonce string, budget *DiscoveryBudget) (candidate rolloutCandidate, err error) {
	if err := budget.claim(path); err != nil {
		return candidate, err
	}
	file, err := os.Open(path)
	if err != nil {
		return rolloutCandidate{}, err
	}
	defer func() { err = errors.Join(err, file.Close()) }()
	info, err := file.Stat()
	if err != nil {
		return candidate, err
	}
	sourceSize := info.Size()
	var reader io.Reader = file
	if strings.HasSuffix(path, ".gz") {
		sourceSize = -1
		compressed, gzipErr := gzip.NewReader(file)
		if gzipErr != nil {
			return rolloutCandidate{}, gzipErr
		}
		defer func() { err = errors.Join(err, compressed.Close()) }()
		reader = compressed
	}
	if budget != nil {
		reader = &discoveryBudgetReader{reader: reader, budget: budget, remainingSize: sourceSize}
	}
	candidate = rolloutCandidate{path: path}
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), maxDiscoveryLineBytes)
	for scanner.Scan() {
		line := scanner.Text()
		if nonce != "" && strings.Contains(line, nonce) {
			candidate.nonceMatch = true
		}
		if candidate.threadID != "" {
			continue
		}
		var record rolloutRecord
		if json.Unmarshal([]byte(line), &record) != nil || record.Type != "session_meta" {
			continue
		}
		metadata, metadataErr := decodeSessionMetadata(record.Payload)
		if metadataErr != nil {
			continue
		}
		candidate.threadID = strings.TrimSpace(metadata.ID)
		candidate.parentID = strings.TrimSpace(metadata.ParentID)
		if nonce == "" {
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

func matchingRoots(candidates []rolloutCandidate, match func(rolloutCandidate) bool) []string {
	roots := make(map[string]struct{})
	for _, candidate := range candidates {
		if !match(candidate) {
			continue
		}
		roots[rootThreadID(candidates, candidate.threadID)] = struct{}{}
	}
	result := make([]string, 0, len(roots))
	for rootID := range roots {
		result = append(result, rootID)
	}
	sort.Strings(result)
	return result
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

// DiscoverExactRolloutPath selects one owner rather than its complete native
// family. Family discovery remains separate so children validate their own cwd.
func DiscoverExactRolloutPath(home, sessionID string) (string, error) {
	return DiscoverExactRolloutPathWithBudget(home, sessionID, nil)
}

func DiscoverExactRolloutPathWithBudget(home, sessionID string, budget *DiscoveryBudget) (string, error) {
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("exact Codex session ID is required")
	}
	paths, err := listRolloutPathsWithBudget(home, budget)
	if err != nil {
		return "", err
	}
	selected := ""
	for _, path := range paths {
		candidate, e := inspectRolloutWithBudget(path, "", budget)
		if e != nil {
			if errors.Is(e, ErrDiscoveryBudget) {
				return "", e
			}
			continue
		}
		if candidate.threadID != sessionID {
			continue
		}
		if selected != "" {
			return "", fmt.Errorf("exact Codex owner %s has ambiguous rollout paths", sessionID)
		}
		selected = path
	}
	if selected == "" {
		return "", fmt.Errorf("exact Codex owner %s unavailable", sessionID)
	}
	return selected, nil
}

// ValidateRolloutStore proves physical membership in the selected live or archive
// store without rescanning every transcript for each callback. Session identity
// is validated separately from the authoritative metadata by the caller.
func ValidateRolloutStore(home, path string) error {
	var err error
	home, err = physicalRolloutPath(home)
	if err != nil {
		return err
	}

	actual, err := physicalRolloutPath(path)
	if err != nil {
		return err
	}
	for _, name := range []string{"sessions", "archived_sessions"} {
		root, err := physicalRolloutPath(filepath.Join(home, name))
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(root, actual)
		if err == nil && rel != "." && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel) {
			return nil
		}
	}
	return fmt.Errorf("Codex transcript does not belong to selected source store %s", home)
}

func physicalRolloutPath(path string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return "", errors.New("explicit Codex source path must be nonblank")
	}
	return filesystem.PhysicalPath(path)
}
