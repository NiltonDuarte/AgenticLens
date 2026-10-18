package domain_test

import (
	"encoding/json"
	"strings"
	"testing"

	"agenticlens/internal/domain"
)

func TestGetBuildInfo_Defaults(t *testing.T) {
	info := domain.GetBuildInfo()

	if info.Version != domain.DefaultVersion {
		t.Errorf("expected default version %q, got %q", domain.DefaultVersion, info.Version)
	}
	if info.GitCommit != domain.DefaultGitCommit {
		t.Errorf("expected default git commit %q, got %q", domain.DefaultGitCommit, info.GitCommit)
	}
	if info.BuildDate != domain.DefaultBuildDate {
		t.Errorf("expected default build date %q, got %q", domain.DefaultBuildDate, info.BuildDate)
	}
}

func TestGetBuildInfo_EmptyVariableFallback(t *testing.T) {
	origVersion := domain.Version
	origCommit := domain.GitCommit
	origDate := domain.BuildDate
	t.Cleanup(func() {
		domain.Version = origVersion
		domain.GitCommit = origCommit
		domain.BuildDate = origDate
	})

	domain.Version = ""
	domain.GitCommit = ""
	domain.BuildDate = ""

	info := domain.GetBuildInfo()
	if info.Version != domain.DefaultVersion {
		t.Errorf("expected fallback version %q, got %q", domain.DefaultVersion, info.Version)
	}
	if info.GitCommit != domain.DefaultGitCommit {
		t.Errorf("expected fallback git commit %q, got %q", domain.DefaultGitCommit, info.GitCommit)
	}
	if info.BuildDate != domain.DefaultBuildDate {
		t.Errorf("expected fallback build date %q, got %q", domain.DefaultBuildDate, info.BuildDate)
	}
}

func TestGetBuildInfo_CustomVariables(t *testing.T) {
	origVersion := domain.Version
	origCommit := domain.GitCommit
	origDate := domain.BuildDate
	t.Cleanup(func() {
		domain.Version = origVersion
		domain.GitCommit = origCommit
		domain.BuildDate = origDate
	})

	domain.Version = "v1.2.3"
	domain.GitCommit = "abc1234"
	domain.BuildDate = "2026-10-18T12:00:00Z"

	info := domain.GetBuildInfo()
	if info.Version != "v1.2.3" {
		t.Errorf("expected version v1.2.3, got %q", info.Version)
	}
	if info.GitCommit != "abc1234" {
		t.Errorf("expected commit abc1234, got %q", info.GitCommit)
	}
	if info.BuildDate != "2026-10-18T12:00:00Z" {
		t.Errorf("expected build date 2026-10-18T12:00:00Z, got %q", info.BuildDate)
	}
}

