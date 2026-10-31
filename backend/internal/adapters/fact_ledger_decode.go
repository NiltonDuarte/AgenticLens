package adapters

// canonicalFactLedgerDTO is the shared v4/v5 root layout. Their required cost
// evidence policies remain explicit at the named version entry points.
type canonicalFactLedgerDTO AgenticFactLedgerDTO

func decodeSchemaV4FactLedger(data []byte) (AgenticFactLedgerDTO, error) {
	return decodeCanonicalFactLedger(data, false)
}
func decodeSchemaV5FactLedger(data []byte) (AgenticFactLedgerDTO, error) {
	return decodeCanonicalFactLedger(data, true)
}
func decodeCanonicalFactLedger(data []byte, requiredCostEvidence bool) (AgenticFactLedgerDTO, error) {
	var wire canonicalFactLedgerDTO
	decoder := newStrictFactDecoder(data)
	if err := decoder.Decode(&wire); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	if err := requireFactJSONEOF(decoder, "fact ledger"); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	result := AgenticFactLedgerDTO(wire)
	if err := result.validateCostFields(requiredCostEvidence); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	return result, nil
}
