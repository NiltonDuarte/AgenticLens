package capture

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
)

// The environment at package initialization predates any test's Setenv, so
// the fixture build keeps the developer's Go cache even when the first caller
// has already replaced HOME or XDG directories with disposable ones.
var nativeFixtureBuildEnv = os.Environ()

var (
	nativeFixtureOnce sync.Once
	nativeFixtureRoot string
	nativeFixturePath string
	nativeFixtureErr  error
)

// nativeTestFixture returns a read-only host-native executable that answers
// the capture interface probes (see testdata/nativefixture). It is built once
// per test process and without the race detector, because it contains no
// production code; production validation still runs in full against it.
func nativeTestFixture(t testing.TB) string {
	t.Helper()
	nativeFixtureOnce.Do(func() {
		nativeFixtureRoot, nativeFixtureErr = os.MkdirTemp("", "agenticlens-native-fixture-")
		if nativeFixtureErr != nil {
			nativeFixtureErr = fmt.Errorf("create native fixture directory: %w", nativeFixtureErr)
			return
		}
		name := "agentic-lens-native-fixture"
		if runtime.GOOS == "windows" {
			name += ".exe"
		}
		nativeFixturePath = filepath.Join(nativeFixtureRoot, name)
		ldflags := fmt.Sprintf("-X main.captureProtocol=%d", HandlerProtocolVersion)
		command := exec.Command("go", "build", "-ldflags", ldflags, "-o", nativeFixturePath, "./internal/adapters/capture/testdata/nativefixture")
		command.Dir = filepath.Join("..", "..", "..")
		command.Env = nativeFixtureBuildEnv
		if output, err := command.CombinedOutput(); err != nil {
			nativeFixtureErr = fmt.Errorf("build native fixture: %w: %s", err, output)
			return
		}
		if err := os.Chmod(nativeFixturePath, 0o555); err != nil {
			nativeFixtureErr = fmt.Errorf("make native fixture read-only: %w", err)
		}
	})
	if nativeFixtureErr != nil {
		t.Fatal(nativeFixtureErr)
	}
	return nativeFixturePath
}

// NativeTestFixture exposes the shared native fixture to external tests.
func NativeTestFixture(t testing.TB) string {
	t.Helper()
	return nativeTestFixture(t)
}

// RemoveNativeTestFixture deletes the built fixture; TestMain calls it after
// the run. It reports a failure on stderr because tests have already finished.
func RemoveNativeTestFixture() {
	if nativeFixtureRoot == "" {
		return
	}
	if err := os.RemoveAll(nativeFixtureRoot); err != nil {
		fmt.Fprintf(os.Stderr, "remove native fixture: %v\n", err)
	}
}
