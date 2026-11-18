package main_test

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"testing"
	"time"

	adaptershttp "agenticlens/internal/adapters/http"
)

func TestServerReportsUnreadableRegularSQLiteFile(t *testing.T) {
	// The server must traverse the fixture's ancestors after a root runner
	// drops privileges. The fixture contains only a binary and synthetic data.
	project, err := os.MkdirTemp("/tmp", "agenticlens-live-permission-*")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := os.RemoveAll(project); err != nil {
			t.Error(err)
		}
	})
	writeCLISharedPolicy(t, project)
	if err := os.Mkdir(filepath.Join(project, "records"), 0755); err != nil {
		t.Fatal(err)
	}
	source := filepath.Join(project, "denied.db")
	if err := os.WriteFile(source, readCaptureFile(t, setupIsolatedEmptyDB(t)), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(source, 0000); err != nil {
		t.Fatal(err)
	}
	var address, logPath string
	if os.Geteuid() == 0 {
		address, logPath = startUnprivilegedDashboard(t, project, source)
	} else {
		address, logPath = startFixtureDashboard(t, project, "--db-path", source)
	}

	startup := string(readCaptureFile(t, logPath))
	if strings.Contains(startup, "Live telemetry active") {
		t.Fatalf("unreadable source advertised active: %s", startup)
	}
	client := &http.Client{Timeout: 3 * time.Second}
	for _, route := range []string{"sessions", "sessions/owner/trace", "stream"} {
		resp, err := client.Get(address + "/api/live/" + route)
		if err != nil {
			t.Fatal(err)
		}
		var body adaptershttp.ErrorResponse
		err = json.NewDecoder(resp.Body).Decode(&body)
		resp.Body.Close()
		if err != nil || resp.StatusCode != http.StatusServiceUnavailable || body.StatusCode != resp.StatusCode || body.Error != "source_unavailable" {
			t.Fatalf("%s: HTTP %d, %+v (%v)", route, resp.StatusCode, body, err)
		}
		if !strings.Contains(body.Message, source) || !(strings.Contains(body.Message, "unable to open database file") || strings.Contains(body.Message, "permission denied")) || !strings.Contains(startup, body.Message) {
			t.Fatalf("%s: unavailable source reason differs from the denied open: %+v, startup=%s", route, body, startup)
		}
	}
}

// Run the same compiled engine without root's DAC override. A restricted
// root container must permit setuid/setgid to execute this Linux test.
func startUnprivilegedDashboard(t *testing.T, project, source string) (string, string) {
	t.Helper()
	exe := filepath.Join(project, "engine")
	if err := os.WriteFile(exe, readCaptureFile(t, cliBinaryPath), 0555); err != nil {
		t.Fatal(err)
	}
	// The test runner may use umask 077. Make only the non-source fixture
	// files traversable/readable for the server's unprivileged credentials.
	if err := os.Chmod(exe, 0555); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(project, "records"), 0555); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(filepath.Join(project, ".agenticlens-config.yaml"), 0444); err != nil {
		t.Fatal(err)
	}
	if err := os.Chmod(project, 0755); err != nil {
		t.Fatal(err)
	}
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatal(err)
	}
	port := listener.Addr().(*net.TCPAddr).Port
	if err := listener.Close(); err != nil {
		t.Fatal(err)
	}
	log, err := os.CreateTemp(t.TempDir(), "server-log-*")
	if err != nil {
		t.Fatal(err)
	}
	cmd := exec.Command(exe, "server", "--port", fmt.Sprint(port), "--db-path", source)
	cmd.Dir, cmd.Stdout, cmd.Stderr = project, log, log
	cmd.SysProcAttr = &syscall.SysProcAttr{Credential: &syscall.Credential{Uid: 65534, Gid: 65534}}
	if err := cmd.Start(); err != nil {
		log.Close()
		t.Fatalf("start engine with unprivileged credentials: %v", err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait(); _ = log.Close() })
	address := fmt.Sprintf("http://127.0.0.1:%d", port)
	resp := waitForServerReady(t, address+"/api/workspaces", 5*time.Second)
	resp.Body.Close()
	return address, log.Name()
}
