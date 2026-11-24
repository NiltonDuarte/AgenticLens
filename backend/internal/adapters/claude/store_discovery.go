// Package claude reads Claude Code transcript stores for manual ticket
// telemetry. It reads only the explicitly selected store and never consults
// CLAUDE_CONFIG_DIR or a default home directory.
package claude

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"agenticlens/internal/adapters/filesystem"

	"golang.org/x/sys/unix"
)

// ErrOwnerUnavailable distinguishes a missing exact owner from ambiguous or
// failed discovery, whose diagnostics must remain visible to callers.
var ErrOwnerUnavailable = errors.New("Claude Code native session unavailable")

// ErrDiscoveryBudget means discovery or reads stopped before the selected tree
// was established. Callers retain stored facts and report incomplete recovery.
var ErrDiscoveryBudget = errors.New("configured Claude Code source discovery budget reached")

const (
	projectsDirectory  = "projects"
	subagentsDirectory = "subagents"
	agentFilePrefix    = "agent-"
	transcriptSuffix   = ".jsonl"
	metaSuffix         = ".meta.json"
	maxLineBytes       = 64 * 1024 * 1024
)

// DiscoveryBudget is operation-local resource accounting modeled on the Codex
// discovery budget. SeenPaths prevents a repeated read of one file from using
// another session slot; every byte actually read still consumes the allowance.
// A nil budget is unbounded, which manual generation uses.
type DiscoveryBudget struct {
	Context         context.Context
	MaxSessions     int
	MaxSourceBytes  int64
	SessionsRead    int
	SourceBytesRead int64
	SeenPaths       map[string]bool
}

func (b *DiscoveryBudget) readContext() context.Context {
	if b != nil && b.Context != nil {
		return b.Context
	}
	return context.Background()
}

func (b *DiscoveryBudget) validate() error {
	if b == nil {
		return nil
	}
	if b.MaxSessions < 1 || b.MaxSourceBytes < 1 || b.SessionsRead < 0 || b.SourceBytesRead < 0 || b.SessionsRead > b.MaxSessions || b.SourceBytesRead > b.MaxSourceBytes {
		return errors.New("Claude Code discovery requires positive explicit session/byte limits and valid operation-local usage")
	}
	return nil
}

// claim reserves a session slot for one transcript path.
func (b *DiscoveryBudget) claim(path string) error {
	if b == nil {
		return nil
	}
	if err := b.readContext().Err(); err != nil {
		return err
	}
	if b.SeenPaths[path] {
		return nil
	}
	if b.SessionsRead >= b.MaxSessions {
		return ErrDiscoveryBudget
	}
	if b.SeenPaths == nil {
		b.SeenPaths = make(map[string]bool)
	}
	b.SeenPaths[path] = true
	b.SessionsRead++
	return nil
}

// DiscoveryOptions selects one root session by a literal nonce, an exact root
// session ID, or a subagent agent ID that resolves to its root.
type DiscoveryOptions struct {
	ClaudeHome string
	Nonce      string
	SessionID  string
	Budget     *DiscoveryBudget
}

// SubagentFiles locates one native subagent transcript and its sibling meta
// file. MetaPath is empty when the meta file does not exist.
type SubagentFiles struct {
	AgentID        string
	TranscriptPath string
	MetaPath       string
}

// TreeFiles locates a root transcript and every subagent file stored beside it.
// It carries no transcript content and asserts no parentage.
type TreeFiles struct {
	RootID      string
	RootPath    string
	Subagents   []SubagentFiles
	SelectedID  string
	OrphanMetas []string
}

type rootLocation struct {
	id     string
	folder string
	path   string
}

type agentLocation struct {
	agentID string
	rootID  string
	folder  string
}

type storeListing struct {
	roots  []rootLocation
	agents []agentLocation
}

