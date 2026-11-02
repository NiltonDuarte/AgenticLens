package adapters

import (
	"agenticlens/internal/adapters/capture"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const publicationLimit = 128 << 20

type priorArtifact struct {
	Path    string `json:"path"`
	Present bool   `json:"present"`
	Data    []byte `json:"data"`
}

// validatePublicationObject requires exact, unique field names. JSON's normal
// case-insensitive and last-key-wins struct decoding is too permissive here.
func validatePublicationObject(body []byte, fields []string) error {
	decoder := json.NewDecoder(bytes.NewReader(body))
	token, err := decoder.Token()
	if err != nil {
		return err
	}
	if token != json.Delim('{') {
		return errors.New("publication metadata must be an object")
	}
	seen := map[string]bool{}
	for decoder.More() {
		token, err = decoder.Token()
		if err != nil {
			return err
		}
		key, ok := token.(string)
		if !ok {
			return errors.New("invalid publication metadata key")
		}
		allowed := false
		for _, field := range fields {
			if field == key {
				allowed = true
			}
		}
		if !allowed || seen[key] {
			return fmt.Errorf("unknown or duplicate publication metadata field %q", key)
		}
		seen[key] = true
		var value json.RawMessage
		if err := decoder.Decode(&value); err != nil {
			return err
		}
	}
	if _, err := decoder.Token(); err != nil {
		return err
	}
	if len(seen) != len(fields) {
		return errors.New("incomplete publication metadata")
	}
	return nil
}

// UnmarshalJSON rejects omitted presence/data fields rather than interpreting
// malformed metadata as authority to remove an artifact.
func (a *priorArtifact) UnmarshalJSON(body []byte) error {
	if err := validatePublicationObject(body, []string{"path", "present", "data"}); err != nil {
		return err
	}
	type artifactWire struct {
		Path    string          `json:"path"`
		Present *bool           `json:"present"`
		Data    json.RawMessage `json:"data"`
	}
	var wire artifactWire
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return err
	}
	if wire.Path == "" || wire.Present == nil || len(wire.Data) == 0 {
		return errors.New("incomplete publication artifact metadata")
	}
	var data []byte
	if err := json.Unmarshal(wire.Data, &data); err != nil {
		return err
	}
	*a = priorArtifact{Path: wire.Path, Present: *wire.Present, Data: data}
	return nil
}

type publicationJournal struct {
	Version     int             `json:"version"`
	Ledger      string          `json:"ledger"`
	MemberCount int             `json:"member_count"`
	Artifacts   []priorArtifact `json:"artifacts"`
}

func (j *publicationJournal) UnmarshalJSON(body []byte) error {
	if err := validatePublicationObject(body, []string{"version", "ledger", "member_count", "artifacts"}); err != nil {
		return err
	}
	type journalWire publicationJournal
	var wire journalWire
	if err := json.Unmarshal(body, &wire); err != nil {
		return err
	}
	*j = publicationJournal(wire)
	return nil
}

// canonicalPublicationPath resolves existing ancestors and dangling symlink
// targets even before the first artifact or its parent directory exists.
func canonicalPublicationPath(path string) (string, error) {
	return capture.ProjectDirectory(path)
}

