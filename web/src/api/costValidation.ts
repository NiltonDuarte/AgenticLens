import type { CostProvenance, EquivalentCostSummary } from '../domain/types';

function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isAmount(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0;
}

function isNullableAmount(value: unknown): value is number | null {
  return value === null || isAmount(value);
}

function isNonNegativeSafeInteger(value: unknown): value is number {
  return typeof value === 'number' && Number.isSafeInteger(value) && value >= 0;
}

function isNonBlankString(value: unknown): value is string {
  return typeof value === 'string' && value.trim().length > 0;
}

function isCostProvenance(value: unknown): value is CostProvenance {
  return (
    value === 'empty' ||
    value === 'measured' ||
    value === 'estimated' ||
    value === 'hybrid' ||
    value === 'partial' ||
    value === 'unavailable'
  );
}

function sameNullableAmount(a: number | null, b: number | null): boolean {
  if (a === null || b === null) return a === b;
  return Math.abs(a - b) <= 1e-9;
}

function invalid(path: string): never {
  throw new Error(`Invalid cost payload at ${path}`);
}

function validateAppliedRateSnapshot(value: unknown, path: string): void {
  if (!isRecord(value)) invalid(path);
  if (!isNonBlankString(value.matched_pattern)) invalid(`${path}.matched_pattern`);
  if (!isNonBlankString(value.effective_catalog_id)) invalid(`${path}.effective_catalog_id`);
  if (!isAmount(value.input_per_million)) invalid(`${path}.input_per_million`);
  if (!isAmount(value.cache_read_per_million)) invalid(`${path}.cache_read_per_million`);
  if (!isAmount(value.output_per_million)) invalid(`${path}.output_per_million`);
}

function validateCostEstimate(value: unknown, evidence: Record<string, unknown>, path: string): void {
  if (evidence.state !== 'unavailable') invalid(path);
  if (!isRecord(value)) invalid(path);
  if (!isAmount(value.amount_usd)) invalid(`${path}.amount_usd`);
  validateAppliedRateSnapshot(value.rate, `${path}.rate`);
}

function validateEvidence(value: unknown, amount: unknown, path: string): Record<string, unknown> {
  if (!isRecord(value)) invalid(path);
  const subtotal = value.known_subtotal_usd;
  if (value.state === 'observed') {
    if (!isAmount(amount) || subtotal !== amount) invalid(path);
  } else if (value.state === 'partial') {
    if (amount !== null || !isAmount(subtotal)) invalid(path);
  } else if (value.state === 'unavailable') {
    if (amount !== null || subtotal !== null) invalid(path);
  } else invalid(path);
  return value;
}

function validateSummary(value: unknown, amount: unknown, path: string): void {
  if (!isRecord(value) || !isNullableAmount(amount) || !isNullableAmount(value.known_subtotal_usd)) invalid(path);
  const observed = value.fully_observed;
  const partial = value.partial;
  const unavailable = value.unavailable;
  const total = value.total;
  if (
    !isNonNegativeSafeInteger(observed) ||
    !isNonNegativeSafeInteger(partial) ||
    !isNonNegativeSafeInteger(unavailable) ||
    !isNonNegativeSafeInteger(total)
  ) {
    invalid(path);
  }
  if (observed + partial + unavailable !== total) invalid(path);
  const subtotal = value.known_subtotal_usd;
  if (value.state === 'empty') {
    if (total !== 0 || amount !== null || subtotal !== null) invalid(path);
  } else if (value.state === 'complete') {
    if (total === 0 || observed !== total || !isAmount(amount) || amount !== subtotal) invalid(path);
  } else if (value.state === 'partial') {
    if (total === 0 || observed === total || observed + partial === 0 || amount !== null || !isAmount(subtotal)) invalid(path);
  } else if (value.state === 'unavailable') {
    if (total === 0 || unavailable !== total || amount !== null || subtotal !== null) invalid(path);
  } else invalid(path);
}

