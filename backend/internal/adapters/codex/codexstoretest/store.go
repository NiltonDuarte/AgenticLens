// Package codexstoretest builds synthetic Codex source stores for tests.
//
// It is test support only: it imports nothing outside the standard library,
// and no production package may import it. Every store is written under a
// directory that the caller's test owns, normally t.TempDir(); nothing here
// reads ~/.codex or any ledger.
package codexstoretest

import (
	"bufio"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"math/rand/v2"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"
	"time"
	// Machine-local rollout names need a real CET/CEST zone on every host.
	_ "time/tzdata"
)

// The large synthetic Codex store mirrors the shape of a real long-lived
// Codex store:
//   - the first line is session_meta, padded through base_instructions to a
//     configured header size (about 29 KB in the observed store);
//   - live names are sessions/YYYY/MM/DD/rollout-YYYY-MM-DDThh-mm-ss-<uuid>.jsonl
//     in machine-local time, while the UUIDv7 encodes the UTC creation ms;
//   - archived rollouts sit flat in archived_sessions/ with the same names;
//   - a few names break the convention, one carries a non-v7 UUID, one names a
//     different UUID than its metadata and one is gzip-compressed;
//   - one unrelated session has a same-UUID copy in archived_sessions/;
//   - one root, created after every unrelated rollout, delegates through
//     task-name-only spawn_agent calls; its children carry the parent only in
//     session_meta.source.subagent.thread_spawn.parent_thread_id.
//
// Unrelated rollouts larger than a header can be sparse: the session_meta line
// and two real records are written, and the file is then extended with a hole
// to its apparent size. Capture header consumers never read past the first
// line of an unrelated rollout, so the hole only sets the byte size that
// budgets and stat observe. Family rollouts are always dense, parseable JSONL.

const (
	// HeaderBytes is the observed session_meta line size, newline included.
	HeaderBytes  = 29 * 1024
	paddingText  = "Synthetic base instructions for a self-contained capture fixture. "
	responseText = "Synthetic assistant output that only gives the rollout a realistic body size. "
)

// LargeStoreSpec configures WriteLargeStore.
type LargeStoreSpec struct {
	// Location is the machine-local zone used for directory and file names.
	Location          *time.Location
	Seed              uint64
	Project           string
	LiveUnrelated     int
	ArchivedUnrelated int
	HeaderBytes       int
	// SmallEvery makes every n-th unrelated rollout a dense file smaller than
	// the 64 KiB header bound. Zero disables small rollouts.
	SmallEvery       int
	MinUnrelated     int64
	MaxUnrelated     int64
	SparseUnrelated  bool
	UnrelatedFrom    time.Time
	UnrelatedThrough time.Time
	RootCreated      time.Time
	RootBytes        int64
	ChildBytes       int64
	// ChildDelays are the children's creation offsets after the root.
	ChildDelays []time.Duration
	// LaterUnrelated rollouts of other sessions are created after the root,
	// spread evenly over LaterSpan, so they fall inside a child sweep's window.
	LaterUnrelated int
	LaterSpan      time.Duration
	// WindowIrregular adds a name without a UUID and a non-v7 name to the
	// root's own day directory, which a child sweep always lists.
	WindowIrregular bool
}

// File describes one generated rollout.
type File struct {
	Path        string
	SessionID   string
	ParentID    string
	Size        int64
	HeaderBytes int64
	Archived    bool
	Family      bool
	Compressed  bool
	// Later marks an unrelated rollout created after the root.
	Later bool
	// Conventional is true when the name matches the rollout naming
	// convention with a version 7 UUID equal to the session_meta ID.
	Conventional bool
	Note         string
}

// LargeStore is a generated store.
type LargeStore struct {
	Home        string
	Spec        LargeStoreSpec
	Files       []File
	RootID      string
	RootPath    string
	ChildIDs    []string
	ChildPaths  []string
	DuplicateID string
}

// Irregular-name notes. Each is generated exactly once per store.
const (
	NoteNoUUID     = "name without UUID"
	NoteCustomName = "name outside convention"
	NoteNonDateDir = "non-date directory"
	NoteNonV7      = "non-v7 UUID"
	NoteMismatch   = "name UUID differs from metadata"
	NoteGzip       = "gzip-compressed"
	NoteDuplicate  = "same-UUID copy"
	// NoteWindowNoUUID and NoteWindowNonV7 sit in the root's day directory.
	NoteWindowNoUUID = "window name without UUID"
	NoteWindowNonV7  = "window non-v7 UUID"
)

