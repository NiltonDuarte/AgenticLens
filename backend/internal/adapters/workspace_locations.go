package adapters

import (
	"agenticlens/internal/adapters/capture"
	"agenticlens/internal/adapters/config"
	"agenticlens/internal/adapters/filesystem"
	"agenticlens/internal/domain"
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
)

// WorkspaceLocation is an observed ledger location, never a folder-name identity.
type WorkspaceLocation struct {
	Diagnostics []domain.WorkspaceDiagnostic
	Path        string
	LedgerPath  string
	TicketID    string
	Artifacts   []domain.WorkspaceArtifactConfig
	// directoryInfo binds a resolved read to its observed physical home only
	// for this operation; it is never persisted or used as ticket identity.
	directoryInfo os.FileInfo
}

type WorkspaceResolutionError struct {
	Outcome     string
	TicketID    string
	Candidates  []string
	Diagnostics []domain.WorkspaceDiagnostic
}

func (e *WorkspaceResolutionError) Error() string {
	message := fmt.Sprintf("workspace %s for ticket %q; candidates=%v; inspect configured roots and select an explicit ledger directory", e.Outcome, e.TicketID, e.Candidates)
	for i, d := range e.Diagnostics {
		if i == 3 {
			message += "; additional location diagnostics available"
			break
		}
		message += fmt.Sprintf("; %s: %s", d.Path, d.Message)
	}
	return message
}

// LoadWorkspaceProject selects the project policy; directories are selected solely by roots configuration.
func LoadWorkspaceProject(project string) (string, domain.ProjectConfig, error) {
	cfg, err := config.LoadShared(project)
	if err != nil {
		return "", cfg, err
	}
	return filepath.Dir(cfg.ConfigPath), cfg, nil
}
func insideWorkspace(root, path string) bool {
	rel, e := filepath.Rel(root, path)
	return e == nil && rel != ".." && !strings.HasPrefix(rel, ".."+string(filepath.Separator)) && !filepath.IsAbs(rel)
}

// DiscoverWorkspaceLocations applies each configured scope, reporting partial roots independently.
// Validating full ledger evidence belongs to the common ledger decoder, not folder enumeration.
func DiscoverWorkspaceLocations(cfg domain.ProjectConfig) ([]WorkspaceLocation, []domain.WorkspaceDiagnostic) {
	locations, diagnostics, _ := discoverWorkspaceLocations(cfg)
	return locations, diagnostics
}

