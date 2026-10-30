package opencode_test

import (
	"database/sql"
	"testing"

	"agenticlens/internal/adapters/opencode"
	_ "modernc.org/sqlite"
)

func TestSchemaV2SessionRootResolutionIsExplicit(t *testing.T) {
	db, err := sql.Open("sqlite", ":memory:")
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()
	createSchemaV2TestDB(t, db)
	if _, err := db.Exec(`INSERT INTO session_v2 VALUES('empty',NULL,'build','mock-model','AL-58 empty',1793231971000,1793231971000)`); err != nil {
		t.Fatal(err)
	}
	root, err := opencode.ResolveActiveRootSession(db, "", "", "empty")
	if err != nil || root == nil || *root != "empty" {
		t.Fatalf("explicit v2 root=%v err=%v", root, err)
	}
}
