package codex_test

import (
	"compress/gzip"
	"context"
	"errors"
	"fmt"
	"math/rand/v2"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"agenticlens/internal/adapters/codex"
	"agenticlens/internal/adapters/codex/codexstoretest"
)

// FIX-21 capture-only discovery pre-filters. File names only pre-filter
// candidates; identity always comes from session_meta. Manual discovery is
// covered, unchanged, by discovery_test.go.

const windowHeaderChunk = 4096

var windowParent = time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)

func windowUUID(at time.Time, version byte, seed uint64) string {
	return codexstoretest.UUID(at, version, rand.New(rand.NewPCG(seed, seed)))
}

func windowName(at time.Time, id string) string {
	return "rollout-" + at.Format("2006-01-02T15-04-05") + "-" + id + ".jsonl"
}

func TestCreationWindowFileRule(t *testing.T) {
	window := codex.NewCreationWindow(windowUUID(windowParent, 7, 1), codex.RolloutHeader{})
	if !window.Prunes() {
		t.Fatal("a UUIDv7 parent must give the window an instant")
	}
	for _, tc := range []struct {
		name  string
		file  string
		skips bool
	}{
		{"child created after the parent", windowName(windowParent.Add(7*time.Second), windowUUID(windowParent.Add(7*time.Second), 7, 2)), false},
		{"child created at the parent instant", windowName(windowParent, windowUUID(windowParent, 7, 3)), false},
		{"child earlier within the tolerance", windowName(windowParent, windowUUID(windowParent.Add(-4999*time.Millisecond), 7, 4)), false},
		{"child exactly at the tolerance", windowName(windowParent, windowUUID(windowParent.Add(-codex.ChildCreationTolerance), 7, 5)), false},
		{"candidate just beyond the tolerance", windowName(windowParent, windowUUID(windowParent.Add(-5001*time.Millisecond), 7, 6)), true},
		{"candidate created a month earlier", windowName(windowParent.AddDate(0, -1, 0), windowUUID(windowParent.AddDate(0, -1, 0), 7, 7)), true},
		{"compressed candidate created earlier", windowName(windowParent.AddDate(0, -1, 0), windowUUID(windowParent.AddDate(0, -1, 0), 7, 8)) + ".gz", true},
		// The UUID decides, not the local stamp in the name.
		{"earlier UUID behind a later local stamp", windowName(windowParent.Add(time.Hour), windowUUID(windowParent.AddDate(0, 0, -2), 7, 9)), true},
		{"later UUID behind an earlier local stamp", windowName(windowParent.AddDate(0, 0, -2), windowUUID(windowParent.Add(time.Minute), 7, 10)), false},
		{"non-v7 UUID is always read", windowName(windowParent.AddDate(0, -6, 0), windowUUID(windowParent.AddDate(0, -6, 0), 4, 11)), false},
		{"name without a UUID is always read", "rollout-2026-05-10T09-15-00.jsonl", false},
		{"custom name is always read", "custom-export.jsonl", false},
		{"truncated UUID is always read", "rollout-2026-05-10T09-15-00-0190f6c2-0000-7000-8000.jsonl", false},
		{"invalid stamp is always read", "rollout-2026-13-10T09-15-00-" + windowUUID(windowParent.AddDate(0, -6, 0), 7, 12) + ".jsonl", false},
		{"non-rollout suffix is always read", strings.TrimSuffix(windowName(windowParent.AddDate(0, -1, 0), windowUUID(windowParent.AddDate(0, -1, 0), 7, 13)), ".jsonl") + ".json", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := window.SkipsFile(tc.file); got != tc.skips {
				t.Fatalf("SkipsFile(%s) = %t, want %t", tc.file, got, tc.skips)
			}
		})
	}
}

