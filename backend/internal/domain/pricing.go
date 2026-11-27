package domain

import (
	"bytes"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"sort"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

// ModelPricingConfig holds operator-declared API-equivalent per-million USD rate rules.
type ModelPricingConfig struct {
	Rates []ModelPricingRate `json:"rates,omitempty" yaml:"rates,omitempty"`
}

// ModelPricingRate defines per-million-token USD rates for a model or glob pattern.
type ModelPricingRate struct {
	Model               string  `json:"model" yaml:"model"`
	InputPerMillion     float64 `json:"input_per_million" yaml:"input_per_million"`
	CacheReadPerMillion float64 `json:"cache_read_per_million" yaml:"cache_read_per_million"`
	OutputPerMillion    float64 `json:"output_per_million" yaml:"output_per_million"`
}

// AppliedRateSnapshot records the exact configured rate rule applied when an estimate was computed.
type AppliedRateSnapshot struct {
	MatchedPattern      string  `json:"matched_pattern"`
	InputPerMillion     float64 `json:"input_per_million"`
	CacheReadPerMillion float64 `json:"cache_read_per_million"`
	OutputPerMillion    float64 `json:"output_per_million"`
	EffectiveCatalogID  string  `json:"effective_catalog_id"`
}

// CostEstimate stores a persisted API-equivalent USD estimate derived from an applied rate rule.
type CostEstimate struct {
	AmountUSD float64             `json:"amount_usd"`
	Rate      AppliedRateSnapshot `json:"rate"`
}

// AmbiguousModelPricingError reports an equal-specificity glob tie during model pricing lookup.
type AmbiguousModelPricingError struct {
	Model    string
	Patterns []string
}

func (e *AmbiguousModelPricingError) Error() string {
	return fmt.Sprintf("ambiguous model pricing match for %q: equal-specificity glob patterns %s", e.Model, strings.Join(e.Patterns, ", "))
}

type appliedRateSnapshotWire struct {
	MatchedPattern      *string  `json:"matched_pattern"`
	InputPerMillion     *float64 `json:"input_per_million"`
	CacheReadPerMillion *float64 `json:"cache_read_per_million"`
	OutputPerMillion    *float64 `json:"output_per_million"`
	EffectiveCatalogID  *string  `json:"effective_catalog_id"`
}

type costEstimateWire struct {
	AmountUSD *float64                 `json:"amount_usd"`
	Rate      *appliedRateSnapshotWire `json:"rate"`
}

// UnmarshalJSON enforces strict non-null required members on applied rate snapshots.
func (s *AppliedRateSnapshot) UnmarshalJSON(data []byte) error {
	var raw appliedRateSnapshotWire
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected trailing data in applied rate snapshot")
	}
	snapshot, err := appliedRateSnapshotFromWire(&raw)
	if err != nil {
		return err
	}
	*s = snapshot
	return nil
}

// UnmarshalJSON enforces strict non-null required members on persisted cost estimates.
func (e *CostEstimate) UnmarshalJSON(data []byte) error {
	var raw costEstimateWire
	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	if err := dec.Decode(&raw); err != nil {
		return err
	}
	var trailing json.RawMessage
	if err := dec.Decode(&trailing); err != io.EOF {
		return fmt.Errorf("unexpected trailing data in cost_estimate")
	}
	if raw.AmountUSD == nil {
		return fmt.Errorf("cost_estimate.amount_usd is required")
	}
	if !ValidCostAmount(*raw.AmountUSD) {
		return fmt.Errorf("cost_estimate.amount_usd must be a finite non-negative number")
	}
	snapshot, err := appliedRateSnapshotFromWire(raw.Rate)
	if err != nil {
		return fmt.Errorf("cost_estimate.rate: %w", err)
	}
	*e = CostEstimate{
		AmountUSD: normalizePricingZero(*raw.AmountUSD),
		Rate:      snapshot,
	}
	return nil
}