// DiscoverTree selects one root and lists its subagent files. Every project
// folder is searched, including those belonging to linked worktrees. Folder
// names are never decoded into project identity.
func DiscoverTree(options DiscoveryOptions) (TreeFiles, error) {
	ctx := options.Budget.readContext()
	if err := ctx.Err(); err != nil {
		return TreeFiles{}, err
	}
	if err := options.Budget.validate(); err != nil {
		return TreeFiles{}, err
	}
	nonce := options.Nonce
	sessionID := strings.TrimSpace(options.SessionID)
	if strings.TrimSpace(nonce) == "" && sessionID == "" {
		return TreeFiles{}, errors.New("Claude Code session discovery requires --nonce or --session-id")
	}
	listing, err := listStore(ctx, options.ClaudeHome)
	if err != nil {
		return TreeFiles{}, err
	}
	var selected rootLocation
	if strings.TrimSpace(nonce) != "" {
		selected, err = selectByNonce(ctx, listing, nonce, options.Budget)
		if err != nil {
			return TreeFiles{}, err
		}
		if sessionID != "" {
			explicit, explicitErr := selectByID(listing, sessionID)
			if explicitErr != nil {
				return TreeFiles{}, explicitErr
			}
			if explicit.id != selected.id || explicit.folder != selected.folder {
				return TreeFiles{}, errors.New("supplied nonce and session ID select different Claude Code root sessions")
			}
		}
	} else {
		selected, err = selectByID(listing, sessionID)
		if err != nil {
			return TreeFiles{}, err
		}
	}
	tree, err := treeFiles(ctx, selected)
	if err != nil {
		return TreeFiles{}, err
	}
	tree.SelectedID = sessionID
	if tree.SelectedID == "" {
		tree.SelectedID = selected.id
	}
	return tree, ctx.Err()
}

// ValidateStore requires the selected store to be a directory whose projects
// subdirectory is a readable directory. A home directory or any other folder
// without a Claude Code transcript layout is rejected before any source read,
// so no selection can publish while Claude Code owners are silently retained.
func ValidateStore(home string) error {
	const hint = "select the Claude Code store that contains projects/ with --claude-home or telemetry.sources.claude.home"
	if strings.TrimSpace(home) == "" {
		return fmt.Errorf("Claude Code store is required: %s", hint)
	}
	info, err := os.Stat(home)
	if err != nil {
		return fmt.Errorf("accessing selected Claude Code store %q: %w; %s", home, err, hint)
	}
	if !info.IsDir() {
		return fmt.Errorf("selected Claude Code store %q must be a directory; %s", home, hint)
	}
	projects := filepath.Join(home, projectsDirectory)
	info, err = os.Stat(projects)
	if err != nil {
		return fmt.Errorf("selected Claude Code store %q has no readable %s directory: %w; %s", home, projectsDirectory, err, hint)
	}
	if !info.IsDir() {
		return fmt.Errorf("selected Claude Code store %q: %s is not a directory; %s", home, projectsDirectory, hint)
	}
	directory, err := os.Open(projects)
	if err != nil {
		return fmt.Errorf("selected Claude Code store %q: reading %s: %w; %s", home, projectsDirectory, err, hint)
	}
	_, readErr := directory.Readdirnames(1)
	if readErr != nil && !errors.Is(readErr, io.EOF) {
		err = fmt.Errorf("selected Claude Code store %q: reading %s: %w; %s", home, projectsDirectory, readErr, hint)
	}
	return errors.Join(err, directory.Close())
}

// listStore enumerates candidate root transcripts and subagent transcripts
// without reading them. Non-transcript files such as tool-results are ignored.
func listStore(ctx context.Context, home string) (storeListing, error) {
	if strings.TrimSpace(home) == "" {
		return storeListing{}, errors.New("Claude Code store is required: supply --claude-home or telemetry.sources.claude.home in the selected local YAML")
	}
	physical, err := filesystem.PhysicalPath(home)
	if err != nil {
		return storeListing{}, fmt.Errorf("resolving Claude Code store: %w", err)
	}
	projects := filepath.Join(physical, projectsDirectory)
	folders, err := os.ReadDir(projects)
	if err != nil {
		return storeListing{}, fmt.Errorf("reading Claude Code projects directory: %w", err)
	}
	var listing storeListing
	for _, folder := range folders {
		if err := ctx.Err(); err != nil {
			return storeListing{}, err
		}
		if !folder.IsDir() {
			continue
		}
		folderPath := filepath.Join(projects, folder.Name())
		entries, err := os.ReadDir(folderPath)
		if err != nil {
			return storeListing{}, fmt.Errorf("reading Claude Code project folder %s: %w", folder.Name(), err)
		}
		for _, entry := range entries {
			name := entry.Name()
			switch {
			case entry.Type().IsRegular() && strings.HasSuffix(name, transcriptSuffix):
				id := strings.TrimSuffix(name, transcriptSuffix)
				if validNativeID(id) {
					listing.roots = append(listing.roots, rootLocation{id: id, folder: folderPath, path: filepath.Join(folderPath, name)})
				}
			case entry.IsDir() && validNativeID(name):
				agents, err := listAgentFiles(filepath.Join(folderPath, name, subagentsDirectory))
				if err != nil {
					return storeListing{}, err
				}
				for _, agentID := range agents {
					listing.agents = append(listing.agents, agentLocation{agentID: agentID, rootID: name, folder: folderPath})
				}
			}
		}
	}
	return listing, ctx.Err()
}

