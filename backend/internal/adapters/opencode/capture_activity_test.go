package opencode

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
)

func TestCaptureChildFinishedRequiresNativeParentAndCompletedTask(t *testing.T) {
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "source.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := db.Close(); err != nil {
			t.Error(err)
		}
	}()
	_, err = db.Exec(`CREATE TABLE session(id TEXT,parent_id TEXT); CREATE TABLE part(session_id TEXT,time_updated INTEGER,data TEXT); INSERT INTO session VALUES('child','parent'); INSERT INTO part VALUES('parent',1,'{"type":"tool","tool":"task","state":{"status":"running","metadata":{"sessionId":"child"}}}')`)
	if err != nil {
		t.Fatal(err)
	}
	if finished, err := CaptureChildFinished(context.Background(), db, "parent", "child"); err != nil || finished {
		t.Fatalf("running %v %v", finished, err)
	}
	if _, err = db.Exec(`UPDATE part SET data=replace(data,'running','completed')`); err != nil {
		t.Fatal(err)
	}
	if finished, err := CaptureChildFinished(context.Background(), db, "parent", "child"); err != nil || !finished {
		t.Fatalf("completed %v %v", finished, err)
	}
	if finished, err := CaptureChildFinished(context.Background(), db, "wrong", "child"); err != nil || finished {
		t.Fatalf("wrong parent %v %v", finished, err)
	}
	if _, err = db.Exec(`INSERT INTO part SELECT session_id,2,replace(data,'completed','running') FROM part`); err != nil {
		t.Fatal(err)
	}
	if finished, err := CaptureChildFinished(context.Background(), db, "parent", "child"); err != nil || finished {
		t.Fatalf("resumed task %v %v", finished, err)
	}

}