func appliedRateSnapshotFromWire(raw *appliedRateSnapshotWire) (AppliedRateSnapshot, error) {
	if raw == nil {
		return AppliedRateSnapshot{}, fmt.Errorf("rate is required")
	}
	if raw.MatchedPattern == nil || strings.TrimSpace(*raw.MatchedPattern) == "" {
		return AppliedRateSnapshot{}, fmt.Errorf("matched_pattern must be nonblank")
	}
	if raw.EffectiveCatalogID == nil || strings.TrimSpace(*raw.EffectiveCatalogID) == "" {
		return AppliedRateSnapshot{}, fmt.Errorf("effective_catalog_id must be nonblank")
	}
	if raw.InputPerMillion == nil || !ValidCostAmount(*raw.InputPerMillion) {
		return AppliedRateSnapshot{}, fmt.Errorf("input_per_million must be a finite non-negative number")
	}
	if raw.CacheReadPerMillion == nil || !ValidCostAmount(*raw.CacheReadPerMillion) {
		return AppliedRateSnapshot{}, fmt.Errorf("cache_read_per_million must be a finite non-negative number")
	}
	if raw.OutputPerMillion == nil || !ValidCostAmount(*raw.OutputPerMillion) {
		return AppliedRateSnapshot{}, fmt.Errorf("output_per_million must be a finite non-negative number")
	}
	return AppliedRateSnapshot{
		MatchedPattern:      *raw.MatchedPattern,
		InputPerMillion:     normalizePricingZero(*raw.InputPerMillion),
		CacheReadPerMillion: normalizePricingZero(*raw.CacheReadPerMillion),
		OutputPerMillion:    normalizePricingZero(*raw.OutputPerMillion),
		EffectiveCatalogID:  *raw.EffectiveCatalogID,
	}, nil
}

// CloneCostEstimate returns an independent copy of a CostEstimate pointer.
func CloneCostEstimate(estimate *CostEstimate) *CostEstimate {
	if estimate == nil {
		return nil
	}
	copyVal := *estimate
	return &copyVal
}

func normalizeModelName(raw string) string {
	return strings.ToLower(strings.TrimSpace(raw))
}

func normalizePricingZero(v float64) float64 {
	if v == 0 {
		return 0
	}
	return v
}

func isModelGlob(normalizedPattern string) bool {
	return strings.ContainsAny(normalizedPattern, "*?")
}

func validateModelGlobSyntax(pattern string) error {
	if pattern == "" {
		return fmt.Errorf("pattern must be nonblank")
	}
	for _, r := range pattern {
		if unicode.IsControl(r) {
			return fmt.Errorf("invalid glob syntax %q: control characters are not allowed", pattern)
		}
		if r == '[' || r == ']' || r == '\\' {
			return fmt.Errorf("invalid glob syntax %q: only '*' and '?' wildcards are supported", pattern)
		}
	}
	return nil
}

// matchModelGlob matches text against pattern where '*' matches any rune sequence
// (including '/') and '?' matches any single rune. Both inputs must be normalized.
func matchModelGlob(pattern, text string) bool {
	p := []rune(pattern)
	t := []rune(text)
	pi, ti := 0, 0
	starIdx, matchIdx := -1, 0
	for ti < len(t) {
		if pi < len(p) && (p[pi] == '?' || p[pi] == t[ti]) {
			pi++
			ti++
		} else if pi < len(p) && p[pi] == '*' {
			starIdx = pi
			matchIdx = ti
			pi++
		} else if starIdx != -1 {
			pi = starIdx + 1
			matchIdx++
			ti = matchIdx
		} else {
			return false
		}
	}
	for pi < len(p) && p[pi] == '*' {
		pi++
	}
	return pi == len(p)
}

func globLiteralRuneCount(pattern string) int {
	count := 0
	for _, r := range pattern {
		if r != '*' && r != '?' {
			count++
		}
	}
	return count
}

// ValidateModelPricingConfig validates operator-declared model pricing rules without I/O.
func ValidateModelPricingConfig(cfg ModelPricingConfig) error {
	seen := make(map[string]int, len(cfg.Rates))
	for i, rate := range cfg.Rates {
		norm := normalizeModelName(rate.Model)
		if norm == "" {
			return fmt.Errorf("model_pricing.rates[%d].model must be nonblank", i)
		}
		if err := validateModelGlobSyntax(norm); err != nil {
			return fmt.Errorf("model_pricing.rates[%d].model: %w", i, err)
		}
		if prevIdx, exists := seen[norm]; exists {
			return fmt.Errorf("model_pricing.rates[%d].model %q duplicates normalized rule at model_pricing.rates[%d]", i, rate.Model, prevIdx)
		}
		seen[norm] = i
		for _, field := range []struct {
			name  string
			value float64
		}{
			{"input_per_million", rate.InputPerMillion},
			{"cache_read_per_million", rate.CacheReadPerMillion},
			{"output_per_million", rate.OutputPerMillion},
		} {
			if !ValidCostAmount(field.value) {
				return fmt.Errorf("model_pricing.rates[%d].%s (%q) must be a finite non-negative number", i, field.name, rate.Model)
			}
		}
	}
	return nil
}

