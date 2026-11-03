package capture

// Registry protocol 2 stores logical projects once and explicitly scopes every
// operational row by its exact worktree. Diagnostic sequences are global IDs;
// every lookup and retention operation also requires worktree_id.
const registrySchema = `CREATE TABLE IF NOT EXISTS capture_meta(version INTEGER NOT NULL);
INSERT INTO capture_meta SELECT 2 WHERE NOT EXISTS(SELECT 1 FROM capture_meta);
CREATE TABLE IF NOT EXISTS project_identity(singleton INTEGER PRIMARY KEY CHECK(singleton=1),name TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS worktrees(id INTEGER PRIMARY KEY,directory TEXT NOT NULL UNIQUE);
CREATE TABLE IF NOT EXISTS worktree_ownership(worktree_id INTEGER PRIMARY KEY,ownership TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS worktree_relocations(worktree_id INTEGER PRIMARY KEY,migration_id TEXT NOT NULL,state TEXT NOT NULL);
CREATE TABLE IF NOT EXISTS registrations(worktree_id INTEGER NOT NULL,identity TEXT NOT NULL,source TEXT NOT NULL,store TEXT NOT NULL,session_id TEXT NOT NULL,ticket TEXT NOT NULL,record BLOB NOT NULL,PRIMARY KEY(worktree_id,identity));
CREATE INDEX IF NOT EXISTS registrations_history_owner ON registrations(session_id,source,ticket,worktree_id);
CREATE TABLE IF NOT EXISTS capture_auxiliary_receipts(worktree_id INTEGER NOT NULL,ticket TEXT NOT NULL,record BLOB NOT NULL,PRIMARY KEY(worktree_id,ticket));
CREATE TABLE IF NOT EXISTS pending_starts(worktree_id INTEGER NOT NULL,store TEXT,parent TEXT,agent TEXT,PRIMARY KEY(worktree_id,store,parent,agent));
CREATE TABLE IF NOT EXISTS pending_children(worktree_id INTEGER NOT NULL,store TEXT,path TEXT,parent TEXT,PRIMARY KEY(worktree_id,store,path));
CREATE TABLE IF NOT EXISTS discovery_cursor(worktree_id INTEGER NOT NULL,parent TEXT,path TEXT NOT NULL,PRIMARY KEY(worktree_id,parent));
CREATE TABLE IF NOT EXISTS child_cursor(worktree_id INTEGER NOT NULL,kind TEXT,parent TEXT,key TEXT,more INTEGER NOT NULL DEFAULT 0,PRIMARY KEY(worktree_id,kind,parent));
CREATE TABLE IF NOT EXISTS project_proof_cursor(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),identity TEXT NOT NULL,PRIMARY KEY(worktree_id,singleton));
CREATE TABLE IF NOT EXISTS publication_request(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),generation INTEGER NOT NULL,acknowledged INTEGER NOT NULL,requested_at INTEGER NOT NULL,PRIMARY KEY(worktree_id,singleton));
CREATE TABLE IF NOT EXISTS publication_cursor(worktree_id INTEGER NOT NULL,singleton INTEGER NOT NULL CHECK(singleton=1),ticket TEXT NOT NULL,PRIMARY KEY(worktree_id,singleton));
CREATE TABLE IF NOT EXISTS recovery_phase(worktree_id INTEGER NOT NULL,ticket TEXT,phase INTEGER NOT NULL,PRIMARY KEY(worktree_id,ticket));
CREATE TABLE IF NOT EXISTS reconcile_cursor(worktree_id INTEGER NOT NULL,ticket TEXT,identity TEXT NOT NULL,PRIMARY KEY(worktree_id,ticket));
` + publicationAuxiliarySchema + diagnosticSchema