func TestCreationWindowParentInstant(t *testing.T) {
	old := windowName(windowParent.AddDate(0, -1, 0), windowUUID(windowParent.AddDate(0, -1, 0), 7, 20))
	recent := windowName(windowParent.Add(-time.Hour), windowUUID(windowParent.Add(-time.Hour), 7, 21))
	for _, tc := range []struct {
		name         string
		parent       string
		header       codex.RolloutHeader
		prunes       bool
		old, recent  bool
		directory    bool
		directoryRel string
	}{
		// The parent's own UUIDv7 decides even when its header disagrees.
		{"UUIDv7 parent", windowUUID(windowParent, 7, 22), codex.RolloutHeader{CreatedAt: windowParent.AddDate(0, -2, 0)}, true, true, true, true, "sessions/2026/10/01"},
		{"session_meta timestamp of a non-v7 parent", "named-parent", codex.RolloutHeader{CreatedAt: windowParent}, true, true, true, true, "sessions/2026/10/01"},
		{"session_meta timestamp of a non-v7 UUID parent", windowUUID(windowParent, 4, 23), codex.RolloutHeader{CreatedAt: windowParent}, true, true, true, true, "sessions/2026/10/01"},
		{"no parent instant disables pruning", "named-parent", codex.RolloutHeader{}, false, false, false, false, "sessions/2020/01/01"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			window := codex.NewCreationWindow(tc.parent, tc.header)
			if window.Prunes() != tc.prunes || window.SkipsFile(old) != tc.old || window.SkipsFile(recent) != tc.recent {
				t.Fatalf("prunes=%t old=%t recent=%t, want %t %t %t", window.Prunes(), window.SkipsFile(old), window.SkipsFile(recent), tc.prunes, tc.old, tc.recent)
			}
			home := "/fixture/home"
			if got := window.SkipsDirectory(home, filepath.Join(home, filepath.FromSlash(tc.directoryRel))); got != tc.directory {
				t.Fatalf("SkipsDirectory(%s) = %t, want %t", tc.directoryRel, got, tc.directory)
			}
		})
	}
	var zero codex.CreationWindow
	if zero.Prunes() || zero.SkipsFile(old) || zero.SkipsDirectory("/h", "/h/sessions/2000") {
		t.Fatal("the zero window must prune nothing")
	}
}

func TestCreationWindowDirectoryRule(t *testing.T) {
	home := "/fixture/home"
	dir := func(relative string) string { return filepath.Join(home, filepath.FromSlash(relative)) }
	zone := func(name string) *time.Location {
		location, err := time.LoadLocation(name)
		if err != nil {
			t.Fatal(err)
		}
		return location
	}
	// A child created shortly after its parent sits in the machine-local day
	// directory of its own creation, which can differ from the parent's UTC date.
	for _, tc := range []struct {
		name     string
		location *time.Location
		parent   time.Time
		delay    time.Duration
	}{
		{"CET after local midnight", zone("Europe/Berlin"), time.Date(2026, 11, 24, 23, 59, 55, 0, zone("Europe/Berlin")), 7 * time.Second},
		{"CEST after local midnight and month end", zone("Europe/Berlin"), time.Date(2026, 6, 30, 23, 59, 55, 0, zone("Europe/Berlin")), 7 * time.Second},
		{"CET year end", zone("Europe/Berlin"), time.Date(2026, 12, 31, 23, 59, 55, 0, zone("Europe/Berlin")), 7 * time.Second},
		{"UTC-12 local date behind the UTC date", zone("Etc/GMT+12"), time.Date(2026, 11, 23, 23, 59, 59, 0, zone("Etc/GMT+12")), 0},
		{"UTC+14 local date ahead of the UTC date", zone("Pacific/Kiritimati"), time.Date(2026, 11, 24, 23, 59, 55, 0, zone("Pacific/Kiritimati")), 7 * time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			window := codex.NewCreationWindow(windowUUID(tc.parent, 7, 30), codex.RolloutHeader{})
			child := tc.parent.Add(tc.delay).In(tc.location)
			for _, relative := range []string{child.Format("sessions/2006"), child.Format("sessions/2006/01"), child.Format("sessions/2006/01/02"),
				tc.parent.In(tc.location).Format("sessions/2006/01/02"), tc.parent.UTC().Format("sessions/2006/01/02")} {
				if window.SkipsDirectory(home, dir(relative)) {
					t.Fatalf("%s would hide a child created %s after its parent", relative, tc.delay)
				}
			}
			if !window.SkipsFile(windowName(child, windowUUID(tc.parent.Add(-time.Minute), 7, 31))) {
				t.Fatal("the file rule must still apply inside a listed directory")
			}
			// Day directories that end more than the maximum UTC offset before
			// the parent's instant are not listed.
			stale := tc.parent.UTC().AddDate(0, 0, -2).Format("sessions/2006/01/02")
			if !window.SkipsDirectory(home, dir(stale)) {
				t.Fatalf("%s ends more than 14 h before the parent and must be skipped", stale)
			}
		})
	}

	parent := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	window := codex.NewCreationWindow(windowUUID(parent, 7, 40), codex.RolloutHeader{})
	margin := codex.MaxLocalDayOffset + codex.ChildCreationTolerance
	if margin != 14*time.Hour+5*time.Second {
		t.Fatalf("directory margin = %s, want 14h5s", margin)
	}
	// The local day 2026-11-23 ends at 2026-11-24T00:00Z read as UTC.
	dayEnd := time.Date(2026, 11, 24, 0, 0, 0, 0, time.UTC)
	for _, tc := range []struct {
		name   string
		parent time.Time
		path   string
		skips  bool
	}{
		{"day ending exactly at the margin is listed", dayEnd.Add(margin), "sessions/2026/11/23", false},
		{"day ending just beyond the margin is skipped", dayEnd.Add(margin + time.Millisecond), "sessions/2026/11/23", true},
		{"month ending beyond the margin is skipped", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Add(margin + time.Millisecond), "sessions/2026/10", true},
		{"month ending within the margin is listed", time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC).Add(margin), "sessions/2026/10", false},
		{"year ending beyond the margin is skipped", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC).Add(margin + time.Millisecond), "sessions/2026", true},
		{"year ending within the margin is listed", time.Date(2027, 1, 1, 0, 0, 0, 0, time.UTC).Add(margin), "sessions/2026", false},
		{"sessions root is listed", parent, "sessions", false},
		{"archived sessions are always listed", parent, "archived_sessions/2020/01/01", false},
		{"non-date directory is always listed", parent, "sessions/notes", false},
		{"one-digit month is always listed", parent, "sessions/2020/1", false},
		{"invalid month is always listed", parent, "sessions/2020/13", false},
		{"one-digit day is always listed", parent, "sessions/2020/01/1", false},
		{"nested below a day is always listed", parent, "sessions/2020/01/01/extra", false},
		{"short year is always listed", parent, "sessions/20", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			check := window
			if tc.parent != parent {
				check = codex.NewCreationWindow(windowUUID(tc.parent, 7, 41), codex.RolloutHeader{})
			}
			if got := check.SkipsDirectory(home, dir(tc.path)); got != tc.skips {
				t.Fatalf("SkipsDirectory(%s) = %t, want %t", tc.path, got, tc.skips)
			}
		})
	}
	if window.SkipsDirectory(home, "/elsewhere/sessions/2020/01/01") {
		t.Fatal("a directory outside the store's sessions tree must be listed")
	}
}