// Completeness describes enumeration, independently of rejected individual ledgers.
// Only a complete enumeration can establish uniqueness or absence without selection.
func discoverWorkspaceLocations(cfg domain.ProjectConfig) ([]WorkspaceLocation, []domain.WorkspaceDiagnostic, bool) {
	found := map[string]WorkspaceLocation{}
	rejected := map[string]bool{}
	diagnostics := []domain.WorkspaceDiagnostic{}
	incomplete := false
	reject := func(path string, err error) {
		diagnostics = append(diagnostics, domain.WorkspaceDiagnostic{Path: path, Message: "workspace rejected: " + err.Error()})
	}
	add := func(path string, err error) {
		incomplete = true
		diagnostics = append(diagnostics, domain.WorkspaceDiagnostic{Path: path, Message: "workspace discovery incomplete: " + err.Error()})
	}
	for _, root := range cfg.Workspaces.Roots {
		selected, err := config.ResolveConfiguredPath(filepath.Dir(cfg.ConfigPath), root.Path)
		if err != nil {
			add(root.Path, err)
			continue
		}
		physical, err := filepath.EvalSymlinks(selected)
		if err != nil {
			add(selected, fmt.Errorf("workspace root unavailable: %w", err))
			continue
		}
		info, err := os.Stat(physical)
		if err != nil || !info.IsDir() {
			add(selected, fmt.Errorf("workspace root is not an available directory"))
			continue
		}
		type entry struct {
			path  string
			depth int
		}
		queue := []entry{{physical, 0}}
		seen := map[string]bool{}
		scheduled := map[string]bool{physical: true}
		visited := 0
		limited := false
		for len(queue) > 0 {
			current := queue[0]
			queue = queue[1:]
			actual, err := filepath.EvalSymlinks(current.path)
			if err != nil {
				add(current.path, err)
				continue
			}
			if !insideWorkspace(physical, actual) {
				reject(current.path, fmt.Errorf("workspace alias points outside configured physical root"))
				continue
			}
			if seen[actual] {
				continue
			}
			seen[actual] = true
			if visited >= root.Discovery.MaxDirectories {
				limited = true
				break
			}
			visited++
			ledger := filepath.Join(actual, "agentic_metrics.json")
			if fi, e := os.Lstat(ledger); e == nil {
				if !fi.Mode().IsRegular() {
					reject(ledger, fmt.Errorf("ledger must be a regular file"))
				} else {
					artifacts := root.Artifacts
					if artifacts == nil {
						artifacts = cfg.Workspaces.Artifacts
					}
					candidate := WorkspaceLocation{Path: actual, LedgerPath: ledger, Artifacts: artifacts}
					if prior, ok := found[ledger]; ok && !reflect.DeepEqual(prior.Artifacts, artifacts) {
						reject(ledger, fmt.Errorf("overlapping roots have conflicting artifact policy"))
						rejected[ledger] = true
					} else {
						found[ledger] = candidate
					}
				}
			} else if !errors.Is(e, os.ErrNotExist) {
				add(ledger, e)
			}
			if current.depth >= root.Discovery.MaxDepth {
				continue
			}
			dir, e := os.Open(actual)
			if e != nil {
				add(actual, e)
				continue
			}
			for {
				children, e := dir.ReadDir(64)
				for _, child := range children {
					if child.IsDir() || child.Type()&os.ModeSymlink != 0 {
						path := filepath.Join(actual, child.Name())
						fi, e := os.Stat(path)
						if e != nil {
							if child.IsDir() {
								add(path, e)
							}
							continue
						}
						if !fi.IsDir() {
							continue
						}
						childPhysical, resolveErr := filepath.EvalSymlinks(path)
						if resolveErr != nil {
							add(path, resolveErr)
							continue
						}
						if !insideWorkspace(physical, childPhysical) {
							reject(path, fmt.Errorf("workspace alias points outside configured physical root"))
							continue
						}
						if scheduled[childPhysical] {
							continue
						}
						if len(queue)+visited >= root.Discovery.MaxDirectories {
							limited = true
							break
						}
						scheduled[childPhysical] = true
						queue = append(queue, entry{childPhysical, current.depth + 1})
					}
				}
				if limited || errors.Is(e, io.EOF) {
					break
				}
				if e != nil {
					add(actual, e)
					break
				}
			}
			if e := dir.Close(); e != nil {
				add(actual, e)
			}

		}
		if limited {
			add(selected, fmt.Errorf("incomplete workspace discovery: directory budget %d exhausted; narrow the configured root or explicitly increase max_directories", root.Discovery.MaxDirectories))
		}
	}
	locations := []WorkspaceLocation{}
	for path, location := range found {
		if !rejected[path] {
			locations = append(locations, location)
		}
	}
	sort.Slice(locations, func(i, j int) bool { return locations[i].Path < locations[j].Path })
	return locations, diagnostics, incomplete
}

type WorkspaceLedgerValidationError struct{ Cause error }

func (e *WorkspaceLedgerValidationError) Error() string { return e.Cause.Error() }
func (e *WorkspaceLedgerValidationError) Unwrap() error { return e.Cause }

func readWorkspaceLedger(location WorkspaceLocation, cfg domain.ProjectConfig) (domain.AgenticMetricsLedger, error) {
	ledger, _, _, err := readWorkspaceLedgerSnapshot(location, cfg, false)
	return ledger, err
}