func listAgentFiles(directory string) ([]string, error) {
	entries, err := os.ReadDir(directory)
	if errors.Is(err, os.ErrNotExist) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading Claude Code subagent directory: %w", err)
	}
	var result []string
	for _, entry := range entries {
		name := entry.Name()
		if !entry.Type().IsRegular() || !strings.HasPrefix(name, agentFilePrefix) || !strings.HasSuffix(name, transcriptSuffix) {
			continue
		}
		id := strings.TrimSuffix(strings.TrimPrefix(name, agentFilePrefix), transcriptSuffix)
		if validNativeID(id) {
			result = append(result, id)
		}
	}
	sort.Strings(result)
	return result, nil
}

func validNativeID(id string) bool {
	return strings.TrimSpace(id) == id && id != "" && id != "." && id != ".." && !strings.ContainsAny(id, "/\\\x00")
}

func selectByID(listing storeListing, id string) (rootLocation, error) {
	var roots []rootLocation
	for _, root := range listing.roots {
		if root.id == id {
			roots = append(roots, root)
		}
	}
	if len(roots) > 1 {
		return rootLocation{}, fmt.Errorf("Claude Code session %s appears in %d project folders; ambiguous source", id, len(roots))
	}
	if len(roots) == 1 {
		return roots[0], nil
	}
	var agents []agentLocation
	for _, agent := range listing.agents {
		if agent.agentID == id {
			agents = append(agents, agent)
		}
	}
	if len(agents) > 1 {
		return rootLocation{}, fmt.Errorf("Claude Code agent %s appears under %d root sessions; ambiguous source", id, len(agents))
	}
	if len(agents) == 0 {
		return rootLocation{}, fmt.Errorf("Claude Code session %s: %w", id, ErrOwnerUnavailable)
	}
	return selectRoot(listing, agents[0].rootID, agents[0].folder)
}

func selectRoot(listing storeListing, rootID, folder string) (rootLocation, error) {
	var matches []rootLocation
	for _, root := range listing.roots {
		if root.id == rootID {
			matches = append(matches, root)
		}
	}
	if len(matches) > 1 {
		return rootLocation{}, fmt.Errorf("Claude Code session %s appears in %d project folders; ambiguous source", rootID, len(matches))
	}
	if len(matches) == 0 || matches[0].folder != folder {
		return rootLocation{}, fmt.Errorf("Claude Code root session %s transcript: %w", rootID, ErrOwnerUnavailable)
	}
	return matches[0], nil
}

// selectByNonce reads every root and subagent transcript to find the literal
// nonce. A match in a subagent transcript selects that subagent's root.
func selectByNonce(ctx context.Context, listing storeListing, nonce string, budget *DiscoveryBudget) (rootLocation, error) {
	type rootKey struct{ id, folder string }
	matched := make(map[rootKey]bool)
	for _, root := range listing.roots {
		found, err := fileContains(ctx, root.path, nonce, budget)
		if err != nil {
			if errors.Is(err, ErrDiscoveryBudget) || ctx.Err() != nil {
				return rootLocation{}, err
			}
			return rootLocation{}, fmt.Errorf("reading Claude Code transcript %s during nonce discovery: %w", root.id, err)
		}
		if found {
			matched[rootKey{root.id, root.folder}] = true
		}
	}
	for _, agent := range listing.agents {
		path := filepath.Join(agent.folder, agent.rootID, subagentsDirectory, agentFilePrefix+agent.agentID+transcriptSuffix)
		found, err := fileContains(ctx, path, nonce, budget)
		if err != nil {
			if errors.Is(err, ErrDiscoveryBudget) || ctx.Err() != nil {
				return rootLocation{}, err
			}
			return rootLocation{}, fmt.Errorf("reading Claude Code subagent transcript %s during nonce discovery: %w", agent.agentID, err)
		}
		if found {
			matched[rootKey{agent.rootID, agent.folder}] = true
		}
	}
	ids := make(map[string]bool)
	var keys []rootKey
	for key := range matched {
		keys = append(keys, key)
		ids[key.id] = true
	}
	if len(keys) == 0 {
		return rootLocation{}, errors.New("no Claude Code transcript contains the supplied nonce")
	}
	if len(ids) > 1 {
		names := make([]string, 0, len(ids))
		for id := range ids {
			names = append(names, id)
		}
		sort.Strings(names)
		return rootLocation{}, fmt.Errorf("nonce matches multiple Claude Code root sessions (%s); use --session-id", strings.Join(names, ", "))
	}
	return selectRoot(listing, keys[0].id, keys[0].folder)
}