// IncidentScaleSpec reproduces the AL-71 store: at divisor 1, about 1,900
// rollouts of about 3 GB with a 50 MB family against a 512 MiB budget. A
// larger divisor shrinks the rollout count and the family sizes together, so
// the ratios stay the same: header sizes, the per-file size distribution and
// the family member count do not change.
func IncidentScaleSpec(project string, divisor int) LargeStoreSpec {
	berlin, err := time.LoadLocation("Europe/Berlin")
	if err != nil {
		panic(err)
	}
	return LargeStoreSpec{
		Location:          berlin,
		Seed:              21,
		Project:           project,
		LiveUnrelated:     1490 / divisor,
		ArchivedUnrelated: 400 / divisor,
		HeaderBytes:       HeaderBytes,
		SmallEvery:        10,
		MinUnrelated:      96 * 1024,
		MaxUnrelated:      3200 * 1024,
		SparseUnrelated:   true,
		UnrelatedFrom:     time.Date(2026, 3, 1, 0, 0, 0, 0, berlin),
		UnrelatedThrough:  time.Date(2026, 11, 23, 23, 0, 0, 0, berlin),
		RootCreated:       time.Date(2026, 11, 24, 10, 0, 0, 0, berlin),
		RootBytes:         (12 << 20) / int64(divisor),
		ChildBytes:        (4 << 20) / int64(divisor),
		ChildDelays: []time.Duration{
			7 * time.Second, 30 * time.Second, 2 * time.Minute, 5 * time.Minute, 10 * time.Minute,
			15 * time.Minute, 25 * time.Minute, 35 * time.Minute, 50 * time.Minute,
		},
	}
}

