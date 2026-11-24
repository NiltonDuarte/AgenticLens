package codexstoretest

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// The generator's own invariants. Consumers rely on them for their byte and
// window oracles, so a generator regression must fail here first.
func TestWriteLargeStoreInvariants(t *testing.T) {
	spec := IncidentScaleSpec("/fixture/project", 16)
	spec.LaterUnrelated, spec.LaterSpan, spec.WindowIrregular = 4, 90*time.Minute, true
	store := WriteLargeStore(t, spec)
	rootMillis := UUIDv7Millis(t, store.RootID)
	if rootMillis != spec.RootCreated.UnixMilli() {
		t.Fatalf("root UUID encodes %d, want its creation %d", rootMillis, spec.RootCreated.UnixMilli())
	}
	notes := map[string]int{}
	for _, file := range store.Files {
		notes[file.Note]++
		info, err := os.Stat(file.Path)
		if err != nil || info.Size() != file.Size {
			t.Fatalf("%s: recorded size %d, stat %v %v", file.Path, file.Size, info, err)
		}
		if file.HeaderBytes != HeaderBytes {
			t.Fatalf("%s header is %d bytes, want %d", file.Path, file.HeaderBytes, HeaderBytes)
		}
		if !file.Conventional {
			continue
		}
		// A conventional name carries the session's own UUIDv7, whose instant
		// is on the correct side of the root.
		if !strings.Contains(filepath.Base(file.Path), file.SessionID) {
			t.Fatalf("%s does not name its session %s", file.Path, file.SessionID)
		}
		millis := UUIDv7Millis(t, file.SessionID)
		child := file.Family && file.SessionID != store.RootID
		if (child && millis <= rootMillis) || (!file.Family && !file.Later && millis >= rootMillis) {
			t.Fatalf("%s instant %d is on the wrong side of the root instant %d", file.Path, millis, rootMillis)
		}
		if file.Later && (millis <= rootMillis || millis > rootMillis+spec.LaterSpan.Milliseconds()) {
			t.Fatalf("%s instant %d is on the wrong side of the root instant %d", file.Path, millis, rootMillis)
		}
		if !file.Archived {
			local := time.UnixMilli(millis).In(spec.Location).Format("2006/01/02")
			if !strings.Contains(filepath.ToSlash(file.Path), "/sessions/"+local+"/") {
				t.Fatalf("%s is not in its machine-local day directory %s", file.Path, local)
			}
		}
	}
	for _, note := range []string{NoteNoUUID, NoteCustomName, NoteNonDateDir, NoteNonV7, NoteMismatch, NoteGzip, NoteDuplicate, NoteWindowNoUUID, NoteWindowNonV7} {
		if notes[note] != 1 {
			t.Fatalf("irregular name %q generated %d times, want once", note, notes[note])
		}
	}
	rootDay := filepath.Dir(store.RootPath)
	for _, note := range []string{NoteWindowNoUUID, NoteWindowNonV7} {
		if filepath.Dir(store.Noted(note).Path) != rootDay {
			t.Fatalf("%s is not in the root's day directory", note)
		}
	}
	if later := store.Count(func(f File) bool { return f.Later }); later != spec.LaterUnrelated {
		t.Fatalf("generated %d later rollouts, want %d", later, spec.LaterUnrelated)
	}
	if len(store.ChildIDs) != len(spec.ChildDelays) || notes["root"] != 1 {
		t.Fatalf("family has %d children and %d roots", len(store.ChildIDs), notes["root"])
	}
}

func TestTaskFamilyNamesEncodeCreation(t *testing.T) {
	created := time.Date(2026, 11, 24, 9, 0, 0, 0, time.UTC)
	family := NewTaskFamily(t, t.TempDir(), "/fixture/project", created, 3, 2, 4096)
	if got := UUIDv7Millis(t, family.ParentID); got != created.UnixMilli() {
		t.Fatalf("parent UUID encodes %d, want %d", got, created.UnixMilli())
	}
	for i, id := range family.ChildIDs {
		if got := UUIDv7Millis(t, id); got != family.ChildTimes[i].UnixMilli() || !strings.Contains(family.Path(id), id) {
			t.Fatalf("child %d UUID or name does not encode its creation: %s %s", i, id, family.Path(id))
		}
		info, err := os.Stat(family.Path(id))
		if err != nil {
			t.Fatal(err)
		}
		if large := info.Size() >= 4096; large != (i >= 2) {
			t.Fatalf("child %d size %d does not follow LargeFrom", i, info.Size())
		}
	}
}