func treeFiles(ctx context.Context, root rootLocation) (TreeFiles, error) {
	tree := TreeFiles{RootID: root.id, RootPath: root.path}
	directory := filepath.Join(root.folder, root.id, subagentsDirectory)
	agents, err := listAgentFiles(directory)
	if err != nil {
		return TreeFiles{}, err
	}
	known := make(map[string]bool, len(agents))
	for _, agentID := range agents {
		if err := ctx.Err(); err != nil {
			return TreeFiles{}, err
		}
		known[agentID] = true
		files := SubagentFiles{AgentID: agentID, TranscriptPath: filepath.Join(directory, agentFilePrefix+agentID+transcriptSuffix)}
		meta := filepath.Join(directory, agentFilePrefix+agentID+metaSuffix)
		if info, statErr := os.Lstat(meta); statErr == nil && info.Mode().IsRegular() {
			files.MetaPath = meta
		}
		tree.Subagents = append(tree.Subagents, files)
	}
	entries, err := os.ReadDir(directory)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return TreeFiles{}, fmt.Errorf("reading Claude Code subagent directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if strings.HasPrefix(name, agentFilePrefix) && strings.HasSuffix(name, metaSuffix) {
			id := strings.TrimSuffix(strings.TrimPrefix(name, agentFilePrefix), metaSuffix)
			if validNativeID(id) && !known[id] {
				tree.OrphanMetas = append(tree.OrphanMetas, id)
			}
		}
	}
	sort.Strings(tree.OrphanMetas)
	return tree, nil
}

// fileContains streams one transcript, charging the budget, and reports a
// literal match. It retains no content.
func fileContains(ctx context.Context, path, needle string, budget *DiscoveryBudget) (found bool, err error) {
	reader, closeReader, err := openSource(ctx, path, budget, true)
	if err != nil {
		return false, err
	}
	defer func() { err = errors.Join(err, closeReader()) }()
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 64*1024), maxLineBytes)
	for scanner.Scan() {
		if err := ctx.Err(); err != nil {
			return false, err
		}
		if strings.Contains(scanner.Text(), needle) {
			found = true
		}
	}
	if err := ctx.Err(); err != nil {
		return false, err
	}
	return found, scanner.Err()
}

// openSource opens one regular file for reading. A transcript claims a budget
// session slot; every read is wrapped with cancellation and byte accounting.
func openSource(ctx context.Context, path string, budget *DiscoveryBudget, transcript bool) (io.Reader, func() error, error) {
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if transcript {
		if err := budget.claim(path); err != nil {
			return nil, nil, err
		}
	} else if budget != nil && budget.SourceBytesRead >= budget.MaxSourceBytes {
		return nil, nil, ErrDiscoveryBudget
	}
	file, err := os.OpenFile(path, os.O_RDONLY|unix.O_NONBLOCK, 0)
	if err != nil {
		return nil, nil, err
	}
	info, err := file.Stat()
	if err == nil && !info.Mode().IsRegular() {
		err = errors.New("Claude Code source is not a regular file")
	}
	if err != nil {
		return nil, nil, errors.Join(err, file.Close())
	}
	var reader io.Reader = contextReader{ctx: ctx, reader: file}
	if budget != nil {
		reader = &budgetReader{reader: reader, budget: budget, remaining: info.Size()}
	}
	return reader, file.Close, nil
}

type contextReader struct {
	ctx    context.Context
	reader io.Reader
}

func (r contextReader) Read(body []byte) (int, error) {
	if err := r.ctx.Err(); err != nil {
		return 0, err
	}
	const maxReadBytes = 64 * 1024
	n, err := r.reader.Read(body[:min(len(body), maxReadBytes)])
	if cancelled := r.ctx.Err(); cancelled != nil {
		return n, cancelled
	}
	return n, err
}

type budgetReader struct {
	reader    io.Reader
	budget    *DiscoveryBudget
	remaining int64
}

func (r *budgetReader) Read(body []byte) (int, error) {
	if len(body) == 0 {
		return 0, nil
	}
	allowance := r.budget.MaxSourceBytes - r.budget.SourceBytesRead
	if allowance <= 0 {
		if r.remaining <= 0 {
			return 0, io.EOF
		}
		return 0, ErrDiscoveryBudget
	}
	if int64(len(body)) > allowance {
		body = body[:allowance]
	}
	n, err := r.reader.Read(body)
	r.budget.SourceBytesRead += int64(n)
	r.remaining -= int64(n)
	return n, err
}