// WriteLargeStore generates spec under a new t.TempDir().
func WriteLargeStore(t testing.TB, spec LargeStoreSpec) LargeStore {
	t.Helper()
	home, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	store := LargeStore{Home: home, Spec: spec}
	rnd := rand.New(rand.NewPCG(spec.Seed, spec.Seed^0x9e3779b97f4a7c15))
	span := spec.UnrelatedThrough.Sub(spec.UnrelatedFrom)
	unrelated := spec.LiveUnrelated + spec.ArchivedUnrelated
	for i := 0; i < unrelated; i++ {
		created := spec.UnrelatedFrom.Add(time.Duration(rnd.Int64N(int64(span)))).Truncate(time.Millisecond)
		id := UUID(created, 7, rnd)
		archived := i >= spec.LiveUnrelated
		size := spec.MinUnrelated + rnd.Int64N(spec.MaxUnrelated-spec.MinUnrelated+1)
		dense := !spec.SparseUnrelated
		if spec.SmallEvery > 0 && i%spec.SmallEvery == 0 {
			size = int64(spec.HeaderBytes) + 512 + rnd.Int64N(int64(64*1024-spec.HeaderBytes-1024))
			dense = true
		}
		file := File{Path: ConventionalPath(home, created, id, archived), SessionID: id, Archived: archived, Conventional: true}
		store.Files = append(store.Files, writeUnrelatedRollout(t, spec, file, created, size, dense))
	}

	// One unrelated live session also exists as a same-UUID archived copy.
	original := store.Files[1]
	copied := original
	copied.Path = filepath.Join(home, "archived_sessions", filepath.Base(original.Path))
	copied.Archived = true
	copied.Note = NoteDuplicate
	copyFile(t, original.Path, copied.Path)
	store.Files = append(store.Files, copied)
	store.DuplicateID = original.SessionID

	irregular := func(relative string, created time.Time, id string, note string, compressed bool) {
		file := File{Path: filepath.Join(home, filepath.FromSlash(relative)), SessionID: id, Note: note, Compressed: compressed}
		file.Archived = strings.HasPrefix(relative, "archived_sessions/")
		store.Files = append(store.Files, writeUnrelatedRollout(t, spec, file, created, 200*1024, !spec.SparseUnrelated || compressed))
	}
	when := func(month time.Month, day int) time.Time {
		return time.Date(2026, month, day, 9, 15, 0, 0, spec.Location)
	}
	irregular("sessions/2026/05/10/rollout-2026-05-10T09-15-00.jsonl", when(5, 10), UUID(when(5, 10), 7, rnd), NoteNoUUID, false)
	irregular("sessions/2026/05/11/custom-export.jsonl", when(5, 11), UUID(when(5, 11), 7, rnd), NoteCustomName, false)
	irregular("sessions/notes/rollout-garbage.jsonl", when(5, 12), UUID(when(5, 12), 7, rnd), NoteNonDateDir, false)
	nonV7 := UUID(when(6, 1), 4, rnd)
	irregular(RolloutRelative(when(6, 1), nonV7, false), when(6, 1), nonV7, NoteNonV7, false)
	mismatch := UUID(when(6, 2), 7, rnd)
	mismatchFile := File{Path: ConventionalPath(home, when(6, 2), UUID(when(6, 2), 7, rnd), false), SessionID: mismatch, Note: NoteMismatch}
	store.Files = append(store.Files, writeUnrelatedRollout(t, spec, mismatchFile, when(6, 2), 200*1024, !spec.SparseUnrelated))
	compressedID := UUID(when(7, 1), 7, rnd)
	irregular(RolloutRelative(when(7, 1), compressedID, false)+".gz", when(7, 1), compressedID, NoteGzip, true)

	rootCreated := spec.RootCreated.Truncate(time.Millisecond)
	store.RootID = UUID(rootCreated, 7, rnd)
	type child struct {
		id, task string
		created  time.Time
	}
	children := make([]child, len(spec.ChildDelays))
	for i, delay := range spec.ChildDelays {
		created := rootCreated.Add(delay)
		children[i] = child{id: UUID(created, 7, rnd), task: fmt.Sprintf("task-%02d", i+1), created: created}
	}
	var spawns []spawn
	for _, c := range children {
		spawns = append(spawns, spawn{at: c.created.Add(-200 * time.Millisecond), task: c.task})
	}
	root := File{Path: ConventionalPath(home, rootCreated, store.RootID, false), SessionID: store.RootID, Family: true, Conventional: true, Note: "root"}
	store.Files = append(store.Files, writeFamilyRollout(t, spec, root, rootCreated, `"cli"`, spec.RootBytes, spawns))
	store.RootPath = root.Path
	for _, c := range children {
		source := fmt.Sprintf(`{"subagent":{"thread_spawn":{"parent_thread_id":%q,"depth":1,"agent_path":%q,"agent_role":"worker"}}}`, store.RootID, "/root/"+c.task)
		file := File{Path: ConventionalPath(home, c.created, c.id, false), SessionID: c.id, ParentID: store.RootID, Family: true, Conventional: true, Note: "child " + c.task}
		store.Files = append(store.Files, writeFamilyRollout(t, spec, file, c.created, source, spec.ChildBytes, nil))
		store.ChildIDs = append(store.ChildIDs, c.id)
		store.ChildPaths = append(store.ChildPaths, file.Path)
	}
	// Optional later and in-window rollouts draw from the generator only after
	// the family, so they leave every other identity unchanged.
	for i := 0; i < spec.LaterUnrelated; i++ {
		created := rootCreated.Add(spec.LaterSpan * time.Duration(i+1) / time.Duration(spec.LaterUnrelated+1)).Truncate(time.Millisecond)
		id := UUID(created, 7, rnd)
		size := spec.MinUnrelated + rnd.Int64N(spec.MaxUnrelated-spec.MinUnrelated+1)
		file := File{Path: ConventionalPath(home, created, id, false), SessionID: id, Conventional: true, Later: true}
		store.Files = append(store.Files, writeUnrelatedRollout(t, spec, file, created, size, !spec.SparseUnrelated))
	}
	if spec.WindowIrregular {
		local := rootCreated.In(spec.Location)
		noUUID := local.Add(3 * time.Minute)
		irregular(noUUID.Format("sessions/2006/01/02/rollout-2006-01-02T15-04-05")+".jsonl", noUUID, UUID(noUUID, 7, rnd), NoteWindowNoUUID, false)
		nonV7At := local.Add(4 * time.Minute)
		windowNonV7 := UUID(nonV7At, 4, rnd)
		irregular(RolloutRelative(nonV7At, windowNonV7, false), nonV7At, windowNonV7, NoteWindowNonV7, false)
	}
	sort.Slice(store.Files, func(i, j int) bool { return store.Files[i].Path < store.Files[j].Path })
	return store
}

// TotalBytes sums the sizes of the matching files, or of all files.
func (s LargeStore) TotalBytes(match func(File) bool) int64 {
	var total int64
	for _, file := range s.Files {
		if match == nil || match(file) {
			total += file.Size
		}
	}
	return total
}

// Count counts the matching files.
func (s LargeStore) Count(match func(File) bool) int {
	total := 0
	for _, file := range s.Files {
		if match(file) {
			total++
		}
	}
	return total
}

