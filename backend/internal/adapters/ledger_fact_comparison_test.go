package adapters_test

import (
	"strings"
	"testing"

	"agenticlens/internal/adapters"
)

func TestLedgerFactsEqualPreservesExactWireSemantics(t *testing.T) {
	const accepted = `{"ticket_id":"TEST-100","updated_at":"2026-01-01","completed_at":"2026-01-01","sessions":[{"id":"a","tools":[{"name":"z","calls":9007199254740993},{"name":"a","calls":1}],"skills":[]}]}`
	cases := []struct {
		name      string
		candidate string
		equal     bool
		errorText string
	}{
		{
			name:      "root clocks and representation are immaterial",
			candidate: ` { "completed_at":"2026-01-02", "sessions":[{"skills":[],"tools":[{"calls":9.007199254740993e15,"name":"z"},{"calls":1.0,"name":"\u0061"}],"id":"a"}],"ticket_id":"TEST-100" } `,
			equal:     true,
		},
		{
			name:      "large integer differs by one",
			candidate: strings.Replace(accepted, "9007199254740993", "9007199254740992", 1),
		},
		{
			name:      "array sequence differs",
			candidate: strings.Replace(accepted, `{"name":"z","calls":9007199254740993},{"name":"a","calls":1}`, `{"name":"a","calls":1},{"name":"z","calls":9007199254740993}`, 1),
		},
		{
			name:      "absence differs from recorded empty",
			candidate: strings.Replace(accepted, `,"skills":[]`, "", 1),
		},
		{
			name:      "duplicate member cannot authorize a no-op",
			candidate: strings.Replace(accepted, `"ticket_id":"TEST-100"`, `"ticket_id":"TEST-100","ticket\u005fid":"TEST-100"`, 1),
			errorText: `duplicate JSON member "ticket_id"`,
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			equal, err := adapters.LedgerFactsEqual([]byte(accepted), []byte(tc.candidate))
			if tc.errorText != "" {
				if equal || err == nil || !strings.Contains(err.Error(), tc.errorText) {
					t.Fatalf("comparison returned (%t, %v), want false and %q", equal, err, tc.errorText)
				}
				return
			}
			if equal != tc.equal || err != nil {
				t.Fatalf("comparison returned (%t, %v), want (%t, nil)", equal, err, tc.equal)
			}
		})
	}
}