func openResolvedWorkspaceDirectory(location WorkspaceLocation) (*capture.ReadDirectory, error) {
	if err := CheckPublication(location.LedgerPath); err != nil {
		return nil, err
	}
	directory, err := capture.ReadPublicationDirectory(location.LedgerPath)
	if err != nil {
		return nil, err
	}
	if location.directoryInfo != nil {
		info, statErr := directory.Stat(".")
		if statErr != nil || !os.SameFile(location.directoryInfo, info) {
			retry := &WorkspaceResolutionError{Outcome: "unavailable", TicketID: location.TicketID, Candidates: []string{location.Path}, Diagnostics: []domain.WorkspaceDiagnostic{{Path: location.Path, Message: "workspace changed after resolution; resolve and retry"}}}
			return nil, errors.Join(retry, statErr, directory.Close())
		}
	}
	return directory, nil
}

func readWorkspaceLedgerSnapshot(location WorkspaceLocation, cfg domain.ProjectConfig, retainDirectory bool) (ledger domain.AgenticMetricsLedger, info os.FileInfo, reference *os.File, err error) {
	directory, err := openResolvedWorkspaceDirectory(location)
	if err != nil {
		return ledger, nil, nil, err
	}
	defer func() {
		err = errors.Join(err, directory.Close())
		if err != nil && reference != nil {
			err = errors.Join(err, reference.Close())
			reference = nil
		}
	}()
	info, err = directory.Stat(".")
	if err != nil {
		return ledger, nil, nil, err
	}
	content, err := readDirectoryLedger(directory, location.LedgerPath)
	if err != nil {
		return ledger, nil, nil, err
	}
	ledger, err = decodeWorkspaceLedger(content, location, cfg)
	if err == nil && retainDirectory {
		// This independent descriptor retains the observed inode without the
		// read lease's flock, so writers can proceed between read phases.
		reference, err = directory.Open(".")
	}
	return ledger, info, reference, err
}

func decodeWorkspaceLedger(content []byte, location WorkspaceLocation, cfg domain.ProjectConfig) (domain.AgenticMetricsLedger, error) {
	var err error
	if err = ValidateRuntimeFactLedger(content); err != nil {
		return domain.AgenticMetricsLedger{}, &WorkspaceLedgerValidationError{Cause: err}
	}
	var dto AgenticFactLedgerDTO
	if err = json.Unmarshal(content, &dto); err != nil {
		return domain.AgenticMetricsLedger{}, &WorkspaceLedgerValidationError{Cause: err}
	}
	if err = dto.Validate(); err != nil {
		return domain.AgenticMetricsLedger{}, &WorkspaceLedgerValidationError{Cause: err}
	}
	facts, err := domain.NormalizeLedgerForRead(dto.ToDomain())
	if err != nil {
		return domain.AgenticMetricsLedger{}, &WorkspaceLedgerValidationError{Cause: err}
	}
	if facts.TicketID == nil || strings.TrimSpace(*facts.TicketID) == "" {
		return domain.AgenticMetricsLedger{}, &WorkspaceLedgerValidationError{Cause: fmt.Errorf("ledger ticket identity is unavailable")}
	}
	ledger := domain.DeriveMetricsLedger(facts, cfg)
	ledger.WorkspacePath = location.Path
	ledger.WorkspaceDiagnostics = append([]domain.WorkspaceDiagnostic{}, location.Diagnostics...)
	return ledger, nil
}
func sameTicket(left, right string, policy domain.TicketConfig) bool {
	if policy.CaseInsensitive {
		return strings.EqualFold(left, right)
	}
	return left == right
}

// ResolveTicketLocation returns only ledger-backed identity matches; no first-root precedence.
func ResolveTicketLocation(cfg domain.ProjectConfig, ticket, selection string) (WorkspaceLocation, error) {
	location, _, err := resolveTicketLocation(cfg, ticket, selection, false, false)
	return location, err
}