export function validateEquivalentCostSummary(value: unknown, path: string): EquivalentCostSummary {
  if (!isRecord(value)) invalid(path);
  const provenance = value.provenance;
  if (!isCostProvenance(provenance)) invalid(`${path}.provenance`);

  const measuredCost = value.measured_cost_usd;
  const estimatedCost = value.estimated_cost_usd;
  const totalEquivalent = value.total_equivalent_cost_usd;
  const knownSubtotal = value.known_subtotal_usd;
  if (
    !hasOwn(value, 'measured_cost_usd') || !isNullableAmount(measuredCost) ||
    !hasOwn(value, 'estimated_cost_usd') || !isNullableAmount(estimatedCost) ||
    !hasOwn(value, 'total_equivalent_cost_usd') || !isNullableAmount(totalEquivalent) ||
    !hasOwn(value, 'known_subtotal_usd') || !isNullableAmount(knownSubtotal)
  ) {
    invalid(path);
  }

  const measured = value.measured;
  const estimated = value.estimated;
  const partialMeasured = value.partial_measured;
  const unavailable = value.unavailable;
  const total = value.total;
  if (
    !isNonNegativeSafeInteger(measured) ||
    !isNonNegativeSafeInteger(estimated) ||
    !isNonNegativeSafeInteger(partialMeasured) ||
    !isNonNegativeSafeInteger(unavailable) ||
    !isNonNegativeSafeInteger(total)
  ) {
    invalid(path);
  }
  if (measured + estimated + partialMeasured + unavailable !== total) invalid(path);

  if (measured + partialMeasured === 0) {
    if (measuredCost !== null) invalid(path);
  } else if (!isAmount(measuredCost)) {
    invalid(path);
  }
  if (estimated === 0) {
    if (estimatedCost !== null) invalid(path);
  } else if (!isAmount(estimatedCost)) {
    invalid(path);
  }

  switch (provenance) {
    case 'empty':
      if (total !== 0 || measuredCost !== null || estimatedCost !== null || totalEquivalent !== null || knownSubtotal !== null) {
        invalid(path);
      }
      break;
    case 'measured':
      if (
        total === 0 ||
        measured !== total ||
        !isAmount(measuredCost) ||
        estimatedCost !== null ||
        !isAmount(totalEquivalent) ||
        !isAmount(knownSubtotal) ||
        !sameNullableAmount(totalEquivalent, measuredCost) ||
        !sameNullableAmount(knownSubtotal, measuredCost)
      ) {
        invalid(path);
      }
      break;
    case 'estimated':
      if (
        total === 0 ||
        estimated !== total ||
        measuredCost !== null ||
        !isAmount(estimatedCost) ||
        !isAmount(totalEquivalent) ||
        !isAmount(knownSubtotal) ||
        !sameNullableAmount(totalEquivalent, estimatedCost) ||
        !sameNullableAmount(knownSubtotal, estimatedCost)
      ) {
        invalid(path);
      }
      break;
    case 'hybrid':
      if (
        measured === 0 ||
        estimated === 0 ||
        measured + estimated !== total ||
        !isAmount(measuredCost) ||
        !isAmount(estimatedCost) ||
        !isAmount(knownSubtotal) ||
        !isAmount(totalEquivalent) ||
        !sameNullableAmount(knownSubtotal, measuredCost + estimatedCost) ||
        !sameNullableAmount(totalEquivalent, knownSubtotal)
      ) {
        invalid(path);
      }
      break;
    case 'partial':
      if (
        total === 0 ||
        partialMeasured + unavailable === 0 ||
        measured + estimated + partialMeasured === 0 ||
        totalEquivalent !== null ||
        !isAmount(knownSubtotal) ||
        !sameNullableAmount(knownSubtotal, (measuredCost ?? 0) + (estimatedCost ?? 0))
      ) {
        invalid(path);
      }
      break;
    case 'unavailable':
      if (
        total === 0 ||
        unavailable !== total ||
        measuredCost !== null ||
        estimatedCost !== null ||
        totalEquivalent !== null ||
        knownSubtotal !== null
      ) {
        invalid(path);
      }
      break;
  }

  return {
    provenance,
    measured_cost_usd: measuredCost,
    estimated_cost_usd: estimatedCost,
    total_equivalent_cost_usd: totalEquivalent,
    known_subtotal_usd: knownSubtotal,
    measured,
    estimated,
    partial_measured: partialMeasured,
    unavailable,
    total,
  };
}

function validateEquivalentAgainstMeasured(
  equiv: EquivalentCostSummary,
  measuredSummary: unknown,
  path: string
): void {
  if (!isRecord(measuredSummary)) return;
  const knownSubtotal = measuredSummary.known_subtotal_usd;
  if (!isNullableAmount(knownSubtotal)) invalid(path);
  if (
    equiv.total !== measuredSummary.total ||
    equiv.measured !== measuredSummary.fully_observed ||
    equiv.partial_measured !== measuredSummary.partial ||
    equiv.estimated + equiv.unavailable !== measuredSummary.unavailable ||
    !sameNullableAmount(equiv.measured_cost_usd, knownSubtotal)
  ) {
    invalid(path);
  }
}