// windowRollout writes a rollout whose first line is a session_meta of
// exactly headerBytes and returns its size.
func windowRollout(t *testing.T, path, id string, headerBytes, size int) int64 {
	t.Helper()
	prefix := fmt.Sprintf(`{"timestamp":"2026-11-24T09:00:00.000Z","type":"session_meta","payload":{"id":%q,"timestamp":"2026-11-24T09:00:00.000Z","cwd":"/fixture/project","source":"cli","base_instructions":{"text":"`, id)
	suffix := "\"}}}\n"
	padding := headerBytes - len(prefix) - len(suffix)
	if padding < 0 {
		t.Fatalf("header %d too small", headerBytes)
	}
	body := prefix + strings.Repeat("x", padding) + suffix
	turn := `{"timestamp":"2026-11-24T09:00:01Z","type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}` + "\n"
	for len(body) < size {
		body += turn
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
		t.Fatal(err)
	}
	return int64(len(body))
}

// windowHeaderCost is what a header read consumes: whole reads of the read
// chunk until the first line ends, never past the file or the 64 KiB bound.
func windowHeaderCost(headerBytes int, size int64) int64 {
	chunks := (int64(headerBytes) + windowHeaderChunk - 1) / windowHeaderChunk
	return min(size, chunks*windowHeaderChunk, 64*1024)
}

