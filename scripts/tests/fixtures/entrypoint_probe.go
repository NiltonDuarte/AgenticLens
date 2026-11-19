// Native distribution fixture. Compilers belong to test tooling, not installed runtime.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"syscall"
	"time"
)

var version = "v9.8.7"
var protocol = "3"

// Older managers supported the generic manager interface but had no OpenCode
// router. Forward their other operations to the real test manager.
func simulatePreviousManager(path string, args []string) {
	if len(args) == 2 && args[0] == "manager" && args[1] == "opencode-hook-protocol" {
		fmt.Fprintln(os.Stderr, "unknown command")
		os.Exit(1)
	}
	runPreviousBinary(path, args)
}

// Previous install-user accepted --source but always installed Codex hooks.
// Its flag parser rejected the newer --user-source before calling the installer.
func simulatePreviousEngine(path string, args []string) {
	if len(args) >= 2 && args[0] == "capture" && (args[1] == "install-user" || args[1] == "uninstall-user") {
		legacyArgs := append([]string{}, args[:2]...)
		for i := 2; i < len(args); i++ {
			switch {
			case args[i] == "--user-source" || strings.HasPrefix(args[i], "--user-source="):
				fmt.Fprintln(os.Stderr, "flag provided but not defined: -user-source")
				os.Exit(1)
			case args[i] == "--source":
				i++
			case strings.HasPrefix(args[i], "--source="):
				continue
			default:
				legacyArgs = append(legacyArgs, args[i])
			}
		}
		args = legacyArgs
	}
	runPreviousBinary(path, args)
}

func runPreviousBinary(path string, args []string) {
	if err := syscall.Exec(path, append([]string{path}, args...), os.Environ()); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func main() {
	if manager := os.Getenv("AL_PROBE_PREVIOUS_MANAGER"); manager != "" {
		simulatePreviousManager(manager, os.Args[1:])
	}
	if engine := os.Getenv("AL_PROBE_PREVIOUS_ENGINE"); engine != "" {
		simulatePreviousEngine(engine, os.Args[1:])
	}
	// The previous protected worker offered v1, before publication-state pinning.
	if len(os.Args) == 3 && os.Args[1] == "capture" && os.Args[2] == "protected-worker-protocol" {
		fmt.Println("1")
		return
	}
	if len(os.Args) == 3 && os.Args[1] == "engine" && os.Args[2] == "protocol" {
		fmt.Println("1")
		return
	}

	if len(os.Args) == 3 && os.Args[1] == "version" && os.Args[2] == "--short" {
		fmt.Println(version)
		return
	}
	if len(os.Args) == 3 && os.Args[1] == "version" && os.Args[2] == "--json" {
		fmt.Printf("{\"version\":%q,\"commit\":\"fixture\",\"build_date\":\"fixture\"}\n", version)
		return
	}
	if application := os.Getenv("AL_PROBE_CAPTURE_APPLY"); application != "" && len(os.Args) > 1 && os.Args[1] == "dispatch" {
		if len(os.Args) == 3 && os.Args[2] == "protocol" {
			fmt.Println("1")
			return
		}
		command := exec.Command(application, os.Args[1:]...)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(89)
		}
		return
	}
	if len(os.Args) == 3 && os.Args[1] == "capture" && os.Args[2] == "protocol" {
		actualProtocol := os.Getenv("AL_PROBE_PROTOCOL")
		if actualProtocol == "" {
			actualProtocol = protocol
		}
		if actualProtocol == "none" {
			fmt.Fprintln(os.Stderr, "legacy release has no capture handler")
			os.Exit(1)
		}
		fmt.Println(actualProtocol)
		return
	}
	if len(os.Args) > 2 && os.Args[1] == "capture" && os.Args[2] == "check" && os.Getenv("AL_PROBE_HANG") == "1" {
		if path := os.Getenv("AL_PROBE_PID"); path != "" {
			if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
				os.Exit(95)
			}
		}
		for {
			time.Sleep(time.Hour)
		}
	}
	if application := os.Getenv("AL_PROBE_CAPTURE_APPLY"); application != "" && len(os.Args) > 2 && os.Args[1] == "capture" {
		command := exec.Command(application, os.Args[1:]...)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(89)
		}
		if os.Args[2] == "install" || os.Args[2] == "install-user" {
			if path := os.Getenv("AL_PROBE_PID"); path != "" {
				if err := os.WriteFile(path, []byte(strconv.Itoa(os.Getpid())), 0600); err != nil {
					os.Exit(88)
				}
			}
			switch os.Getenv("AL_PROBE_CAPTURE_RESULT") {
			case "hang":
				for {
					time.Sleep(time.Hour)
				}
			case "overflow":
				if _, err := os.Stdout.Write([]byte(strings.Repeat("x", 2*1024*1024))); err != nil {
					os.Exit(87)
				}
			case "both-streams":
				if _, err := os.Stderr.Write([]byte(strings.Repeat("diagnostic ", 2000))); err != nil {
					os.Exit(86)
				}
			case "exit":
				os.Exit(17)
			}
		}
		return
	}
	if value := os.Getenv("AL_PROBE_SELF_SIGNAL"); value != "" {
		number, err := strconv.Atoi(value)
		if err != nil {
			os.Exit(94)
		}
		process, err := os.FindProcess(os.Getpid())
		if err != nil {
			os.Exit(94)
		}
		if err := process.Signal(syscall.Signal(number)); err != nil {
			os.Exit(94)
		}
		for {
			time.Sleep(time.Hour)
		}
	}
	mode := os.Getenv("AL_PROBE_TRANSPORT_MODE")
	if mode != "" {
		for _, entry := range []struct {
			key   string
			value int
		}{{"AL_PROBE_PID", os.Getpid()}, {"AL_PROBE_PARENT_PID", os.Getppid()}} {
			if path := os.Getenv(entry.key); path != "" {
				if err := os.WriteFile(path, []byte(strconv.Itoa(entry.value)), 0600); err != nil {
					os.Exit(93)
				}
			}
		}
		if os.Getenv("AL_PROBE_READY") == "1" {
			fmt.Println("ready")
		}
		switch mode {
		case "delay":
			time.Sleep(1800 * time.Millisecond)
		case "stdin-close":
			if err := os.Stdin.Close(); err != nil {
				os.Exit(92)
			}
		case "stdout-overflow":
			if _, err := os.Stdout.Write([]byte(strings.Repeat("é", 40000))); err != nil {
				os.Exit(91)
			}
		case "stderr-overflow":
			if _, err := os.Stderr.Write([]byte(strings.Repeat("é", 40000))); err != nil {
				os.Exit(90)
			}
		}
		if mode != "delay" {
			for {
				time.Sleep(time.Hour)
			}
		}
	}
	input, err := io.ReadAll(os.Stdin)
	if err != nil {
		os.Exit(96)
	}
	cwd, err := os.Getwd()
	if err != nil {
		os.Exit(97)
	}
	result := struct {
		Version string   `json:"version"`
		Args    []string `json:"args"`
		Input   string   `json:"stdin"`
		CWD     string   `json:"cwd"`
	}{version, os.Args[1:], string(input), cwd}
	data, err := json.Marshal(result)
	if err != nil {
		os.Exit(98)
	}
	if path := os.Getenv("AL_PROBE_LOG"); path != "" {
		if err := os.WriteFile(path, data, 0600); err != nil {
			os.Exit(99)
		}
	}
	fmt.Println(string(data))
	if status := os.Getenv("AL_PROBE_EXIT"); status != "" {
		code, err := strconv.Atoi(status)
		if err != nil {
			os.Exit(94)
		}
		os.Exit(code)
	}
}
