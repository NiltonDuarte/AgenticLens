package opencode_test

import "testing"

func TestSchemaV2ExplicitRootSelectors(t *testing.T) {
	exerciseExplicitRootSelectors(t, "v2")
}

func TestSchemaV2ExplicitRecoveryUsesConfiguredLiteralEvidence(t *testing.T) {
	exerciseExplicitRecovery(t, "v2")
}
