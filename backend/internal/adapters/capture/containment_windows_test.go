package capture

import "testing"

// Synthetic operands exercise native Windows filepath semantics without creating
// either drive path or claiming a two-volume application installation.
func TestWindowsStateContainmentTreatsDistinctDrivesAsExternal(t *testing.T) {
	project := `D:\repo`
	registry := `C:\Users\fixture\.agenticlens\state\projects\Fixture`
	if actual, err := validateStatePlacement(project, registry); err != nil || actual != registry {
		t.Fatalf("different-volume home rejected: %q %v", actual, err)
	}
	if relative, contained, err := relativeContainedPath(project, registry); err != nil || contained || relative != "" {
		t.Fatalf("external registry became a project ignore: relative=%q contained=%t err=%v", relative, contained, err)
	}
	forbidden := `D:\repo\.codex\hooks.json\state\projects\Fixture`
	if _, err := validateStatePlacement(project, forbidden); err == nil {
		t.Fatal("same-volume owned-file descendant accepted")
	}
	local := `D:\repo\.capture-local\projects\Fixture`
	if relative, contained, err := relativeContainedPath(project, local); err != nil || !contained || relative != `.capture-local\projects\Fixture` {
		t.Fatalf("local state containment lost: %q %t %v", relative, contained, err)
	}
	if _, _, err := relativeContainedPath(project, `D:relative`); err == nil {
		t.Fatal("genuine same-volume relative resolution error suppressed")
	}
}
