package migration

import "agenticlens/internal/domain"

// StepV5ToV6OwnedAuxiliary preserves unresolved legacy facts while expanding owners.
type StepV5ToV6OwnedAuxiliary struct{}

func NewStepV5ToV6OwnedAuxiliary() *StepV5ToV6OwnedAuxiliary { return &StepV5ToV6OwnedAuxiliary{} }
func (*StepV5ToV6OwnedAuxiliary) SourceVersion() int         { return 5 }
func (*StepV5ToV6OwnedAuxiliary) TargetVersion() int         { return 6 }
func (*StepV5ToV6OwnedAuxiliary) Description() string {
	return "Adds session-owned skill and delegation occurrences"
}
func (*StepV5ToV6OwnedAuxiliary) Apply(ledger domain.AgenticFactLedger) (domain.AgenticFactLedger, error) {
	cloned := domain.NormalizeAuxiliaryOwnership(ledger)
	version := 6
	cloned.SchemaVersion = &version
	if cloned.Migration != nil {
		info := *cloned.Migration
		info.SchemaVersion = version
		info.StepsCompleted = append(append([]string{}, info.StepsCompleted...), "schema_v6_owned_auxiliary")
		cloned.Migration = &info
	}
	return cloned, nil
}