function validateEquivalentFields(
  record: Record<string, unknown>,
  path: string,
  measuredSummary: unknown,
  prefix = ''
): void {
  const measuredKey = `${prefix}measured_cost_usd`;
  const estimatedKey = `${prefix}estimated_cost_usd`;
  const totalEquivKey = `${prefix}total_equivalent_cost_usd`;
  const provenanceKey = `${prefix}cost_provenance`;
  const summaryKey = `${prefix}equivalent_cost_summary`;

  const keys = [measuredKey, estimatedKey, totalEquivKey, provenanceKey, summaryKey];
  if (!keys.some((key) => hasOwn(record, key) && record[key] !== undefined)) {
    return;
  }

  for (const key of keys) {
    if (!hasOwn(record, key) || record[key] === undefined) {
      invalid(`${path}.${key}`);
    }
  }

  const equiv = validateEquivalentCostSummary(record[summaryKey], `${path}.${summaryKey}`);

  if (!isNullableAmount(record[measuredKey]) || !sameNullableAmount(record[measuredKey], equiv.measured_cost_usd)) {
    invalid(`${path}.${measuredKey}`);
  }
  if (!isNullableAmount(record[estimatedKey]) || !sameNullableAmount(record[estimatedKey], equiv.estimated_cost_usd)) {
    invalid(`${path}.${estimatedKey}`);
  }
  if (!isNullableAmount(record[totalEquivKey]) || !sameNullableAmount(record[totalEquivKey], equiv.total_equivalent_cost_usd)) {
    invalid(`${path}.${totalEquivKey}`);
  }
  if (record[provenanceKey] !== equiv.provenance) {
    invalid(`${path}.${provenanceKey}`);
  }

  validateEquivalentAgainstMeasured(equiv, measuredSummary, `${path}.${summaryKey}`);
}

function hasOwn(value: Record<string, unknown>, key: string): boolean {
  return Object.prototype.hasOwnProperty.call(value, key);
}

function requireRecord(value: unknown, path: string): Record<string, unknown> {
  if (!isRecord(value)) invalid(path);
  return value;
}

function validateAmountSummary(amount: unknown, summary: unknown, path: string): void {
  if (!isNullableAmount(amount)) invalid(path);
  if (summary !== undefined) validateSummary(summary, amount, `${path}.cost_summary`);
  else if (amount === null) invalid(`${path}.cost_summary`);
}

function validateCostMap(amounts: unknown, summaries: unknown, path: string): void {
  if (summaries !== undefined) {
    if (!isRecord(summaries) || !isRecord(amounts)) invalid(path);
    for (const key of Object.keys(summaries)) {
      if (!hasOwn(amounts, key)) invalid(`${path}.${key}`);
    }
  }
  if (amounts === undefined) return;
  if (!isRecord(amounts)) invalid(path);
  for (const [key, amount] of Object.entries(amounts)) {
    if (!isNullableAmount(amount)) invalid(`${path}.${key}`);
    if (summaries === undefined) {
      if (amount === null) invalid(`${path}.${key}`);
    } else {
      if (!isRecord(summaries) || !hasOwn(summaries, key)) invalid(`${path}.${key}`);
      validateSummary(summaries[key], amount, `${path}.${key}`);
    }
  }
}

function validateEquivalentCostMap(
  equivSummaries: unknown,
  amounts: unknown,
  measuredSummaries: unknown,
  path: string
): void {
  if (equivSummaries === undefined) return;
  if (!isRecord(equivSummaries) || !isRecord(amounts)) invalid(path);
  for (const [key, entry] of Object.entries(equivSummaries)) {
    if (!hasOwn(amounts, key)) invalid(`${path}.${key}`);
    const equiv = validateEquivalentCostSummary(entry, `${path}.${key}`);
    if (isRecord(measuredSummaries) && hasOwn(measuredSummaries, key)) {
      validateEquivalentAgainstMeasured(equiv, measuredSummaries[key], `${path}.${key}`);
    }
  }
  for (const key of Object.keys(amounts)) {
    if (!hasOwn(equivSummaries, key)) invalid(`${path}.${key}`);
  }
}

function validateFactCost(value: unknown, path: string): void {
  const fact = requireRecord(value, path);
  if (!hasOwn(fact, 'cost_usd') || !hasOwn(fact, 'cost_evidence')) invalid(path);
  if (hasOwn(fact, 'cost_usd') && !isNullableAmount(fact.cost_usd)) invalid(`${path}.cost_usd`);
  const evidence = validateEvidence(fact.cost_evidence, fact.cost_usd, `${path}.cost_evidence`);
  if (hasOwn(fact, 'cost_estimate') && fact.cost_estimate !== undefined) {
    validateCostEstimate(fact.cost_estimate, evidence, `${path}.cost_estimate`);
  }
}

function validateFactLedger(value: unknown, path: string): void {
  const facts = requireRecord(value, path);
  const version = facts.schema_version;
  if (version !== 6) throw new Error(`Migration required: ${path} must use schema v6; migrate older ledgers explicitly.`);
  if (['invocations', 'tools', 'events', 'events_are_partial', 'skills', 'task_links', 'delegations', 'unresolved_facts'].some((key) => hasOwn(facts, key))) invalid(`${path}: retired root fact collections`);
  if (!Array.isArray(facts.sessions)) invalid(`${path}.sessions`);
  facts.sessions.forEach((value: unknown, index) => {
    const session = requireRecord(value, `${path}.sessions[${index}]`);
    if (!Array.isArray(session.invocations)) invalid(`${path}.sessions[${index}].invocations`);
    session.invocations.forEach((invocation: unknown, row) => validateFactCost(invocation, `${path}.sessions[${index}].invocations[${row}]`));
  });
}