// resolveTicketReadLocation retains the actual observed directory until its
// caller completes final read admission. The caller must close the reference.
func resolveTicketReadLocation(cfg domain.ProjectConfig, ticket, selection string) (WorkspaceLocation, *os.File, error) {
	return resolveTicketLocation(cfg, ticket, selection, false, true)
}

// ResolveMaintenanceLocation selects recorded identity without decoding historical
// facts. Formatting and explicit migration validate their own representation under
// ownership; an explicit directory may select its recorded ID without a ticket filter.
// Runtime projections use validated runtime resolution with an explicit ticket ID.
func ResolveMaintenanceLocation(cfg domain.ProjectConfig, ticket, selection string) (WorkspaceLocation, error) {
	location, _, err := resolveTicketLocation(cfg, ticket, selection, true, false)
	return location, err
}

func RecordedLedgerTicket(data []byte) (string, error) {
	if !json.Valid(data) {
		return "", fmt.Errorf("invalid ledger JSON")
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	start, err := decoder.Token()
	if err != nil || start != json.Delim('{') {
		return "", fmt.Errorf("ledger must be an object")
	}
	var ticket string
	seen := false
	for decoder.More() {
		key, e := decoder.Token()
		if e != nil {
			return "", e
		}
		var value json.RawMessage
		if e = decoder.Decode(&value); e != nil {
			return "", e
		}
		if key == "ticket_id" {
			if seen {
				return "", fmt.Errorf("duplicate ledger ticket identity")
			}
			seen = true
			if e = json.Unmarshal(value, &ticket); e != nil {
				return "", e
			}
		}
	}
	if strings.TrimSpace(ticket) == "" {
		return "", fmt.Errorf("ledger has no recorded ticket identity")
	}
	return ticket, nil
}

func resolveTicketLocation(cfg domain.ProjectConfig, ticket, selection string, maintenance, retainDirectory bool) (resolved WorkspaceLocation, reference *os.File, err error) {
	var retained *os.File
	defer func() {
		if err != nil && retained != nil {
			err = errors.Join(err, retained.Close())
			reference = nil
		}
	}()
	if strings.TrimSpace(ticket) == "" && !(maintenance && selection != "") {
		return WorkspaceLocation{}, nil, fmt.Errorf("ticket ID must be nonblank")
	}
	locations, diagnostics, incomplete := discoverWorkspaceLocations(cfg)
	matches := []WorkspaceLocation{}
	selected := ""
	if selection != "" {
		var err error
		selected, err = filesystem.PhysicalPath(selection)
		if err != nil {
			return WorkspaceLocation{}, nil, err
		}
		// Selection does not rely on enumeration reaching the selected home. It
		// must still satisfy every configured scope and overlapping policy check.
		location, err := selectedWorkspaceLocation(cfg, selected)
		if err != nil {
			return WorkspaceLocation{}, nil, &WorkspaceResolutionError{Outcome: "unavailable", TicketID: ticket, Candidates: []string{selected}, Diagnostics: append(diagnostics, domain.WorkspaceDiagnostic{Path: selected, Message: err.Error()})}
		}
		locations = []WorkspaceLocation{location}
	}
	for _, location := range locations {
		var identity string
		var err error
		var candidateReference *os.File
		if maintenance {
			// This is advisory selection before ownership, so it must not wait on
			// the writer whose completed baseline the operation will validate.
			var body []byte
			body, err = capture.ReadPublicationFile(location.LedgerPath)
			if err == nil {
				identity, err = RecordedLedgerTicket(body)
			}
		} else {
			var ledger domain.AgenticMetricsLedger
			ledger, location.directoryInfo, candidateReference, err = readWorkspaceLedgerSnapshot(location, cfg, retainDirectory)
			if err == nil {
				identity = *ledger.TicketID
			}
		}
		if err != nil {
			diagnostic := domain.WorkspaceDiagnostic{Path: location.LedgerPath, Message: err.Error()}
			if selected == location.Path {
				return WorkspaceLocation{}, nil, &WorkspaceResolutionError{Outcome: func() string {
					var invalid *WorkspaceLedgerValidationError
					if errors.As(err, &invalid) {
						return "invalid_ledger"
					}
					return "unavailable"
				}(), TicketID: ticket, Candidates: []string{location.Path}, Diagnostics: []domain.WorkspaceDiagnostic{diagnostic}}
			}
			diagnostics = append(diagnostics, diagnostic)
			continue
		}
		if strings.TrimSpace(ticket) != "" && !sameTicket(identity, ticket, cfg.Ticket) {
			if candidateReference != nil {
				if closeErr := candidateReference.Close(); closeErr != nil {
					return WorkspaceLocation{}, nil, closeErr
				}
			}
			if selected == location.Path {
				return WorkspaceLocation{}, nil, &WorkspaceResolutionError{Outcome: "identity_conflict", TicketID: ticket, Candidates: []string{location.Path}}
			}
			continue
		}
		location.TicketID = identity
		if selected == "" || selected == location.Path {
			matches = append(matches, location)
			if retained == nil {
				retained = candidateReference
			} else if candidateReference != nil {
				if closeErr := candidateReference.Close(); closeErr != nil {
					return WorkspaceLocation{}, nil, closeErr
				}
			}
		}
	}
	if len(matches) == 1 && (selected != "" || !incomplete) {
		matches[0].Diagnostics = append([]domain.WorkspaceDiagnostic{}, diagnostics...)
		return matches[0], retained, nil
	}
	outcome := "not_found"
	if len(matches) > 1 {
		outcome = "ambiguous"
	} else if len(diagnostics) > 0 || incomplete {
		outcome = "unavailable"
	}
	paths := []string{}
	if selected != "" && len(matches) == 0 {
		paths = append(paths, selected)
	}
	for _, m := range matches {
		paths = append(paths, m.Path)
	}
	return WorkspaceLocation{}, nil, &WorkspaceResolutionError{Outcome: outcome, TicketID: ticket, Candidates: paths, Diagnostics: diagnostics}
}

// ValidateLedgerDestinationScope admits only canonical ledger files in selected scope.
func ValidateLedgerDestinationScope(cfg domain.ProjectConfig, path string) error {
	if filepath.Base(path) != "agentic_metrics.json" {
		return fmt.Errorf("ledger destination must name the canonical agentic_metrics.json ledger")
	}
	dir, err := filesystem.PhysicalPath(filepath.Dir(path))
	if err != nil {
		return err
	}
	_, err = workspaceDirectoryPolicy(cfg, dir)
	return err
}

func selectedWorkspaceLocation(cfg domain.ProjectConfig, dir string) (WorkspaceLocation, error) {
	artifacts, err := workspaceDirectoryPolicy(cfg, dir)
	if err != nil {
		return WorkspaceLocation{}, err
	}
	ledger := filepath.Join(dir, "agentic_metrics.json")
	info, err := os.Lstat(ledger)
	if err != nil {
		return WorkspaceLocation{}, fmt.Errorf("selected ledger unavailable: %w", err)
	}
	if !info.Mode().IsRegular() {
		return WorkspaceLocation{}, fmt.Errorf("ledger must be a regular file")
	}
	return WorkspaceLocation{Path: dir, LedgerPath: ledger, Artifacts: artifacts}, nil
}

// An explicit home uses the same physical scope and artifact policy as discovery,
// even when an enumeration budget prevented discovery from visiting it.
func workspaceDirectoryPolicy(cfg domain.ProjectConfig, dir string) ([]domain.WorkspaceArtifactConfig, error) {
	info, err := os.Stat(dir)
	if err != nil {
		return nil, fmt.Errorf("selected ledger directory must already exist: %w", err)
	}
	if !info.IsDir() {
		return nil, fmt.Errorf("selected ledger home must be a directory")
	}
	matched := false
	var artifacts []domain.WorkspaceArtifactConfig
	for _, r := range cfg.Workspaces.Roots {
		root, e := config.ResolveConfiguredPath(filepath.Dir(cfg.ConfigPath), r.Path)
		if e != nil {
			return nil, e
		}
		root, e = filesystem.PhysicalPath(root)
		if e != nil || !insideWorkspace(root, dir) {
			continue
		}
		rel, _ := filepath.Rel(root, dir)
		depth := 0
		if rel != "." {
			depth = len(strings.Split(rel, string(filepath.Separator)))
		}
		if depth <= r.Discovery.MaxDepth {
			policy := r.Artifacts
			if policy == nil {
				policy = cfg.Workspaces.Artifacts
			}
			if matched && !reflect.DeepEqual(artifacts, policy) {
				return nil, fmt.Errorf("overlapping roots have conflicting artifact policy")
			}
			matched = true
			artifacts = policy
		}
	}
	if !matched {
		return nil, fmt.Errorf("ledger destination is outside configured workspace discovery scope")
	}
	return artifacts, nil
}

// ValidateInitialLedgerDestination distinguishes explicit first publication from refresh.
func ValidateInitialLedgerDestination(cfg domain.ProjectConfig, ticket, path string) error {
	if ticket == "" || path == "" {
		return fmt.Errorf("first generation requires an explicit ticket ID and ledger destination")
	}
	if err := ValidateLedgerDestinationScope(cfg, path); err != nil {
		return err
	}
	dir, err := filesystem.PhysicalPath(filepath.Dir(path))
	if err != nil {
		return err
	}
	locations, diagnostics, incomplete := discoverWorkspaceLocations(cfg)
	if incomplete {
		return &WorkspaceResolutionError{Outcome: "unavailable", TicketID: ticket, Candidates: []string{dir}, Diagnostics: diagnostics}
	}
	for _, location := range locations {
		ledger, e := readWorkspaceLedger(location, cfg)
		if e == nil && sameTicket(*ledger.TicketID, ticket, cfg.Ticket) && location.Path != dir {
			return fmt.Errorf("ticket already has a ledger at %s; resolve its current location rather than recreating a prior destination", location.Path)
		}
	}
	return nil
}

// ValidateWorkspaceArtifactDestination keeps every public member in its selected ledger home.
// Root configuration authorizes location; source ownership is checked independently.
func ValidateWorkspaceArtifactDestination(cfg domain.ProjectConfig, ledger, destination string) error {
	ledger = capture.PublicationPath(ledger)
	destination = capture.PublicationPath(destination)
	if owner := capture.PublicationOwner(ledger); owner != nil {
		if capture.PublicationOwner(destination) != owner {
			return fmt.Errorf("artifact destination must remain inside selected ledger directory")
		}
		if err := owner.ValidateLocation(); err != nil {
			return err
		}
		home, err := owner.PhysicalMemberPath(filepath.Dir(ledger))
		if err != nil {
			return err
		}
		target, err := owner.PhysicalMemberPath(destination)
		if err != nil {
			return err
		}
		if target == home || !insideWorkspace(home, target) {
			return fmt.Errorf("artifact destination must remain inside selected ledger directory")
		}
		return nil
	}
	if err := ValidateLedgerDestinationScope(cfg, ledger); err != nil {
		return err
	}
	home, err := filesystem.PhysicalPath(filepath.Dir(ledger))
	if err != nil {
		return err
	}
	target, err := filesystem.PhysicalPath(destination)
	if err != nil {
		return err
	}
	if !insideWorkspace(home, target) || target == home {
		return fmt.Errorf("artifact destination must remain inside the selected ledger directory")
	}
	return nil
}
