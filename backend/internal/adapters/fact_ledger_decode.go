package adapters

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
)

// schemaV4FactLedgerDTO shares the canonical field layout with v5 while keeping
// numeric legacy cost optional and evidence optional at its own decode boundary.
type schemaV4FactLedgerDTO AgenticFactLedgerDTO

// schemaV5FactLedgerDTO requires both cost fields at each known invocation/task
// location. Presence is checked before nullable pointers are projected to domain.
type schemaV5FactLedgerDTO AgenticFactLedgerDTO

func decodeSchemaV4FactLedger(data []byte) (AgenticFactLedgerDTO, error) {
	var wire schemaV4FactLedgerDTO
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return AgenticFactLedgerDTO{}, errors.New("multiple JSON values in fact ledger")
		}
		return AgenticFactLedgerDTO{}, err
	}
	result := AgenticFactLedgerDTO(wire)
	if err := result.validateCostFields(false); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	return result, nil
}
func decodeSchemaV5FactLedger(data []byte) (AgenticFactLedgerDTO, error) {
	var wire schemaV5FactLedgerDTO
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&wire); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	var trailing json.RawMessage
	if err := decoder.Decode(&trailing); err != io.EOF {
		if err == nil {
			return AgenticFactLedgerDTO{}, errors.New("multiple JSON values in fact ledger")
		}
		return AgenticFactLedgerDTO{}, err
	}
	result := AgenticFactLedgerDTO(wire)
	if err := result.validateCostFields(true); err != nil {
		return AgenticFactLedgerDTO{}, err
	}
	return result, nil
}