// EffectiveCatalogID computes the deterministic SHA-256 identity of the configured
// rate catalog. Empty catalogs return "".
func (c ModelPricingConfig) EffectiveCatalogID() string {
	if len(c.Rates) == 0 {
		return ""
	}
	type canonicalRate struct {
		model     string
		input     string
		cacheRead string
		output    string
	}
	entries := make([]canonicalRate, len(c.Rates))
	for i, rate := range c.Rates {
		entries[i] = canonicalRate{
			model:     normalizeModelName(rate.Model),
			input:     strconv.FormatFloat(normalizePricingZero(rate.InputPerMillion), 'f', -1, 64),
			cacheRead: strconv.FormatFloat(normalizePricingZero(rate.CacheReadPerMillion), 'f', -1, 64),
			output:    strconv.FormatFloat(normalizePricingZero(rate.OutputPerMillion), 'f', -1, 64),
		}
	}
	sort.Slice(entries, func(i, j int) bool {
		if entries[i].model != entries[j].model {
			return entries[i].model < entries[j].model
		}
		if entries[i].input != entries[j].input {
			return entries[i].input < entries[j].input
		}
		if entries[i].cacheRead != entries[j].cacheRead {
			return entries[i].cacheRead < entries[j].cacheRead
		}
		return entries[i].output < entries[j].output
	})
	var b strings.Builder
	for _, entry := range entries {
		b.WriteString(strconv.Quote(entry.model))
		b.WriteByte('\t')
		b.WriteString(entry.input)
		b.WriteByte('\t')
		b.WriteString(entry.cacheRead)
		b.WriteByte('\t')
		b.WriteString(entry.output)
		b.WriteByte('\n')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return fmt.Sprintf("sha256:%x", sum)
}

// MatchModelPricingRate deterministically selects the configured rate for model.
// Exact normalized matches precede glob patterns; matching globs rank by highest
// literal rune count and then longest normalized pattern length. Ties fail closed.
func (c ModelPricingConfig) MatchModelPricingRate(model string) (ModelPricingRate, bool, error) {
	norm := normalizeModelName(model)
	if norm == "" || len(c.Rates) == 0 {
		return ModelPricingRate{}, false, nil
	}

	var exactMatches []ModelPricingRate
	for _, rate := range c.Rates {
		ruleNorm := normalizeModelName(rate.Model)
		if ruleNorm == "" || isModelGlob(ruleNorm) {
			continue
		}
		if ruleNorm == norm {
			exactMatches = append(exactMatches, rate)
		}
	}
	if len(exactMatches) == 1 {
		return exactMatches[0], true, nil
	}
	if len(exactMatches) > 1 {
		patterns := make([]string, len(exactMatches))
		for i, r := range exactMatches {
			patterns[i] = normalizeModelName(r.Model)
		}
		sort.Strings(patterns)
		return ModelPricingRate{}, false, &AmbiguousModelPricingError{Model: model, Patterns: patterns}
	}

	bestLiterals := -1
	bestLen := -1
	var winners []ModelPricingRate
	var winnerPatterns []string
	for _, rate := range c.Rates {
		ruleNorm := normalizeModelName(rate.Model)
		if ruleNorm == "" || !isModelGlob(ruleNorm) {
			continue
		}
		if !matchModelGlob(ruleNorm, norm) {
			continue
		}
		literals := globLiteralRuneCount(ruleNorm)
		length := utf8.RuneCountInString(ruleNorm)
		if literals > bestLiterals || (literals == bestLiterals && length > bestLen) {
			bestLiterals = literals
			bestLen = length
			winners = []ModelPricingRate{rate}
			winnerPatterns = []string{ruleNorm}
		} else if literals == bestLiterals && length == bestLen {
			winners = append(winners, rate)
			winnerPatterns = append(winnerPatterns, ruleNorm)
		}
	}

	if len(winners) == 0 {
		return ModelPricingRate{}, false, nil
	}
	if len(winners) > 1 {
		sort.Strings(winnerPatterns)
		return ModelPricingRate{}, false, &AmbiguousModelPricingError{Model: model, Patterns: winnerPatterns}
	}
	return winners[0], true, nil
}

// ValidateCostEstimate verifies that a persisted estimate obeys measured-cost exclusivity and value bounds.
func ValidateCostEstimate(evidence CostEvidence, estimate *CostEstimate) error {
	if estimate == nil {
		return nil
	}
	if evidence.State != CostUnavailable {
		return fmt.Errorf("cost_estimate is only allowed when cost_evidence.state is unavailable, got %q", evidence.State)
	}
	if !ValidCostAmount(estimate.AmountUSD) {
		return fmt.Errorf("cost_estimate.amount_usd must be a finite non-negative number")
	}
	if strings.TrimSpace(estimate.Rate.MatchedPattern) == "" {
		return fmt.Errorf("cost_estimate.rate.matched_pattern must be nonblank")
	}
	if strings.TrimSpace(estimate.Rate.EffectiveCatalogID) == "" {
		return fmt.Errorf("cost_estimate.rate.effective_catalog_id must be nonblank")
	}
	if !ValidCostAmount(estimate.Rate.InputPerMillion) {
		return fmt.Errorf("cost_estimate.rate.input_per_million must be a finite non-negative number")
	}
	if !ValidCostAmount(estimate.Rate.CacheReadPerMillion) {
		return fmt.Errorf("cost_estimate.rate.cache_read_per_million must be a finite non-negative number")
	}
	if !ValidCostAmount(estimate.Rate.OutputPerMillion) {
		return fmt.Errorf("cost_estimate.rate.output_per_million must be a finite non-negative number")
	}
	return nil
}

// EstimateInvocationCost computes an API-equivalent USD estimate for an eligible
// unavailable-cost invocation without modifying measured spend or partial subtotals.
func EstimateInvocationCost(inv InvocationFact, cfg ModelPricingConfig) (*CostEstimate, error) {
	if len(cfg.Rates) == 0 {
		return nil, nil
	}
	if err := ValidateModelPricingConfig(cfg); err != nil {
		return nil, err
	}
	return estimateInvocationCostValidated(inv, cfg, cfg.EffectiveCatalogID())
}

func estimateInvocationCostValidated(inv InvocationFact, cfg ModelPricingConfig, catalogID string) (*CostEstimate, error) {
	evidence := EffectiveCost(inv.CostUSD, inv.CostEvidence)
	if evidence.State != CostUnavailable {
		return nil, nil
	}
	if normalizeModelName(inv.Model) == "" {
		return nil, nil
	}
	if inv.Tokens.Input < 0 || inv.Tokens.CacheRead < 0 || inv.Tokens.Output < 0 {
		return nil, nil
	}
	rate, matched, err := cfg.MatchModelPricingRate(inv.Model)
	if err != nil {
		return nil, err
	}
	if !matched {
		return nil, nil
	}
	inRate := normalizePricingZero(rate.InputPerMillion)
	cacheRate := normalizePricingZero(rate.CacheReadPerMillion)
	outRate := normalizePricingZero(rate.OutputPerMillion)
	raw := (float64(inv.Tokens.Input)*inRate +
		float64(inv.Tokens.CacheRead)*cacheRate +
		float64(inv.Tokens.Output)*outRate) / 1_000_000.0
	if !ValidCostAmount(raw) {
		return nil, fmt.Errorf("computed cost estimate for model %q is not a finite non-negative amount", inv.Model)
	}
	amount := normalizePricingZero(math.Round(raw*1e6) / 1e6)
	if !ValidCostAmount(amount) {
		return nil, fmt.Errorf("computed cost estimate for model %q is not a finite non-negative amount", inv.Model)
	}
	return &CostEstimate{
		AmountUSD: amount,
		Rate: AppliedRateSnapshot{
			MatchedPattern:      normalizeModelName(rate.Model),
			InputPerMillion:     inRate,
			CacheReadPerMillion: cacheRate,
			OutputPerMillion:    outRate,
			EffectiveCatalogID:  catalogID,
		},
	}, nil
}

func applyModelPricingToInvocationsValidated(invocations []InvocationFact, cfg ModelPricingConfig, catalogID string) ([]InvocationFact, error) {
	if invocations == nil {
		return nil, nil
	}
	result := make([]InvocationFact, len(invocations))
	for i, inv := range invocations {
		updated := inv
		if len(cfg.Rates) == 0 {
			updated.CostEstimate = nil
		} else {
			estimate, err := estimateInvocationCostValidated(inv, cfg, catalogID)
			if err != nil {
				return nil, err
			}
			updated.CostEstimate = estimate
		}
		result[i] = updated
	}
	return result, nil
}

// ApplyModelPricingToInvocations returns a new invocation slice with configured
// API-equivalent estimates applied to eligible unavailable-cost invocations.
func ApplyModelPricingToInvocations(invocations []InvocationFact, cfg ModelPricingConfig) ([]InvocationFact, error) {
	if invocations == nil {
		return nil, nil
	}
	if len(cfg.Rates) > 0 {
		if err := ValidateModelPricingConfig(cfg); err != nil {
			return nil, err
		}
	}
	return applyModelPricingToInvocationsValidated(invocations, cfg, cfg.EffectiveCatalogID())
}

func applyModelPricingToSessionValidated(session SessionFact, cfg ModelPricingConfig, catalogID string) (SessionFact, error) {
	updated := session
	invocations, err := applyModelPricingToInvocationsValidated(session.Invocations, cfg, catalogID)
	if err != nil {
		return SessionFact{}, err
	}
	updated.Invocations = invocations
	return updated, nil
}

// ApplyModelPricingToSession returns a copy of session with configured pricing
// applied to its eligible unavailable-cost invocations.
func ApplyModelPricingToSession(session SessionFact, cfg ModelPricingConfig) (SessionFact, error) {
	updated := session
	invocations, err := ApplyModelPricingToInvocations(session.Invocations, cfg)
	if err != nil {
		return SessionFact{}, err
	}
	updated.Invocations = invocations
	return updated, nil
}

// ApplyModelPricingToSessions returns a copy of sessions with configured pricing
// applied to all eligible unavailable-cost invocations across every session.
func ApplyModelPricingToSessions(sessions []SessionFact, cfg ModelPricingConfig) ([]SessionFact, error) {
	if sessions == nil {
		return nil, nil
	}
	if len(cfg.Rates) > 0 {
		if err := ValidateModelPricingConfig(cfg); err != nil {
			return nil, err
		}
	}
	catalogID := cfg.EffectiveCatalogID()
	result := make([]SessionFact, len(sessions))
	for i, session := range sessions {
		updated, err := applyModelPricingToSessionValidated(session, cfg, catalogID)
		if err != nil {
			return nil, err
		}
		result[i] = updated
	}
	return result, nil
}

// DeduplicateAndSortAppliedRates returns a deterministic sorted slice of unique applied rate snapshots.
func DeduplicateAndSortAppliedRates(rates []AppliedRateSnapshot) []AppliedRateSnapshot {
	if len(rates) == 0 {
		return nil
	}
	seen := make(map[AppliedRateSnapshot]struct{}, len(rates))
	unique := make([]AppliedRateSnapshot, 0, len(rates))
	for _, r := range rates {
		normalized := AppliedRateSnapshot{
			MatchedPattern:      r.MatchedPattern,
			InputPerMillion:     normalizePricingZero(r.InputPerMillion),
			CacheReadPerMillion: normalizePricingZero(r.CacheReadPerMillion),
			OutputPerMillion:    normalizePricingZero(r.OutputPerMillion),
			EffectiveCatalogID:  r.EffectiveCatalogID,
		}
		if _, exists := seen[normalized]; exists {
			continue
		}
		seen[normalized] = struct{}{}
		unique = append(unique, normalized)
	}
	sort.Slice(unique, func(i, j int) bool {
		if unique[i].MatchedPattern != unique[j].MatchedPattern {
			return unique[i].MatchedPattern < unique[j].MatchedPattern
		}
		if unique[i].EffectiveCatalogID != unique[j].EffectiveCatalogID {
			return unique[i].EffectiveCatalogID < unique[j].EffectiveCatalogID
		}
		if unique[i].InputPerMillion != unique[j].InputPerMillion {
			return unique[i].InputPerMillion < unique[j].InputPerMillion
		}
		if unique[i].CacheReadPerMillion != unique[j].CacheReadPerMillion {
			return unique[i].CacheReadPerMillion < unique[j].CacheReadPerMillion
		}
		return unique[i].OutputPerMillion < unique[j].OutputPerMillion
	})
	return unique
}
