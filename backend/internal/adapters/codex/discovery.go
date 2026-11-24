package codex

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"agenticlens/internal/adapters/filesystem"
)

const maxDiscoveryLineBytes = 8 * 1024 * 1024

// ErrOwnerUnavailable distinguishes a missing exact owner from ambiguous or
// failed source discovery, whose diagnostics must remain visible to callers.
var ErrOwnerUnavailable = errors.New("Codex native session unavailable")

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
	path          string
	threadID      string
	parentID      string
	nonceMatch    bool
	inspectionErr error
}

// DiscoverRolloutPaths finds the selected root thread and its known child
// threads under the explicitly selected home's sessions and archived_sessions. It reads
// rollout records only to inspect session metadata and compare the requested
// nonce/ticket text; it does not retain or return transcript content.
func DiscoverRolloutPaths(options DiscoveryOptions) ([]string, error) {
	ctx := options.Budget.readContext()
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	home := options.CodexHome
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply an explicit source store")
	}
	if err := options.Budget.validate(); err != nil {
		return nil, err
	}
	paths, err := listRolloutPaths(ctx, home, options.Budget)
	if err != nil {
		return nil, err
	}
	candidates := make([]rolloutCandidate, 0, len(paths))
	for _, path := range paths {
		candidate, inspectErr := inspectRollout(ctx, path, options.Nonce, options.Budget)
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if inspectErr != nil {
			if errors.Is(inspectErr, ErrDiscoveryBudget) {
				return nil, inspectErr
			}
			// Keep native identity/ancestry already observed before a read failure.
			// Only an incomplete member of the selected family blocks extraction;
			// unrelated or entirely unattributable rollouts remain tolerable.
			if candidate.threadID == "" {
				continue
			}
			candidate.inspectionErr = inspectErr
		}
		candidates = append(candidates, candidate)
	}
	if len(candidates) == 0 {
		return nil, fmt.Errorf("no readable Codex rollout JSONL files found under %s", home)
	}

	// Check every relationship asserted by a copy before reducing ancestry to
	// one parent per owner. Otherwise a conflicting copy could route a selected
	// child into a different family and escape the ambiguity check.
	if err := rejectAmbiguousFamily(ctx, candidates, options); err != nil {
		return nil, err
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		if candidate.threadID == selectedID || isDescendant(candidates, candidate.threadID, selectedID) {
			if candidate.inspectionErr != nil {
				return nil, fmt.Errorf("incomplete Codex discovery for selected owner %s at %s: %w", candidate.threadID, candidate.path, candidate.inspectionErr)
			}
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

// rejectAmbiguousFamily follows all observed relationships, without selecting
// an authoritative copy. Ambiguity in an unrelated family does not block reads.
func rejectAmbiguousFamily(ctx context.Context, candidates []rolloutCandidate, options DiscoveryOptions) error {
	members := make(map[string]bool)
	counts := make(map[string]int)
	relations := make(map[string][]string)
	pending := make([]string, 0)
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		counts[candidate.threadID]++
		if candidate.parentID != "" {
			relations[candidate.threadID] = append(relations[candidate.threadID], candidate.parentID)
			relations[candidate.parentID] = append(relations[candidate.parentID], candidate.threadID)
		}
		if candidate.threadID == strings.TrimSpace(options.SessionID) ||
			(strings.TrimSpace(options.Nonce) != "" && candidate.nonceMatch) {
			pending = append(pending, candidate.threadID)
		}
	}
	for len(pending) > 0 {
		if err := ctx.Err(); err != nil {
			return err
		}
		id := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if members[id] {
			continue
		}
		members[id] = true
		pending = append(pending, relations[id]...)
	}
	for _, candidate := range candidates {
		if err := ctx.Err(); err != nil {
			return err
		}
		if members[candidate.threadID] && counts[candidate.threadID] > 1 {
			return fmt.Errorf("incomplete Codex discovery: owner %s has ambiguous rollout paths", candidate.threadID)
		}
	}
	return nil
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
	return ParseRolloutPathContext(context.Background(), path)
}

func ParseRolloutPathAt(path string, cutoff time.Time) (Transcript, error) {
	return ParseRolloutPathAtContext(context.Background(), path, cutoff)
}

// ParseRolloutPathContext reads and parses one rollout with the caller's context.
// On cancellation it may return owner metadata and observed events for a
// partial trace. A result accompanied by an error is never a snapshot.
func ParseRolloutPathContext(ctx context.Context, path string) (Transcript, error) {
	return ParseRolloutPathAtContext(ctx, path, time.Now().UTC())
}

// ParseRolloutPathAtContext parses a rollout through the selected activity cutoff.
func ParseRolloutPathAtContext(ctx context.Context, path string, cutoff time.Time) (Transcript, error) {
	transcript, _, err := parseRolloutPathContext(ctx, path, cutoff, nil)
	return transcript, err
}

// ParseRolloutPathWithinBudgetContext bounds decompressed bytes and reports the
// amount consumed even when the read is cancelled or otherwise incomplete.
func ParseRolloutPathWithinBudgetContext(ctx context.Context, path string, cutoff time.Time, maxBytes int64) (Transcript, int64, error) {
	if err := ctx.Err(); err != nil {
		return Transcript{}, 0, err
	}
	if maxBytes <= 0 {
		return Transcript{}, 0, ErrDiscoveryBudget
	}
	return parseRolloutPathContext(ctx, path, cutoff, &DiscoveryBudget{Context: ctx, MaxSourceBytes: maxBytes})
}

func parseRolloutPathContext(ctx context.Context, path string, cutoff time.Time, budget *DiscoveryBudget) (transcript Transcript, consumed int64, err error) {
	reader, size, closeReader, err := openRolloutReader(ctx, path)
	if err != nil {
		return Transcript{}, 0, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	var limited *discoveryBudgetReader
	if budget != nil {
		limited = &discoveryBudgetReader{reader: reader, budget: budget, remainingSize: size}
		reader = limited
	}
	transcript, err = ParseRolloutAtContext(ctx, reader, cutoff)
	if budget != nil {
		consumed = budget.SourceBytesRead
		if consumed == budget.MaxSourceBytes && !limited.eof && limited.remainingSize != 0 {
			err = ErrDiscoveryBudget
		}
	}
	if cancelled := ctx.Err(); cancelled != nil {
		err = cancelled
	}
	if err != nil {
		var prefix Transcript
		if budget == nil && (errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)) {
			prefix = transcript.ownedTracePrefix()
		}
		return prefix, consumed, fmt.Errorf("parsing Codex rollout %s: %w", path, err)
	}
	return transcript, consumed, nil
}

func (transcript Transcript) ownedTracePrefix() Transcript {
	ownerID := transcript.sessionIDOrThread()
	if ownerID == "" {
		return Transcript{}
	}
	prefix := Transcript{
		ThreadID: transcript.ThreadID, SessionID: transcript.SessionID,
		ParentThread: transcript.ParentThread, AgentPath: transcript.AgentPath,
		Warnings: []string{"interrupted Codex rollout"},
	}
	calls := make(map[string]struct{})
	for _, event := range transcript.Events {
		if event.SessionID == ownerID {
			prefix.Events = append(prefix.Events, event)
			if event.CallID != "" {
				calls[event.CallID] = struct{}{}
			}
		}
	}
	for _, delegation := range transcript.Delegations {
		if _, observed := calls[delegation.Provenance.CallID]; observed {
			prefix.Delegations = append(prefix.Delegations, delegation)
		}
	}
	return prefix
}

func listRolloutPaths(ctx context.Context, home string, budget *DiscoveryBudget) ([]string, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
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
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		root := filepath.Join(home, directory)
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("accessing Codex %s directory: %w", directory, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
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
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if len(paths) == 0 {
		return nil, fmt.Errorf("no rollout files found under %s/sessions or %s/archived_sessions", home, home)
	}
	sort.Strings(paths)
	return paths, nil
}

func inspectRollout(ctx context.Context, path, nonce string, budget *DiscoveryBudget) (candidate rolloutCandidate, err error) {
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if err := budget.claim(path); err != nil {
		return candidate, err
	}
	reader, sourceSize, closeReader, err := openRolloutReader(ctx, path)
	if err != nil {
		return rolloutCandidate{}, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	if budget != nil {
		reader = &discoveryBudgetReader{reader: reader, budget: budget, remainingSize: sourceSize}
	}
	candidate = rolloutCandidate{path: path}
	scanner := bufio.NewScanner(rolloutContextReader{ctx: ctx, reader: reader})
	scanner.Buffer(make([]byte, 64*1024), maxDiscoveryLineBytes)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return candidate, err
		}
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
	if err := ctx.Err(); err != nil {
		return candidate, err
	}
	if err := scanner.Err(); err != nil {
		return candidate, err
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
	return DiscoverExactRolloutPathContext(context.Background(), home, sessionID)
}

func DiscoverExactRolloutPathWithBudget(home, sessionID string, budget *DiscoveryBudget) (string, error) {
	return discoverExactRolloutPath(budget.readContext(), home, sessionID, budget)
}

// DiscoverExactRolloutPathContext resolves a single unambiguous native owner.
func DiscoverExactRolloutPathContext(ctx context.Context, home, sessionID string) (string, error) {
	return discoverExactRolloutPath(ctx, home, sessionID, nil)
}

func discoverExactRolloutPath(ctx context.Context, home, sessionID string, budget *DiscoveryBudget) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if strings.TrimSpace(sessionID) == "" {
		return "", errors.New("exact Codex session ID is required")
	}
	paths, err := listRolloutPaths(ctx, home, budget)
	if err != nil {
		return "", err
	}
	selected := ""
	for _, path := range paths {
		candidate, e := inspectRollout(ctx, path, "", budget)
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if e != nil {
			if errors.Is(e, ErrDiscoveryBudget) {
				return "", e
			}
			if candidate.threadID == sessionID {
				return "", fmt.Errorf("incomplete Codex discovery for exact owner %s at %s: %w", sessionID, path, e)
			}
			continue
		}
		if candidate.threadID != sessionID {
			continue
		}
		if selected != "" {
			return "", ambiguousOwnerError{sessionID: sessionID}
		}
		selected = path
	}
	if err := ctx.Err(); err != nil {
		return "", err
	}
	if selected == "" {
		return "", fmt.Errorf("exact Codex owner %s unavailable: %w", sessionID, ErrOwnerUnavailable)
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

// ErrOwnerAmbiguous distinguishes several rollouts proving one exact owner.
var ErrOwnerAmbiguous = errors.New("Codex native session has ambiguous rollout paths")

type ambiguousOwnerError struct {
	sessionID string
}

func (e ambiguousOwnerError) Error() string {
	return fmt.Sprintf("exact Codex owner %s has ambiguous rollout paths", e.sessionID)
}

func (e ambiguousOwnerError) Is(target error) bool {
	return target == ErrOwnerAmbiguous
}

// ChildCreationTolerance is how much earlier than its parent a capture child
// candidate's UUIDv7 creation instant may be and still be read. It covers
// clock steps only; a child is created after its parent.
const ChildCreationTolerance = 5 * time.Second

// MaxLocalDayOffset bounds how far a machine-local rollout day directory can
// end after the same calendar day read as UTC. It is the largest UTC offset.
const MaxLocalDayOffset = 14 * time.Hour

// rolloutName is what a conventional rollout file name claims. It is only a
// pre-filter: identity and parent always come from session_meta.
type rolloutName struct {
	created time.Time
	id      string
	v7      bool
}

// parseRolloutName recognizes rollout-YYYY-MM-DDThh-mm-ss-<uuid>.jsonl and its
// gzip-compressed form. Any other name reports false.
func parseRolloutName(name string) (rolloutName, bool) {
	base, found := strings.CutSuffix(name, ".jsonl.gz")
	if !found {
		base, found = strings.CutSuffix(name, ".jsonl")
	}
	if !found {
		return rolloutName{}, false
	}
	stamp, found := strings.CutPrefix(base, "rollout-")
	const stampLength = len("2006-01-02T15-04-05")
	if !found || len(stamp) != stampLength+1+36 || stamp[stampLength] != '-' {
		return rolloutName{}, false
	}
	if _, err := time.Parse("2006-01-02T15-04-05", stamp[:stampLength]); err != nil {
		return rolloutName{}, false
	}
	id := stamp[stampLength+1:]
	created, v7, ok := uuidCreation(id)
	if !ok {
		return rolloutName{}, false
	}
	return rolloutName{id: id, created: created, v7: v7}, true
}

// uuidCreation validates a canonical UUID and, for an RFC 9562 version 7
// UUID, returns the UTC creation instant in its first 48 bits.
func uuidCreation(id string) (time.Time, bool, bool) {
	if len(id) != 36 {
		return time.Time{}, false, false
	}
	var millis int64
	for i := 0; i < len(id); i++ {
		c := id[i]
		if i == 8 || i == 13 || i == 18 || i == 23 {
			if c != '-' {
				return time.Time{}, false, false
			}
			continue
		}
		value, ok := hexValue(c)
		if !ok {
			return time.Time{}, false, false
		}
		if i < 13 {
			millis = millis<<4 | int64(value)
		}
	}
	variant, _ := hexValue(id[19])
	if id[14] != '7' || variant < 8 || variant > 11 {
		return time.Time{}, false, true
	}
	return time.UnixMilli(millis).UTC(), true, true
}

func hexValue(c byte) (byte, bool) {
	switch {
	case c >= '0' && c <= '9':
		return c - '0', true
	case c >= 'a' && c <= 'f':
		return c - 'a' + 10, true
	case c >= 'A' && c <= 'F':
		return c - 'A' + 10, true
	}
	return 0, false
}

// CreationWindow prunes capture child candidates by creation instant. The
// zero value prunes nothing.
type CreationWindow struct {
	parent time.Time
}

// NewCreationWindow uses the parent's own UUIDv7 creation instant, or else its
// session_meta timestamp. Without either it prunes nothing.
func NewCreationWindow(parentID string, header RolloutHeader) CreationWindow {
	if created, v7, ok := uuidCreation(parentID); ok && v7 {
		return CreationWindow{parent: created}
	}
	return CreationWindow{parent: header.CreatedAt}
}

// Prunes reports whether the window has a parent instant.
func (window CreationWindow) Prunes() bool {
	return !window.parent.IsZero()
}

// SkipsFile reports a conventional, version 7 rollout name created more than
// ChildCreationTolerance before the parent. Such a file cannot be its child.
func (window CreationWindow) SkipsFile(name string) bool {
	if !window.Prunes() {
		return false
	}
	parsed, ok := parseRolloutName(name)
	return ok && parsed.v7 && window.parent.Sub(parsed.created) > ChildCreationTolerance
}

// SkipsDirectory reports a sessions/YYYY, sessions/YYYY/MM or sessions/YYYY/MM/DD
// directory of home whose machine-local period, read as UTC, ends more than
// MaxLocalDayOffset plus ChildCreationTolerance before the parent instant.
// Other directories, including archived_sessions, are always listed.
func (window CreationWindow) SkipsDirectory(home, path string) bool {
	if !window.Prunes() {
		return false
	}
	relative, err := filepath.Rel(filepath.Join(home, "sessions"), path)
	if err != nil || !filepath.IsLocal(relative) {
		return false
	}
	parts := strings.Split(filepath.ToSlash(relative), "/")
	if len(parts) > 3 || len(parts[0]) != 4 {
		return false
	}
	layouts := []string{"2006", "2006/01", "2006/01/02"}
	start, err := time.Parse(layouts[len(parts)-1], strings.Join(parts, "/"))
	if err != nil {
		return false
	}
	var end time.Time
	switch len(parts) {
	case 1:
		end = start.AddDate(1, 0, 0)
	case 2:
		if len(parts[1]) != 2 {
			return false
		}
		end = start.AddDate(0, 1, 0)
	default:
		if len(parts[1]) != 2 || len(parts[2]) != 2 {
			return false
		}
		end = start.AddDate(0, 0, 1)
	}
	return window.parent.Sub(end) > MaxLocalDayOffset+ChildCreationTolerance
}

// DiscoverExactCaptureRolloutPath resolves one unambiguous owner for automatic
// capture. Files whose conventional name carries a different UUID are skipped
// without a claim or a read. Every other rollout, including each copy named
// with the requested UUID and every name outside the convention, is claimed
// and its header read through headers, so duplicate copies remain ambiguous.
// Identity always comes from session_meta. headers may be nil.
func DiscoverExactCaptureRolloutPath(home, sessionID string, budget *DiscoveryBudget, headers *HeaderCache) (string, error) {
	ctx := budget.readContext()
	if err := ctx.Err(); err != nil {
		return "", err
	}
	sessionID = strings.TrimSpace(sessionID)
	if sessionID == "" {
		return "", errors.New("exact Codex session ID is required")
	}
	paths, err := listFilteredRolloutPaths(ctx, home, budget, func(name string) bool {
		parsed, conventional := parseRolloutName(name)
		return !conventional || strings.EqualFold(parsed.id, sessionID)
	})
	if errors.Is(err, errNoRolloutFiles) {
		return "", fmt.Errorf("exact Codex owner %s unavailable: %w: %w", sessionID, err, ErrOwnerUnavailable)
	}
	if err != nil {
		return "", err
	}
	selected := ""
	for _, path := range paths {
		if err := budget.claim(path); err != nil {
			return "", err
		}
		read, e := headers.ReadHeader(ctx, path, budget)
		threadID := strings.TrimSpace(read.Header.ID)
		if errors.Is(e, errHeaderBound) {
			// Keep the historical tolerance for an oversized first record.
			threadID, e = headers.oversizedThreadID(ctx, path, budget)
		}
		if err := ctx.Err(); err != nil {
			return "", err
		}
		if e != nil {
			if errors.Is(e, ErrDiscoveryBudget) {
				return "", e
			}
			if threadID == sessionID {
				return "", fmt.Errorf("incomplete Codex discovery for exact owner %s at %s: %w", sessionID, path, e)
			}
			continue
		}
		if threadID != sessionID {
			continue
		}
		if selected != "" {
			return "", ambiguousOwnerError{sessionID: sessionID}
		}
		selected = path
	}
	if selected == "" {
		return "", fmt.Errorf("exact Codex owner %s unavailable: %w", sessionID, ErrOwnerUnavailable)
	}
	return selected, nil
}

var errNoRolloutFiles = errors.New("no rollout files found")

// listFilteredRolloutPaths lists rollouts whose file name keep accepts,
// without claiming them. It reports errNoRolloutFiles when the store holds no
// rollout at all, and an empty list when none is kept.
func listFilteredRolloutPaths(ctx context.Context, home string, budget *DiscoveryBudget, keep func(name string) bool) ([]string, error) {
	if strings.TrimSpace(home) == "" {
		return nil, errors.New("Codex home is required: supply --codex-home or telemetry.sources.codex.home in the selected local YAML")
	}
	if err := budget.validate(); err != nil {
		return nil, err
	}
	home, err := physicalRolloutPath(home)
	if err != nil {
		return nil, err
	}
	var paths []string
	total := 0
	for _, directory := range []string{"sessions", "archived_sessions"} {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		root := filepath.Join(home, directory)
		if _, err := os.Stat(root); errors.Is(err, os.ErrNotExist) {
			continue
		} else if err != nil {
			return nil, fmt.Errorf("accessing Codex %s directory: %w", directory, err)
		}
		err := filepath.WalkDir(root, func(path string, entry os.DirEntry, walkErr error) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			if walkErr != nil {
				return walkErr
			}
			if entry.IsDir() || (!strings.HasSuffix(entry.Name(), ".jsonl") && !strings.HasSuffix(entry.Name(), ".jsonl.gz")) {
				return nil
			}
			total++
			if keep(entry.Name()) {
				paths = append(paths, path)
			}
			return nil
		})
		if err != nil {
			return nil, fmt.Errorf("scanning Codex %s directory: %w", directory, err)
		}
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if total == 0 {
		return nil, fmt.Errorf("%w under %s/sessions or %s/archived_sessions", errNoRolloutFiles, home, home)
	}
	sort.Strings(paths)
	return paths, nil
}
