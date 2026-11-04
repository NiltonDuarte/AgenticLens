package adapters

// canonicalFactLedgerDTO decodes the current schema only. Legacy inputs belong
// exclusively to the explicit versioned migration engine.
type canonicalFactLedgerDTO AgenticFactLedgerDTO

func decodeCanonicalFactLedger(data []byte, requiredCostEvidence bool) (AgenticFactLedgerDTO, error) {
	if err := validateGenerationPolicyJSON(data); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
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