func TestDiscoverExactCaptureRolloutPathSkipsOtherUUIDs(t *testing.T) {
	const header = 29 * 1024
	home := t.TempDir()
	owner := windowUUID(windowParent, 7, 50)
	live := filepath.Join(home, "sessions", "2026", "11", "24", windowName(windowParent, owner))
	archived := filepath.Join(home, "archived_sessions", windowName(windowParent, owner))
	ownerSize := windowRollout(t, live, owner, header, 200*1024)
	// Many conventional rollouts of other sessions, including a non-v7 name.
	for i := 0; i < 40; i++ {
		at := windowParent.Add(time.Duration(i-20) * time.Hour)
		version := byte(7)
		if i == 3 {
			version = 4
		}
		id := windowUUID(at, version, uint64(100+i))
		windowRollout(t, filepath.Join(home, "sessions", at.Format("2006/01/02"), windowName(at, id)), id, header, 96*1024)
	}
	// A name outside the convention is always read; its metadata names
	// another session.
	custom := filepath.Join(home, "sessions", "2026", "11", "24", "custom-export.jsonl")
	customSize := windowRollout(t, custom, "other-session", header, 128*1024)

	lookup := func(t *testing.T, id string) (string, *codex.DiscoveryBudget, error) {
		t.Helper()
		budget := &codex.DiscoveryBudget{MaxSessions: 1 << 20, MaxSourceBytes: 1 << 40}
		path, err := codex.DiscoverExactCaptureRolloutPath(home, id, budget, codex.NewHeaderCache())
		return path, budget, err
	}
	t.Run("unique owner reads only the matching and unconventional names", func(t *testing.T) {
		path, budget, err := lookup(t, owner)
		if err != nil || path != live {
			t.Fatalf("lookup = %q, %v; want %s", path, err, live)
		}
		want := windowHeaderCost(header, ownerSize) + windowHeaderCost(header, customSize)
		if budget.SessionsRead != 2 || budget.SourceBytesRead != want {
			t.Fatalf("claims=%d bytes=%d, want 2 claims and %d bytes", budget.SessionsRead, budget.SourceBytesRead, want)
		}
	})
	t.Run("archived same-name copy is found", func(t *testing.T) {
		if err := os.MkdirAll(filepath.Dir(archived), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Rename(live, archived); err != nil {
			t.Fatal(err)
		}
		defer func() {
			if err := os.Rename(archived, live); err != nil {
				t.Fatal(err)
			}
		}()
		if path, _, err := lookup(t, owner); err != nil || path != archived {
			t.Fatalf("lookup = %q, %v; want the archived copy %s", path, err, archived)
		}
	})
	t.Run("duplicate same-UUID copies stay ambiguous", func(t *testing.T) {
		copySize := windowRollout(t, archived, owner, header, 200*1024)
		defer os.Remove(archived)
		_, budget, err := lookup(t, owner)
		if !errors.Is(err, codex.ErrOwnerAmbiguous) || !strings.Contains(err.Error(), "ambiguous rollout paths") {
			t.Fatalf("duplicate copies not ambiguous: %v", err)
		}
		// Sorted paths read the archived copy, then the unconventional name,
		// then the live copy, which proves the ambiguity. No other UUID is read.
		want := windowHeaderCost(header, copySize) + windowHeaderCost(header, customSize) + windowHeaderCost(header, ownerSize)
		if budget.SessionsRead != 3 || budget.SourceBytesRead != want {
			t.Fatalf("claims=%d bytes=%d, want 3 claims and %d bytes", budget.SessionsRead, budget.SourceBytesRead, want)
		}
	})
	t.Run("metadata decides an unconventional name", func(t *testing.T) {
		path, _, err := lookup(t, "other-session")
		if err != nil || path != custom {
			t.Fatalf("lookup = %q, %v; want the custom-named rollout", path, err)
		}
	})
	t.Run("a matching name whose metadata disagrees is not the owner", func(t *testing.T) {
		mismatch := windowUUID(windowParent.Add(time.Minute), 7, 60)
		named := filepath.Join(home, "sessions", "2026", "11", "24", windowName(windowParent, mismatch))
		windowRollout(t, named, "someone-else", header, 64*1024)
		defer os.Remove(named)
		_, budget, err := lookup(t, mismatch)
		if !errors.Is(err, codex.ErrOwnerUnavailable) || budget.SessionsRead != 2 {
			t.Fatalf("lookup = %v after %d claims; want unavailable after reading the named file and the custom name", err, budget.SessionsRead)
		}
	})
	t.Run("an empty store is unavailable", func(t *testing.T) {
		empty := t.TempDir()
		_, err := codex.DiscoverExactCaptureRolloutPath(empty, owner, &codex.DiscoveryBudget{MaxSessions: 4, MaxSourceBytes: 1 << 20}, nil)
		if !errors.Is(err, codex.ErrOwnerUnavailable) {
			t.Fatalf("empty store: %v", err)
		}
	})
	t.Run("live exact lookup still reads every rollout", func(t *testing.T) {
		budget := &codex.DiscoveryBudget{MaxSessions: 1 << 20, MaxSourceBytes: 1 << 40}
		if _, err := codex.DiscoverExactRolloutPathWithBudget(home, owner, budget); err != nil {
			t.Fatal(err)
		}
		if budget.SessionsRead != 42 {
			t.Fatalf("live lookup claimed %d rollouts, want all 42", budget.SessionsRead)
		}
	})
}

func TestHeaderCacheReadsOncePerFileStateAndChargesActualBytes(t *testing.T) {
	const header = 29 * 1024
	ctx := context.Background()
	path := filepath.Join(t.TempDir(), "rollout.jsonl")
	size := windowRollout(t, path, "cached-owner", header, 200*1024)
	cache := codex.NewHeaderCache()
	budget := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 30}

	first, err := cache.ReadHeader(ctx, path, budget)
	if err != nil || first.Cached || first.Header.ID != "cached-owner" || !first.Header.First {
		t.Fatalf("first read = %+v, %v", first, err)
	}
	if want := windowHeaderCost(header, size); first.Bytes != want || budget.SourceBytesRead != want {
		t.Fatalf("first read charged %d (budget %d), want the %d bytes actually read", first.Bytes, budget.SourceBytesRead, want)
	}
	if budget.SessionsRead != 0 {
		t.Fatal("a header read must not claim a session slot")
	}
	again, err := cache.ReadHeader(ctx, path, budget)
	if err != nil || !again.Cached || again.Bytes != 0 || again.Header != first.Header || budget.SourceBytesRead != first.Bytes {
		t.Fatalf("second read in the same file state = %+v, %v (budget %d)", again, err, budget.SourceBytesRead)
	}

	// A size change, or a modification time change at the same size, is a
	// new file state and is read again.
	file, err := os.OpenFile(path, os.O_APPEND|os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := file.WriteString(`{"type":"turn_context","payload":{}}` + "\n"); err != nil {
		t.Fatal(err)
	}
	if err := file.Close(); err != nil {
		t.Fatal(err)
	}
	grown, err := cache.ReadHeader(ctx, path, budget)
	if err != nil || grown.Cached || grown.Bytes == 0 {
		t.Fatalf("grown file served from the cache: %+v, %v", grown, err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chtimes(path, info.ModTime().Add(time.Second), info.ModTime().Add(time.Second)); err != nil {
		t.Fatal(err)
	}
	touched, err := cache.ReadHeader(ctx, path, budget)
	if err != nil || touched.Cached || touched.Bytes == 0 {
		t.Fatalf("touched file served from the cache: %+v, %v", touched, err)
	}

	t.Run("a small rollout costs its size", func(t *testing.T) {
		small := filepath.Join(t.TempDir(), "small.jsonl")
		smallSize := windowRollout(t, small, "small-owner", 512, 1500)
		read, err := codex.NewHeaderCache().ReadHeader(ctx, small, &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 20})
		if err != nil || read.Bytes != smallSize {
			t.Fatalf("small read = %+v, %v; want %d bytes", read, err, smallSize)
		}
	})
	t.Run("a read starts only when its bound fits", func(t *testing.T) {
		fresh := codex.NewHeaderCache()
		bound := min(size+int64(len(`{"type":"turn_context","payload":{}}`)+1), 64*1024)
		tight := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: bound - 1}
		read, err := fresh.ReadHeader(ctx, path, tight)
		if !errors.Is(err, codex.ErrDiscoveryBudget) || read.Bytes != 0 || tight.SourceBytesRead != 0 {
			t.Fatalf("read below its bound = %+v, %v (budget %d), want nothing read", read, err, tight.SourceBytesRead)
		}
		exact := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: bound}
		read, err = fresh.ReadHeader(ctx, path, exact)
		if err != nil || read.Cached || read.Bytes != windowHeaderCost(header, size) {
			t.Fatalf("a refused read was cached or the exact bound failed: %+v, %v", read, err)
		}
	})
	t.Run("a nil cache reads every request", func(t *testing.T) {
		var none *codex.HeaderCache
		budget := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 30}
		for i := 0; i < 2; i++ {
			if read, err := none.ReadHeader(ctx, path, budget); err != nil || read.Cached {
				t.Fatalf("nil cache read %d = %+v, %v", i, read, err)
			}
		}
		if budget.SourceBytesRead != 2*windowHeaderCost(header, size) {
			t.Fatalf("nil cache charged %d", budget.SourceBytesRead)
		}
	})
}

