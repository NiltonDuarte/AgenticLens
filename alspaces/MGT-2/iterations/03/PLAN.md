# MGT-2 iteration 03: route newly discovered defects

Source: user explicitly requested a management skill rule requiring a new FIX ticket/thread for bugs discovered during supervision unless very closely related to an existing ticket. User granted proactive launch authority for those fixes. This is an approved management-policy delta, not an application implementation request.

Changed acceptance: classify relation before dispatch. Direct regression of the same approved behavior/contract may stay with its owner; shared files, the same failing suite or blocking another PR alone are insufficient. Otherwise allocate an unused FIX identifier and launch an isolated independent owner using the common launch/ownership protocol. Reuse an existing FIX owner; preserve the original ticket's scope and completed work.

Superseded decision: treating every new independent thread as needing another user request despite this standing authorization. A replacement root for a missing existing owner still requires a user request. Review-only authority, material product decisions, human-only release merges, external management merge ownership and unrequested roadmap holds remain unchanged. Existing explicit repair assignments are not retroactively moved into new tickets.

Owned writes: management skill SKILL.md and MGT-2 artifacts only. Initial checkout clean at a1a0040; current upstream after verified #105 merge is 3fbecd2576270772462fe37235d90c5f33e2164c. No application, test or runtime-script edits. No new FIX launch was triggered by this instruction update itself.

Implementation/verification: added one routing section and reconciled the two older launch-authority passages in the canonical skill. Searched active AGENTS/contributor/docs/skill guidance; AGENTS already links to the canonical management skill and no conflicting duplicate rule needed editing. Skill-creator guidance applied. quick_validate.py reports Skill is valid; git diff --check passes. No application suite rerun for documentation-only changes.

Behavior check: unrelated capture installer failure found during a UI feature requires a new FIX; a new formatter producing impossible dates remains with its feature owner; an already owned identical FIX reuses that thread; a review-only request does not launch execution. These inspect decisions rather than adding wording-matching tests.

Parallel delivery preserved: #105 merged at 3fbecd2 after all four checks passed on c031576; #104 owner resumed in run 6 to integrate the actual test-only repair and finish affected verification. #102 and #103 retain their pending predecessors. AL-62's bounded receipt-proof repair continues under its prior assignment. Supervision remains active; management PR requires external merge.
