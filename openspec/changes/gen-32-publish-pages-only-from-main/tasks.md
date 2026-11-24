# GEN-32: Execution tasks

Reference design.md for approved scope and specs/ for observable acceptance.

## Implementation and focused verification
- [x] Restrict the workflow to pushes to main and dispatch from main. Remove the PR trigger, preview input, candidate job and preview reporting.
- [x] Make composition drop retained previews after validating them, and refuse non-main refs. Remove the PR listing.
- [x] Read the Pages artifact up to three times before recording the transaction, and refuse re-run attempts of the publish job.
- [x] Replace the preview tests with main-only, legacy-preview, unsafe-input and ref tests, and update the Pages README.

## Integration and delivery
- [x] Validate the workflow (YAML parse, references, actionlint) and run the Pages suite.
- [x] Run the integrated make gate, `git diff --check` and strict OpenSpec validation.
- [x] Record delivery, note unavailable Claude Code telemetry, and open the PR.

## Workflow follow-up
Confirm the first main run after merge deploys and leaves no previews in the site or on `portal-pages-state`. Archive only after that.