// FIX-21 ARCH-04: a compressed rollout's charge counts decompressed bytes,
// which its file size does not bound. Its header read starts only when the
// whole 64 KiB structural bound fits, so it is never abandoned part way.
func TestHeaderCacheAdmitsCompressedHeaderOnlyWithTheWholeBound(t *testing.T) {
	ctx := context.Background()
	const header = 48 * 1024
	plain := filepath.Join(t.TempDir(), "rollout.jsonl")
	decompressed := windowRollout(t, plain, "compressed-owner", header, header+512)
	body, err := os.ReadFile(plain)
	if err != nil {
		t.Fatal(err)
	}
	path := plain + ".gz"
	file, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	writer := gzip.NewWriter(file)
	if _, err := writer.Write(body); err != nil {
		t.Fatal(err)
	}
	if err := errors.Join(writer.Close(), file.Close()); err != nil {
		t.Fatal(err)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	cost := windowHeaderCost(header, decompressed)
	// The allowance covers the compressed file and is just short of the
	// decompressed header, so a read admitted by file size would stop part way.
	allowance := cost - 1
	if info.Size() >= allowance || allowance >= 64*1024 {
		t.Fatalf("fixture compressed size %d and allowance %d do not separate the bounds", info.Size(), allowance)
	}
	cache := codex.NewHeaderCache()
	short := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: allowance}
	read, err := cache.ReadHeader(ctx, path, short)
	if !errors.Is(err, codex.ErrDiscoveryBudget) || read.Bytes != 0 || short.SourceBytesRead != 0 {
		t.Fatalf("compressed header under its bound = %+v, %v, charged %d; want nothing read", read, err, short.SourceBytesRead)
	}
	almost := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 64*1024 - 1}
	if read, err := cache.ReadHeader(ctx, path, almost); !errors.Is(err, codex.ErrDiscoveryBudget) || almost.SourceBytesRead != 0 {
		t.Fatalf("compressed header one byte under 64 KiB = %+v, %v", read, err)
	}
	whole := &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 64 * 1024}
	read, err = cache.ReadHeader(ctx, path, whole)
	if err != nil || read.Header.ID != "compressed-owner" || read.Bytes != cost || whole.SourceBytesRead != cost {
		t.Fatalf("compressed header with the whole bound = %+v, %v, charged %d; want %d decompressed bytes", read, err, whole.SourceBytesRead, cost)
	}
}

