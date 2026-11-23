package main

import (
	"fmt"

	"agenticlens/internal/adapters/backfill"
)

func printBackfillOwnerCoverage(result backfill.TicketMigrationResult) {
	for _, coverage := range result.Coverage {
		fmt.Printf("  %s refresh: %s", coverage.Harness, coverage.Status)
		if coverage.Reason != "" {
			fmt.Printf(" (%s)", coverage.Reason)
		}
		if coverage.OwnerInventoryUnknown {
			fmt.Print(" | Recorded owner inventory: unknown")
			if coverage.InventoryError != "" {
				fmt.Printf(" (%s)", coverage.InventoryError)
			}
			fmt.Println()
			continue
		}
		printBackfillRefreshCounts(coverage.RefreshCounts)
		for _, owner := range coverage.Owners {
			fmt.Printf("    Session %s: %s", owner.SessionID, owner.Status)
			if owner.Reason != "" {
				fmt.Printf(" (%s)", owner.Reason)
			}
			fmt.Printf(" | Skills: %s | Tasks: %s\n", owner.Skills, owner.Tasks)
		}
		if coverage.Retained > 0 || coverage.BudgetLimited {
			fmt.Printf("    Retry exact recorded sessions with --recompute-db, --ticket-id %q and --source %s. Use --workspace %q to disambiguate the ledger.\n",
				result.TicketID, coverage.Harness, result.WorkspacePath)
		}
		if coverage.BudgetLimited {
			fmt.Println("    An exact-ticket retry starts a fresh batch but uses the same configured session and byte limits per run. Review those limits if the ticket still exceeds them.")
		}
	}
}

func printBackfillRefreshCounts(counts backfill.RefreshCounts) {
	fmt.Printf(" | Recorded: %d | Refreshed: %d | Verified unchanged: %d | Retained: %d | Discovered: %d",
		counts.Recorded, counts.Refreshed, counts.VerifiedUnchanged, counts.Retained, counts.Discovered)
	if counts.Proposed > 0 || counts.ProposedDiscovered > 0 {
		fmt.Printf(" | Verified changes not committed: %d | Discovered not committed: %d", counts.Proposed, counts.ProposedDiscovered)
	}
	if counts.VerifiedUncompared > 0 {
		fmt.Printf(" | Verified reads without comparison: %d", counts.VerifiedUncompared)
	}
	fmt.Println()
}
