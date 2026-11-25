package codexstoretest

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
)

// TaskFamily is a small Codex parent whose spawn_agent results carry only a
// task name, so a delegation resolves to its child only when the child's
// rollout is read with the parent. Names follow the rollout convention with
// UUIDv7 identities that encode each rollout's UTC creation instant.
type TaskFamily struct {
	Home       string
	WorkingDir string
	Created    time.Time
	ParentID   string
	ChildIDs   []string
	ChildTimes []time.Time
	// LargeFrom is the first child index written with a body of LargeBytes.
	LargeFrom  int
	LargeBytes int
	paths      map[string]string
}

// TaskFamilyUUID returns a version 7 UUID for created with a fixed serial.
func TaskFamilyUUID(created time.Time, serial int) string {
	millis := created.UnixMilli()
	return fmt.Sprintf("%08x-%04x-7000-8000-%012x", millis>>16, millis&0xffff, serial)
}

// NewTaskFamily writes a parent with children task-name delegations and one
// rollout per child under home. Children from largeFrom on carry a real,
// parseable body of at least largeBytes.
func NewTaskFamily(t testing.TB, home, workingDir string, created time.Time, children, largeFrom, largeBytes int) *TaskFamily {
	t.Helper()
	family := &TaskFamily{Home: home, WorkingDir: workingDir, Created: created.UTC(), ParentID: TaskFamilyUUID(created, 1), LargeFrom: largeFrom, LargeBytes: largeBytes, paths: map[string]string{}}
	for i := 0; i < children; i++ {
		at := family.Created.Add(time.Duration(7+30*i) * time.Second)
		family.ChildTimes = append(family.ChildTimes, at)
		family.ChildIDs = append(family.ChildIDs, TaskFamilyUUID(at, 10+i))
	}
	family.WriteParent(t, 11, nil)
	for i := range family.ChildIDs {
		family.WriteChild(t, i, family.ParentID, 20+i)
	}
	return family
}

// Path returns the rollout path of the parent or a child.
func (f *TaskFamily) Path(id string) string { return f.paths[id] }

func (f *TaskFamily) path(t testing.TB, id string, created time.Time) string {
	t.Helper()
	if path, ok := f.paths[id]; ok {
		return path
	}
	path := ConventionalPath(f.Home, created, id, false)
	if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
		t.Fatal(err)
	}
	f.paths[id] = path
	return path
}

// TaskName is the task name of child i.
func TaskName(i int) string { return fmt.Sprintf("task-%02d", i+1) }

// SpawnCallID is the call ID of child i's spawn_agent call.
func SpawnCallID(i int) string { return fmt.Sprintf("call-spawn-%02d", i+1) }

// WriteParent rewrites the parent with the given input tokens. outputs
// optionally replaces the function_call_output of child i's spawn.
func (f *TaskFamily) WriteParent(t testing.TB, tokens int, outputs map[int]string) {
	t.Helper()
	var body strings.Builder
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":%q,"source":"cli"}}`+"\n", Timestamp(f.Created), f.ParentID, Timestamp(f.Created), f.WorkingDir)
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}`+"\n", Timestamp(f.Created.Add(time.Second)))
	for i, created := range f.ChildTimes {
		at := created.Add(-200 * time.Millisecond)
		arguments, err := json.Marshal(fmt.Sprintf(`{"agent_type":"worker","task_name":%q,"message":"synthetic delegated task"}`, TaskName(i)))
		if err != nil {
			t.Fatal(err)
		}
		output := fmt.Sprintf(`{"task_name":%q}`, "/root/"+TaskName(i))
		if replacement, ok := outputs[i]; ok {
			output = replacement
		}
		encoded, err := json.Marshal(output)
		if err != nil {
			t.Fatal(err)
		}
		fmt.Fprintf(&body, `{"timestamp":%q,"type":"response_item","payload":{"type":"function_call","name":"spawn_agent","call_id":%q,"arguments":%s}}`+"\n", Timestamp(at), SpawnCallID(i), arguments)
		fmt.Fprintf(&body, `{"timestamp":%q,"type":"response_item","payload":{"type":"function_call_output","call_id":%q,"output":%s}}`+"\n", Timestamp(at.Add(100*time.Millisecond)), SpawnCallID(i), encoded)
	}
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}`+"\n", Timestamp(f.Created.Add(3*time.Minute)), f.ParentID, tokens)
	if err := os.WriteFile(f.path(t, f.ParentID, f.Created), []byte(body.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}

// WriteChild rewrites child i, naming parent only in
// source.subagent.thread_spawn with the agent_path its task name references.
func (f *TaskFamily) WriteChild(t testing.TB, i int, parent string, tokens int) {
	t.Helper()
	id, created := f.ChildIDs[i], f.ChildTimes[i]
	var body strings.Builder
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"session_meta","payload":{"id":%q,"timestamp":%q,"cwd":%q,"source":{"subagent":{"thread_spawn":{"parent_thread_id":%q,"depth":1,"agent_path":%q,"agent_role":"worker"}}}}}`+"\n",
		Timestamp(created), id, Timestamp(created), f.WorkingDir, parent, "/root/"+TaskName(i))
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"turn_context","payload":{"model":"fixture","turn_id":"turn-one"}}`+"\n", Timestamp(created.Add(time.Second)))
	if i >= f.LargeFrom && f.LargeBytes > 0 {
		text := strings.Repeat("synthetic child output ", f.LargeBytes/23+1)
		fmt.Fprintf(&body, `{"timestamp":%q,"type":"response_item","payload":{"type":"message","role":"assistant","content":[{"type":"output_text","text":%q}]}}`+"\n", Timestamp(created.Add(2*time.Second)), text)
	}
	fmt.Fprintf(&body, `{"timestamp":%q,"type":"token_usage_record","payload":{"thread_id":%q,"turn_id":"turn-one","usage":{"input_tokens":%d,"output_tokens":2}}}`+"\n", Timestamp(created.Add(3*time.Second)), id, tokens)
	if err := os.WriteFile(f.path(t, id, created), []byte(body.String()), 0o600); err != nil {
		t.Fatal(err)
	}
}
