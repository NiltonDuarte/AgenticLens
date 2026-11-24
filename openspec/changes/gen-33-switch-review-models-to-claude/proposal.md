# GEN-33 Switch review models to Claude

The parallel review skill runs every selected reviewer under two model profiles. The user chose to replace the two GPT profiles with Claude Opus 5.5 at medium effort and Claude Sonnet 5.5 at high effort. Each selected reviewer still runs under both profiles, so coverage and attribution rules are unchanged.

The skill's dispatch guidance must also pass effort correctly to the Claude provider. It must also say how far a Claude reviewer's identity can be verified while AgenticLens cannot read Claude Code sessions ([AL-98](../../../alspaces/plan/to-do/AL-98%20generate%20ticket%20metrics%20from%20claude%20code%20sessions.md)).
