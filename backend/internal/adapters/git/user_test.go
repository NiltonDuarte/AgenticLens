package git

import "testing"

func TestResolveSessionUser(t *testing.T) {
	tests := []struct {
		name      string
		extracted string
		existing  string
		gitUser   string
		want      string
	}{
		{
			name:      "unknown cached identity falls through to git",
			extracted: " UNKNOWN ",
			existing:  " UnKnOwN ",
			gitUser:   " Git User ",
			want:      "Git User",
		},
		{
			name:      "known session identity precedes git",
			extracted: "unknown",
			existing:  " Session User ",
			gitUser:   "Git User",
			want:      "Session User",
		},
		{
			name:      "known extracted identity is preserved",
			extracted: " Extracted User ",
			existing:  "Session User",
			gitUser:   "Git User",
			want:      "Extracted User",
		},
		{
			name:      "unknown final fallback",
			extracted: " ",
			existing:  "UNKNOWN",
			gitUser:   " unknown ",
			want:      "unknown",
		},
		{
			name:      "other names containing unknown are valid",
			extracted: "Unknown User",
			want:      "Unknown User",
		},
		{
			name:      "blank extracted and cached identities fall through",
			extracted: " \t ",
			existing:  "\n",
			gitUser:   "Fallback User",
			want:      "Fallback User",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ResolveSessionUser(tt.extracted, tt.existing, tt.gitUser); got != tt.want {
				t.Fatalf("ResolveSessionUser(%q, %q, %q) = %q, want %q", tt.extracted, tt.existing, tt.gitUser, got, tt.want)
			}
		})
	}
}