// publicationDestinations derives authority only from caller supplied paths.
// Project/scope validation belongs to the caller; this boundary requires the
// ledger itself, distinct physical members and a complete bounded set.
func publicationDestinations(ledger string, paths []string) (string, []string, error) {
	binding, err := canonicalPublicationPath(ledger)
	if err != nil {
		return "", nil, err
	}
	targets := []string{}
	infos := []os.FileInfo{}
	hasLedger := false
	for _, path := range paths {
		if path == "" {
			continue
		}
		target, e := canonicalPublicationPath(path)
		if e != nil {
			return "", nil, e
		}
		info, e := os.Stat(target)
		if e != nil && !errors.Is(e, os.ErrNotExist) {
			return "", nil, e
		}
		if info != nil && !info.Mode().IsRegular() {
			return "", nil, fmt.Errorf("publication destination is not a regular file: %s", target)
		}
		for i, prior := range targets {
			if target == prior || (info != nil && infos[i] != nil && os.SameFile(info, infos[i])) ||
				strings.HasPrefix(target, prior+string(filepath.Separator)) || strings.HasPrefix(prior, target+string(filepath.Separator)) {
				return "", nil, fmt.Errorf("artifact destinations overlap: %s", target)
			}
		}
		if target == binding {
			hasLedger = true
		}
		targets = append(targets, target)
		infos = append(infos, info)
	}
	if !hasLedger || len(targets) == 0 || len(targets) > 3 {
		return "", nil, errors.New("publication requires ledger and one to three distinct authorized members")
	}
	marker := binding + ".publication-pending"
	markerInfo, e := os.Stat(marker)
	if e != nil && !errors.Is(e, os.ErrNotExist) {
		return "", nil, e
	}
	for i, target := range targets {
		if target == marker || strings.HasPrefix(marker, target+string(filepath.Separator)) ||
			(markerInfo != nil && infos[i] != nil && os.SameFile(markerInfo, infos[i])) {
			return "", nil, errors.New("artifact overlaps publication journal")
		}
	}
	return binding, targets, nil
}

func journalPath(ledger string) string {
	if real, err := canonicalPublicationPath(ledger); err == nil {
		ledger = real
	}
	return ledger + ".publication-pending"
}

// CheckPublication makes interrupted sets visibly unavailable to readers. The
// generator recovers under its per-output lock before reading the baseline.
func CheckPublication(ledger string) error {
	if _, err := os.Lstat(journalPath(ledger)); err == nil {
		return fmt.Errorf("recoverable interrupted artifact publication at %s; rerun generate to restore prior accepted set", ledger)
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}
	return nil
}

func syncDirectory(path string) error {
	f, err := os.Open(path)
	if err != nil {
		return err
	}
	return errors.Join(f.Sync(), f.Close())
}
func durableReplace(path string, data []byte) error {
	f, err := os.CreateTemp(filepath.Dir(path), ".publication-stage-")
	if err != nil {
		return err
	}
	name := f.Name()
	defer func() {
		if e := os.Remove(name); e != nil && !errors.Is(e, os.ErrNotExist) {
			fmt.Fprintln(os.Stderr, "publication staging cleanup:", e)
		}
	}()
	_, err = f.Write(data)
	if e := errors.Join(err, f.Sync(), f.Close()); e != nil {
		return e
	}
	if err = os.Rename(name, path); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(path))
}

// BeginPublication persists a bounded prior-good set before any visible writes.
// This is a recoverable transaction, not portable atomic multi-file rename.
// Caller must hold the canonical ledger publication lock through commit/recovery.
func BeginPublication(ledger string, paths []string) error {
	if err := CheckPublication(ledger); err != nil {
		return err
	}
	binding, targets, err := publicationDestinations(ledger, paths)
	if err != nil {
		return err
	}
	j := publicationJournal{Version: 1, Ledger: binding, MemberCount: len(targets)}
	total := 0
	for _, target := range targets {
		a := priorArtifact{Path: target}
		f, err := os.Open(target)
		if err == nil {
			a.Present = true
			a.Data, err = io.ReadAll(io.LimitReader(f, int64(publicationLimit-total+1)))
			err = errors.Join(err, f.Close())
		} else if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		if err != nil {
			return err
		}
		total += len(a.Data)
		if total > publicationLimit {
			return errors.New("publication prior-good set exceeds 128 MiB recovery bound")
		}
		j.Artifacts = append(j.Artifacts, a)
	}
	body, err := json.Marshal(j)
	if err != nil {
		return err
	}
	return durableReplace(journalPath(ledger), body)
}

