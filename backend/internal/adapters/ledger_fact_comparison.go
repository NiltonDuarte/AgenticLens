package adapters

import (
	"bytes"
	"cmp"
	"encoding/json"
	"fmt"
	"math/big"
	"slices"
	"strconv"
)

// LedgerFactsEqual compares the accepted and candidate wire facts without their
// root publication timestamps. Object presentation is immaterial, while arrays
// and nested timestamps remain part of the recorded facts.
func LedgerFactsEqual(accepted, candidate []byte) (bool, error) {
	prior, err := parseLedgerFormat(accepted)
	if err != nil {
		return false, fmt.Errorf("comparing accepted ledger: %w", err)
	}
	next, err := parseLedgerFormat(candidate)
	if err != nil {
		return false, fmt.Errorf("comparing candidate ledger: %w", err)
	}
	publicationTimestamp := func(member formatMember) bool {
		return member.key == "updated_at" || member.key == "completed_at"
	}
	prior.members = slices.DeleteFunc(prior.members, publicationTimestamp)
	next.members = slices.DeleteFunc(next.members, publicationTimestamp)
	return ledgerValueEqual(prior, next), nil
}

func ledgerValueEqual(a, b *formatValue) bool {
	switch a.kind {
	case '{':
		if b.kind != '{' || len(a.members) != len(b.members) {
			return false
		}
		byKey := func(a, b formatMember) int { return cmp.Compare(a.key, b.key) }
		slices.SortFunc(a.members, byKey)
		slices.SortFunc(b.members, byKey)
		return slices.EqualFunc(a.members, b.members, func(a, b formatMember) bool {
			return a.key == b.key && ledgerValueEqual(a.value, b.value)
		})
	case '[':
		return b.kind == '[' && slices.EqualFunc(a.elements, b.elements, ledgerValueEqual)
	case '"':
		if b.kind != '"' {
			return false
		}
		if bytes.Equal(a.raw, b.raw) {
			return true
		}
		var left, right string
		return json.Unmarshal(a.raw, &left) == nil && json.Unmarshal(b.raw, &right) == nil && left == right
	default:
		return bytes.Equal(a.raw, b.raw) || ledgerNumberEqual(a.raw, b.raw)
	}
}

func ledgerNumberEqual(a, b []byte) bool {
	parse := func(raw []byte) (*big.Rat, bool) {
		// Larger representations stay distinct rather than allocating arbitrary
		// precision storage for spellings outside the writer's numeric range.
		if len(raw) == 0 || len(raw) > 1024 || (raw[0] != '-' && (raw[0] < '0' || raw[0] > '9')) {
			return nil, false
		}
		if i := bytes.IndexAny(raw, "eE"); i >= 0 {
			exponent, err := strconv.ParseInt(string(raw[i+1:]), 10, 16)
			if err != nil || exponent < -1024 || exponent > 1024 {
				return nil, false
			}
		}
		return new(big.Rat).SetString(string(raw))
	}
	left, ok := parse(a)
	if !ok {
		return false
	}
	right, ok := parse(b)
	return ok && left.Cmp(right) == 0
}