// FIX-21 R2-T5: the header path applies the same exact-owner rules as the
// session-context path. Both ancestry spellings must agree when both exist.
func TestHeaderSessionContextRejectsConflictingAncestry(t *testing.T) {
	ctx := context.Background()
	for _, tc := range []struct {
		name, direct, nested, parent string
		conflict                     bool
	}{
		{name: "direct and nested parents disagree", direct: "parent-a", nested: "parent-b", conflict: true},
		{name: "direct and nested parents agree", direct: "parent-a", nested: "parent-a", parent: "parent-a"},
		{name: "nested parent only", nested: "parent-b", parent: "parent-b"},
		{name: "direct parent only", direct: "parent-a", parent: "parent-a"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "child.jsonl")
			source := `"cli"`
			if tc.nested != "" {
				source = fmt.Sprintf(`{"subagent":{"thread_spawn":{"parent_thread_id":%q,"depth":1}}}`, tc.nested)
			}
			direct := ""
			if tc.direct != "" {
				direct = fmt.Sprintf(`"parent_thread_id":%q,`, tc.direct)
			}
			body := fmt.Sprintf(`{"timestamp":"2026-11-24T09:00:00Z","type":"session_meta","payload":{"id":"child",%s"cwd":"/fixture/project","source":%s}}`+"\n", direct, source)
			if err := os.WriteFile(path, []byte(body), 0o600); err != nil {
				t.Fatal(err)
			}
			read, err := codex.NewHeaderCache().ReadHeader(ctx, path, &codex.DiscoveryBudget{MaxSessions: 1, MaxSourceBytes: 1 << 20})
			if err != nil || read.Header.ConflictingParent != tc.conflict {
				t.Fatalf("header = %+v, %v; want conflict %t", read.Header, err, tc.conflict)
			}
			metadata, err := read.Header.SessionContext()
			if tc.conflict {
				if !errors.Is(err, codex.ErrSessionContextMismatch) {
					t.Fatalf("conflicting ancestry accepted: %+v, %v", metadata, err)
				}
				return
			}
			if err != nil || metadata.ID != "child" || metadata.ParentID != tc.parent {
				t.Fatalf("session context = %+v, %v; want parent %q", metadata, err, tc.parent)
			}
		})
	}
}