// RecoverPublication rolls back every member; the journal remains until all
// restoration writes are durable, so a second crash repeats recovery safely.
// authorizedPaths must come from trusted invocation/configuration after project
// validation, never from the journal. Missing/legacy authorization fails closed.
// Caller must hold the canonical ledger publication lock throughout recovery.
func RecoverPublication(ledger string, authorizedPaths []string) (bool, error) {
	marker := journalPath(ledger)
	info, err := os.Lstat(marker)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	if !info.Mode().IsRegular() {
		return true, errors.New("invalid publication journal file; explicit repair required")
	}
	f, err := os.Open(marker)
	if errors.Is(err, os.ErrNotExist) {
		return false, nil
	}
	if err != nil {
		return false, err
	}
	body, err := io.ReadAll(io.LimitReader(f, 2*publicationLimit+1))
	if e := errors.Join(err, f.Close()); e != nil {
		return true, e
	}
	if len(body) > 2*publicationLimit {
		return true, errors.New("publication recovery journal exceeds bound")
	}
	// Validate every member before any restoration or deletion. The journal is
	// untrusted input, including when permissions happen to be 0600.
	binding, allowed, err := publicationDestinations(ledger, authorizedPaths)
	if err != nil {
		return true, fmt.Errorf("publication recovery authorization; explicit repair required: %w", err)
	}
	var j publicationJournal
	decoder := json.NewDecoder(bytes.NewReader(body))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&j); err != nil {
		return true, fmt.Errorf("invalid publication journal; explicit repair required: %w", err)
	}
	if _, err = decoder.Token(); !errors.Is(err, io.EOF) {
		return true, errors.New("invalid trailing publication journal data; explicit repair required")
	}
	if j.Version != 1 || j.Ledger != binding || j.MemberCount != len(allowed) || len(j.Artifacts) != len(allowed) {
		return true, errors.New("invalid publication version, ledger binding or complete member count; explicit repair required")
	}
	seen := map[string]bool{}
	total := 0
	for _, a := range j.Artifacts {
		target, e := canonicalPublicationPath(a.Path)
		if e != nil {
			return true, fmt.Errorf("invalid publication destination; explicit repair required: %w", e)
		}
		authorized := false
		for _, path := range allowed {
			if path == target {
				authorized = true
			}
		}
		if !filepath.IsAbs(a.Path) || a.Path != target || !authorized || seen[target] {
			return true, errors.New("unauthorized, relocated or duplicate publication destination; explicit repair required")
		}
		seen[target] = true
		total += len(a.Data)
		if total > publicationLimit || (!a.Present && len(a.Data) != 0) {
			return true, errors.New("invalid publication prior-good data; explicit repair required")
		}
	}
	for _, a := range j.Artifacts {
		if a.Present {
			err = durableReplace(a.Path, a.Data)
		} else {
			err = os.Remove(a.Path)
			if errors.Is(err, os.ErrNotExist) {
				err = nil
			}
			if err == nil {
				if _, e := os.Stat(filepath.Dir(a.Path)); !errors.Is(e, os.ErrNotExist) {
					err = syncDirectory(filepath.Dir(a.Path))
				}
			}
		}
		if err != nil {
			return true, fmt.Errorf("restoring prior-good artifact %s: %w", a.Path, err)
		}
	}
	return true, CommitPublication(ledger, nil)
}

// CommitPublication removes the marker only after all new files and their
// directories are synced. Failure retains the prior-good journal for recovery.
func CommitPublication(ledger string, paths []string) error {
	for _, path := range paths {
		if path == "" {
			continue
		}
		f, err := os.Open(path)
		if err != nil {
			return err
		}
		if err = errors.Join(f.Sync(), f.Close()); err != nil {
			return err
		}
		if err = syncDirectory(filepath.Dir(path)); err != nil {
			return err
		}
	}
	if err := os.Remove(journalPath(ledger)); err != nil {
		return err
	}
	return syncDirectory(filepath.Dir(journalPath(ledger)))
}

func readPublishedLedger(path string) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	lock, err := capture.AcquirePublication(ctx, path)
	if err != nil {
		return nil, err
	}
	defer func() {
		if e := lock.Close(); e != nil {
			fmt.Fprintln(os.Stderr, "publication read unlock:", e)
		}
	}()
	return readOwnedLedger(path)
}
func readOwnedLedger(path string) ([]byte, error) {
	if err := CheckPublication(path); err != nil {
		return nil, err
	}
	body, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	if err = CheckPublication(path); err != nil {
		return nil, err
	}
	return body, nil
}