// File returns the generated file at path.
func (s LargeStore) File(path string) File {
	for _, file := range s.Files {
		if file.Path == path {
			return file
		}
	}
	panic("fixture file not found: " + path)
}

// Noted returns the generated file with the given irregular-name note.
func (s LargeStore) Noted(note string) File {
	for _, file := range s.Files {
		if file.Note == note {
			return file
		}
	}
	panic("fixture note not found: " + note)
}

// RolloutRelative is the store-relative path Codex uses for a rollout.
func RolloutRelative(created time.Time, id string, archived bool) string {
	name := "rollout-" + created.Format("2006-01-02T15-04-05") + "-" + id + ".jsonl"
	if archived {
		return "archived_sessions/" + name
	}
	return created.Format("sessions/2006/01/02/") + name
}

// ConventionalPath joins RolloutRelative to home.
func ConventionalPath(home string, created time.Time, id string, archived bool) string {
	return filepath.Join(home, filepath.FromSlash(RolloutRelative(created, id, archived)))
}

// UUID returns a version 7 UUID whose first 48 bits are the UTC creation
// milliseconds, or a version 4 UUID that carries no instant.
func UUID(created time.Time, version byte, rnd *rand.Rand) string {
	var b [16]byte
	for i := range b {
		b[i] = byte(rnd.Uint32())
	}
	if version == 7 {
		ms := uint64(created.UnixMilli())
		for i := 0; i < 6; i++ {
			b[i] = byte(ms >> (40 - 8*i))
		}
	}
	b[6] = version<<4 | b[6]&0x0f
	b[8] = 0x80 | b[8]&0x3f
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[0:4], b[4:6], b[6:8], b[8:10], b[10:16])
}

// UUIDv7Millis decodes the creation milliseconds of a version 7 UUID.
func UUIDv7Millis(t testing.TB, id string) int64 {
	t.Helper()
	compact := strings.ReplaceAll(id, "-", "")
	if len(compact) != 32 || compact[12] != '7' {
		t.Fatalf("%s is not a version 7 UUID", id)
	}
	var millis int64
	if _, err := fmt.Sscanf(compact[:12], "%x", &millis); err != nil {
		t.Fatal(err)
	}
	return millis
}

// Timestamp formats a Codex record timestamp.
func Timestamp(at time.Time) string {
	return at.UTC().Format("2006-01-02T15:04:05.000Z")
}

// sessionMetaLine pads base_instructions so the line, including its newline,
// is exactly the configured header size.
func sessionMetaLine(t testing.TB, id string, created time.Time, cwd, source string, size int) string {
	t.Helper()
	build := func(padding string) string {
		text, err := json.Marshal(padding)
		if err != nil {
			t.Fatal(err)
		}
		return fmt.Sprintf(`{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":%q,"originator":"codex_cli_rs","cli_version":"0.63.0","source":%s,"model_provider":"openai","base_instructions":{"text":%s}}}`+"\n",
			Timestamp(created), id, Timestamp(created), cwd, source, text)
	}
	missing := size - len(build(""))
	if missing < 0 {
		t.Fatalf("header size %d is smaller than the bare session_meta line", size)
	}
	line := build(fill(paddingText, missing))
	if len(line) != size {
		t.Fatalf("header size %d, want %d", len(line), size)
	}
	return line
}

func fill(text string, n int) string {
	return strings.Repeat(text, n/len(text)+1)[:n]
}

func turnLines(id string, turn int, at time.Time, cwd string, responseBytes int) string {
	return fmt.Sprintf(`{"timestamp":%q,"type":"turn_context","payload":{"turn_id":"turn-%d","cwd":%q,"model":"gpt-fixture"}}`+"\n"+
		`{"timestamp":%q,"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":%q}]}}`+"\n"+
		`{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-%d","usage":{"input_tokens":%d,"output_tokens":%d}}}`+"\n",
		Timestamp(at), turn, cwd,
		Timestamp(at.Add(time.Second)), fill(responseText, responseBytes),
		Timestamp(at.Add(2*time.Second)), id, turn, 1000+turn, 100+turn)
}