func TestFormatBuildInfo_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    domain.BuildInfo
		expected string
	}{
		{
			name: "standard release",
			input: domain.BuildInfo{
				Version:   "v0.2.0",
				GitCommit: "7b8a9c0",
				BuildDate: "2026-10-18T15:30:00Z",
			},
			expected: "agentic-lens version v0.2.0 (commit: 7b8a9c0, built: 2026-10-18T15:30:00Z)",
		},
		{
			name: "dirty working tree tag",
			input: domain.BuildInfo{
				Version:   "v0.1.0-4-g9a8b7c-dirty",
				GitCommit: "9a8b7c0",
				BuildDate: "2026-10-18T16:00:00Z",
			},
			expected: "agentic-lens version v0.1.0-4-g9a8b7c-dirty (commit: 9a8b7c0, built: 2026-10-18T16:00:00Z)",
		},
		{
			name: "all empty fields fall back to defaults",
			input: domain.BuildInfo{
				Version:   "",
				GitCommit: "",
				BuildDate: "",
			},
			expected: "agentic-lens version v0.1.0-dev (commit: unknown, built: unknown)",
		},
		{
			name: "partial empty fields",
			input: domain.BuildInfo{
				Version:   "v1.0.0",
				GitCommit: "",
				BuildDate: "2026-10-18T17:00:00Z",
			},
			expected: "agentic-lens version v1.0.0 (commit: unknown, built: 2026-10-18T17:00:00Z)",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.FormatBuildInfo(tc.input)
			if result != tc.expected {
				t.Errorf("FormatBuildInfo() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestFormatShortVersion_TableDriven(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		input    domain.BuildInfo
		expected string
	}{
		{
			name: "standard release",
			input: domain.BuildInfo{
				Version:   "v1.5.0",
				GitCommit: "abcdef1",
				BuildDate: "2026-10-18T00:00:00Z",
			},
			expected: "v1.5.0",
		},
		{
			name: "empty version falls back to default",
			input: domain.BuildInfo{
				Version:   "",
				GitCommit: "abcdef1",
				BuildDate: "2026-10-18T00:00:00Z",
			},
			expected: domain.DefaultVersion,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			result := domain.FormatShortVersion(tc.input)
			if result != tc.expected {
				t.Errorf("FormatShortVersion() = %q, expected %q", result, tc.expected)
			}
		})
	}
}

func TestGetBuildInfo_IndividualEmptyFallbacks(t *testing.T) {
	origVersion := domain.Version
	origCommit := domain.GitCommit
	origDate := domain.BuildDate
	t.Cleanup(func() {
		domain.Version = origVersion
		domain.GitCommit = origCommit
		domain.BuildDate = origDate
	})

	tests := []struct {
		name         string
		setVersion   string
		setCommit    string
		setDate      string
		wantVersion  string
		wantCommit   string
		wantDate     string
	}{
		{
			name:        "only version empty",
			setVersion:  "",
			setCommit:   "commit123",
			setDate:     "2026-10-18T10:00:00Z",
			wantVersion: domain.DefaultVersion,
			wantCommit:  "commit123",
			wantDate:    "2026-10-18T10:00:00Z",
		},
		{
			name:        "only commit empty",
			setVersion:  "v2.0.0",
			setCommit:   "",
			setDate:     "2026-10-18T10:00:00Z",
			wantVersion: "v2.0.0",
			wantCommit:  domain.DefaultGitCommit,
			wantDate:    "2026-10-18T10:00:00Z",
		},
		{
			name:        "only build date empty",
			setVersion:  "v3.1.4",
			setCommit:   "deadbeef",
			setDate:     "",
			wantVersion: "v3.1.4",
			wantCommit:  "deadbeef",
			wantDate:    domain.DefaultBuildDate,
		},
		{
			name:        "version and commit empty",
			setVersion:  "",
			setCommit:   "",
			setDate:     "2026-10-18T12:00:00Z",
			wantVersion: domain.DefaultVersion,
			wantCommit:  domain.DefaultGitCommit,
			wantDate:    "2026-10-18T12:00:00Z",
		},
		{
			name:        "commit and date empty",
			setVersion:  "v1.0.0-rc.1",
			setCommit:   "",
			setDate:     "",
			wantVersion: "v1.0.0-rc.1",
			wantCommit:  domain.DefaultGitCommit,
			wantDate:    domain.DefaultBuildDate,
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			domain.Version = tc.setVersion
			domain.GitCommit = tc.setCommit
			domain.BuildDate = tc.setDate

			info := domain.GetBuildInfo()
			if info.Version != tc.wantVersion {
				t.Errorf("expected version %q, got %q", tc.wantVersion, info.Version)
			}
			if info.GitCommit != tc.wantCommit {
				t.Errorf("expected git commit %q, got %q", tc.wantCommit, info.GitCommit)
			}
			if info.BuildDate != tc.wantDate {
				t.Errorf("expected build date %q, got %q", tc.wantDate, info.BuildDate)
			}
		})
	}
}

