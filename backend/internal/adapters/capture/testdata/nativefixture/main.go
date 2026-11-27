// Command nativefixture is a small host-native executable that answers the
// interface probes capture installation runs against an engine or manager. The
// capture tests build it once per test process, so production validation of
// format, architecture, checksum, journal and retention runs against a real
// native file without copying the much larger test executable.
//
// AL79_CAPTURE_NATIVE_FIXTURE selects a variant: "old-interface" reports the
// previous capture protocol and "bad-manager" reports an incompatible manager
// interface. Any other value, including an empty one, gives compatible answers.
package main

import (
	"fmt"
	"os"
	"strings"
)

// captureProtocol is set at build time from capture.HandlerProtocolVersion.
var captureProtocol string

func main() {
	if captureProtocol == "" {
		fmt.Fprintln(os.Stderr, "native fixture built without a capture protocol")
		os.Exit(2)
	}
	answer, ok := respond(strings.Join(os.Args[1:], " "), os.Getenv("AL79_CAPTURE_NATIVE_FIXTURE"))
	if !ok {
		fmt.Fprintf(os.Stderr, "native fixture: unsupported arguments %q\n", os.Args[1:])
		os.Exit(2)
	}
	fmt.Println(answer)
}

func respond(command, variant string) (string, bool) {
	switch command {
	case "engine protocol":
		return "1", true
	case "capture protocol":
		if variant == "old-interface" {
			return "1", true
		}
		return captureProtocol, true
	case "capture protected-worker-protocol":
		return "2", true
	case "manager protocol", "dispatch protocol", "manager codex-binding-wake-protocol", "manager protected-codex-hook-protocol":
		if variant == "bad-manager" {
			return "0", true
		}
		return "1", true
	case "version --short":
		return "v9.8.7", true
	case "version --json", "--manager-version --json":
		return `{"version":"v9.8.7","commit":"native-test-fixture","build_date":"test"}`, true
	default:
		return "", false
	}
}