func writeUnrelatedRollout(t testing.TB, spec LargeStoreSpec, file File, created time.Time, size int64, dense bool) File {
	t.Helper()
	cwd := fmt.Sprintf("/fixture/unrelated/project-%d", created.YearDay()%37)
	header := sessionMetaLine(t, file.SessionID, created, cwd, `"cli"`, spec.HeaderBytes)
	file.HeaderBytes = int64(len(header))
	var body strings.Builder
	body.WriteString(header)
	at := created.Add(time.Second)
	for turn := 1; ; turn++ {
		lines := turnLines(file.SessionID, turn, at, cwd, 256)
		if dense && int64(body.Len()+len(lines)) > size && turn > 1 {
			break
		}
		body.WriteString(lines)
		at = at.Add(10 * time.Second)
		if !dense {
			break
		}
	}
	if err := os.MkdirAll(filepath.Dir(file.Path), 0o700); err != nil {
		t.Fatal(err)
	}
	if file.Compressed {
		f, err := os.Create(file.Path)
		if err != nil {
			t.Fatal(err)
		}
		writer := gzip.NewWriter(f)
		if _, err := writer.Write([]byte(body.String())); err != nil {
			t.Fatal(err)
		}
		if err := writer.Close(); err != nil {
			t.Fatal(err)
		}
		if err := f.Close(); err != nil {
			t.Fatal(err)
		}
	} else {
		if err := os.WriteFile(file.Path, []byte(body.String()), 0o600); err != nil {
			t.Fatal(err)
		}
		if !dense && size > int64(body.Len()) {
			if err := os.Truncate(file.Path, size); err != nil {
				t.Fatal(err)
			}
		}
	}
	info, err := os.Stat(file.Path)
	if err != nil {
		t.Fatal(err)
	}
	file.Size = info.Size()
	return file
}

type spawn struct {
	at   time.Time
	task string
}

func writeFamilyRollout(t testing.TB, spec LargeStoreSpec, file File, created time.Time, source string, size int64, spawns []spawn) File {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file.Path), 0o700); err != nil {
		t.Fatal(err)
	}
	f, err := os.Create(file.Path)
	if err != nil {
		t.Fatal(err)
	}
	writer := bufio.NewWriterSize(f, 1<<20)
	header := sessionMetaLine(t, file.SessionID, created, spec.Project, source, spec.HeaderBytes)
	file.HeaderBytes = int64(len(header))
	written := int64(0)
	write := func(text string) {
		if _, err := writer.WriteString(text); err != nil {
			t.Fatal(err)
		}
		written += int64(len(text))
	}
	write(header)
	const responseBytes = 4000
	turnBytes := int64(len(turnLines(file.SessionID, 1, created, spec.Project, responseBytes)))
	turns := int((size - written) / turnBytes)
	if turns < len(spawns)+1 {
		turns = len(spawns) + 1
	}
	// Spread turns evenly over the hour after creation so every spawn falls
	// between two turns at its own timestamp.
	step := time.Hour / time.Duration(turns)
	next := 0
	for turn := 1; turn <= turns; turn++ {
		at := created.Add(time.Duration(turn) * step)
		for next < len(spawns) && !spawns[next].at.After(at) {
			call := fmt.Sprintf("call-spawn-%02d", next+1)
			arguments, err := json.Marshal(fmt.Sprintf(`{"agent_type":"worker","task_name":%q,"message":"synthetic delegated task"}`, spawns[next].task))
			if err != nil {
				t.Fatal(err)
			}
			output, err := json.Marshal(fmt.Sprintf(`{"task_name":%q}`, "/root/"+spawns[next].task))
			if err != nil {
				t.Fatal(err)
			}
			write(fmt.Sprintf(`{"timestamp":%q,"type":"response_item","payload":{"type":"function_call","name":"spawn_agent","call_id":%q,"arguments":%s}}`+"\n",
				Timestamp(spawns[next].at), call, arguments))
			write(fmt.Sprintf(`{"timestamp":%q,"type":"response_item","payload":{"type":"function_call_output","call_id":%q,"output":%s}}`+"\n",
				Timestamp(spawns[next].at.Add(100*time.Millisecond)), call, output))
			next++
		}
		write(turnLines(file.SessionID, turn, at, spec.Project, responseBytes))
	}
	if next != len(spawns) {
		t.Fatalf("fixture placed %d of %d spawns", next, len(spawns))
	}
	if err := writer.Flush(); err != nil {
		t.Fatal(err)
	}
	if err := f.Close(); err != nil {
		t.Fatal(err)
	}
	file.Size = written
	return file
}

func copyFile(t testing.TB, from, to string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(to), 0o700); err != nil {
		t.Fatal(err)
	}
	source, err := os.Open(from)
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	target, err := os.OpenFile(to, os.O_WRONLY|os.O_CREATE|os.O_EXCL, 0o600)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := io.Copy(target, source); err != nil {
		t.Fatal(err)
	}
	if err := target.Close(); err != nil {
		t.Fatal(err)
	}
}