func TestGetBuildInfo_ImmutabilityValueSemantics(t *testing.T) {
	origVersion := domain.Version
	origCommit := domain.GitCommit
	origDate := domain.BuildDate
	t.Cleanup(func() {
		domain.Version = origVersion
		domain.GitCommit = origCommit
		domain.BuildDate = origDate
	})

	domain.Version = "v1.0.0"
	domain.GitCommit = "initial123"
	domain.BuildDate = "2026-10-18T01:00:00Z"

	snapshot := domain.GetBuildInfo()

	// Mutate package variables
	domain.Version = "v9.9.9"
	domain.GitCommit = "mutated456"
	domain.BuildDate = "2026-12-31T23:59:59Z"

	// Assert snapshot remains untouched
	if snapshot.Version != "v1.0.0" {
		t.Errorf("snapshot.Version was mutated: got %q, want v1.0.0", snapshot.Version)
	}
	if snapshot.GitCommit != "initial123" {
		t.Errorf("snapshot.GitCommit was mutated: got %q, want initial123", snapshot.GitCommit)
	}
	if snapshot.BuildDate != "2026-10-18T01:00:00Z" {
		t.Errorf("snapshot.BuildDate was mutated: got %q, want 2026-10-18T01:00:00Z", snapshot.BuildDate)
	}
}

func TestBuildInfo_JSONSerializationAndBoundaryResilience(t *testing.T) {
	t.Parallel()
	t.Run("marshal standard build info produces expected keys", func(t *testing.T) {
		info := domain.BuildInfo{
			Version:   "v1.2.3",
			GitCommit: "7b8a9c0",
			BuildDate: "2026-10-18T15:30:00Z",
		}

		bytes, err := json.Marshal(info)
		if err != nil {
			t.Fatalf("json.Marshal failed: %v", err)
		}

		raw := string(bytes)
		if !strings.Contains(raw, `"version":"v1.2.3"`) {
			t.Errorf("missing version key: %s", raw)
		}
		if !strings.Contains(raw, `"commit":"7b8a9c0"`) {
			t.Errorf("missing commit key: %s", raw)
		}
		if !strings.Contains(raw, `"build_date":"2026-10-18T15:30:00Z"`) {
			t.Errorf("missing build_date key: %s", raw)
		}
	})

	t.Run("unmarshal valid JSON round trips cleanly", func(t *testing.T) {
		input := `{"version":"v2.1.0","commit":"abcdef0","build_date":"2026-10-18T18:00:00Z"}`
		var parsed domain.BuildInfo
		if err := json.Unmarshal([]byte(input), &parsed); err != nil {
			t.Fatalf("json.Unmarshal failed: %v", err)
		}

		if parsed.Version != "v2.1.0" || parsed.GitCommit != "abcdef0" || parsed.BuildDate != "2026-10-18T18:00:00Z" {
			t.Errorf("unexpected unmarshaled values: %+v", parsed)
		}
	})

	t.Run("unmarshal with unknown extra fields is safely ignored", func(t *testing.T) {
		input := `{"version":"v1.0.0","commit":"1234567","build_date":"2026-10-18T00:00:00Z","extra_key":"ignored","number":42}`
		var parsed domain.BuildInfo
		if err := json.Unmarshal([]byte(input), &parsed); err != nil {
			t.Fatalf("json.Unmarshal with unknown fields failed: %v", err)
		}
		if parsed.Version != "v1.0.0" {
			t.Errorf("expected version v1.0.0, got %q", parsed.Version)
		}
	})

	t.Run("unmarshal corrupted JSON returns wrapped error without panic", func(t *testing.T) {
		corrupted := []string{
			`{broken json`,
			`{"version": 12345}`,
			`not a json at all`,
			``,
		}

		for _, input := range corrupted {
			var parsed domain.BuildInfo
			err := json.Unmarshal([]byte(input), &parsed)
			if err == nil {
				t.Errorf("expected error unmarshaling corrupted input %q, got nil", input)
			}
		}
	})
}
