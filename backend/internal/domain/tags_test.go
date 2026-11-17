package domain

import (
	"reflect"
	"testing"
)

func TestTicketTagsMergeAndWholeTicketSelection(t *testing.T) {
	saved := map[string]string{"team": "core", "environment": "dev"}
	got := MergeTags(saved, map[string]string{"environment": "prod"})
	if got["team"] != "core" || got["environment"] != "prod" || saved["environment"] != "dev" {
		t.Fatalf("tag precedence or immutability: %v / %v", got, saved)
	}
	if !reflect.DeepEqual(MergeTags(got, nil), got) {
		t.Fatal("absent input erased saved ticket tags")
	}
	filters := []TagFilter{{Key: "team", Values: []string{"other", "core"}}, {Key: "environment", Values: []string{"prod"}}}
	facts := AgenticFactLedger{Tags: got, Sessions: []SessionFact{{ID: "one"}, {ID: "two"}}}
	ledgers := []AgenticMetricsLedger{{Facts: facts}, {Facts: AgenticFactLedger{Tags: map[string]string{"team": "core"}}}}
	selected := FilterLedgersByTags(ledgers, filters)
	if len(selected) != 1 || len(selected[0].Facts.Sessions) != 2 {
		t.Fatal("filter must retain complete matching ticket")
	}
	if !MatchTicketTags(nil, []TagFilter{{Key: "team", Missing: true}}) || MatchTicketTags(map[string]string{"team": "Missing"}, []TagFilter{{Key: "team", Missing: true}}) {
		t.Fatal("Missing collided with literal value")
	}
	if len(FilterLedgersByTags(ledgers, nil)) != 2 {
		t.Fatal("no filter changed population")
	}
}

func TestNormalizeTagsRejectsCollisionsAndPreservesText(t *testing.T) {
	tags, err := NormalizeTags(map[string]string{" team ": " Core = one "})
	if err != nil || tags["team"] != "Core = one" {
		t.Fatalf("normalization %v %v", tags, err)
	}
	unicodeTags, err := NormalizeTags(map[string]string{"\ufeffcohort": "\ufeffvalue\ufeff", "\u0085experiment\u0085": "\u0085Canary\u0085"})
	if err != nil || !reflect.DeepEqual(unicodeTags, map[string]string{"\ufeffcohort": "\ufeffvalue\ufeff", "experiment": "Canary"}) {
		t.Fatalf("Unicode producer whitespace semantics: %v %v", unicodeTags, err)
	}
	if err := ValidateTags(unicodeTags); err != nil {
		t.Fatalf("canonical Unicode tags rejected: %v", err)
	}
	if err := ValidateTags(map[string]string{"experiment": "\u0085Canary"}); err == nil {
		t.Fatal("noncanonical Go Unicode whitespace accepted on read")
	}
	for _, bad := range []map[string]string{{"": "v"}, {"a": " "}, {"a": "x", " a ": "y"}} {
		if _, err := NormalizeTags(bad); err == nil {
			t.Fatalf("accepted %v", bad)
		}
	}
}