function validateMetric(value: unknown, path: string): Record<string, unknown> {
  const metric = requireRecord(value, path);
  validateAmountSummary(metric.cost_usd, metric.cost_summary, path);
  validateEquivalentFields(metric, path, metric.cost_summary);
  return metric;
}

function validateModelMetric(value: unknown, path: string): void {
  const metric = validateMetric(value, path);
  if (hasOwn(metric, 'applied_rates') && metric.applied_rates !== undefined) {
    if (!Array.isArray(metric.applied_rates)) invalid(`${path}.applied_rates`);
    metric.applied_rates.forEach((rate: unknown, index) =>
      validateAppliedRateSnapshot(rate, `${path}.applied_rates[${index}]`)
    );
  }
}

function validateWorkflow(value: unknown, path: string): void {
  const workflow = requireRecord(value, path);
  if (workflow.role_effort_distribution_pct !== undefined) {
    const percentages = requireRecord(workflow.role_effort_distribution_pct, `${path}.role_effort_distribution_pct`);
    for (const [role, percentage] of Object.entries(percentages)) {
      if (percentage !== null && (!isAmount(percentage) || percentage > 100)) invalid(`${path}.role_effort_distribution_pct.${role}`);
    }
  }
}

export function validateLedgerCosts(value: unknown): void {
  const path = 'response.ledger';
  const ledger = requireRecord(value, path);
  validateFactLedger(ledger.facts, `${path}.facts`);
  validateMetric(ledger.totals, `${path}.totals`);
  validateWorkflow(ledger.workflow_efficiency, `${path}.workflow_efficiency`);
  if (ledger.roles != null) {
    if (!Array.isArray(ledger.roles)) invalid(`${path}.roles`);
    ledger.roles.forEach((role: unknown, index) => validateMetric(role, `${path}.roles[${index}]`));
  }
  if (ledger.models != null) {
    if (!Array.isArray(ledger.models)) invalid(`${path}.models`);
    ledger.models.forEach((model: unknown, index) => validateModelMetric(model, `${path}.models[${index}]`));
  }
  if (ledger.threads != null) {
    if (!Array.isArray(ledger.threads)) invalid(`${path}.threads`);
    ledger.threads.forEach((value: unknown, index) => {
      const threadPath = `${path}.threads[${index}]`;
      const thread = requireRecord(value, threadPath);
      if (!Array.isArray(thread.roles)) invalid(`${threadPath}.roles`);
      thread.roles.forEach((value: unknown, roleIndex) => {
        const rolePath = `${threadPath}.roles[${roleIndex}]`;
        const role = requireRecord(value, rolePath);
        if (!Array.isArray(role.models)) invalid(`${rolePath}.models`);
        role.models.forEach((model: unknown, modelIndex) => validateModelMetric(model, `${rolePath}.models[${modelIndex}]`));
      });
    });
  }
}

export function validateFleetCosts(value: unknown): void {
  const fleet = requireRecord(value, 'response');
  validateAmountSummary(fleet.total_cost_usd, fleet.cost_summary, 'response.total_cost_usd');
  validateEquivalentFields(fleet, 'response', fleet.cost_summary);
  validateCostMap(fleet.role_costs, fleet.role_cost_summaries, 'response.role_costs');
  validateCostMap(fleet.model_costs, fleet.model_cost_summaries, 'response.model_costs');
  validateEquivalentCostMap(
    fleet.role_equivalent_cost_summaries,
    fleet.role_costs,
    fleet.role_cost_summaries,
    'response.role_equivalent_cost_summaries'
  );
  validateEquivalentCostMap(
    fleet.model_equivalent_cost_summaries,
    fleet.model_costs,
    fleet.model_cost_summaries,
    'response.model_equivalent_cost_summaries'
  );
}

export function validateTimeBucketCosts(value: unknown, path = 'response'): void {
  const bucket = requireRecord(value, path);
  validateAmountSummary(bucket.cost_usd, bucket.cost_summary, path);
  validateEquivalentFields(bucket, path, bucket.cost_summary);
  validateAmountSummary(bucket.cumulative_cost_usd, bucket.cumulative_cost_summary, `${path}.cumulative_cost_usd`);
  validateEquivalentFields(bucket, path, bucket.cumulative_cost_summary, 'cumulative_');
}

export function validateComparisonRowCosts(value: unknown, path = 'response'): void {
  validateMetric(value, path);
}
