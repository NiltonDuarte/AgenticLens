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
	"time"
)

var version = "v9.8.7"
var protocol = "3"

func main() {
	if len(os.Args) == 3 && os.Args[1] == "version" && os.Args[2] == "--short" {
		fmt.Println(version)
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
	if application := os.Getenv("AL_PROBE_CAPTURE_APPLY"); application != "" && len(os.Args) > 2 && os.Args[1] == "capture" {
		command := exec.Command(application, os.Args[1:]...)
		command.Stdin, command.Stdout, command.Stderr = os.Stdin, os.Stdout, os.Stderr
		if err := command.Run(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(89)
		}
		if os.Args[2] == "install" {
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
