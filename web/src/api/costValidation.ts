function isRecord(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === 'object' && !Array.isArray(value);
}

function isAmount(value: unknown): value is number {
  return typeof value === 'number' && Number.isFinite(value) && value >= 0;
}

function isNullableAmount(value: unknown): boolean {
  return value === null || isAmount(value);
}

function invalid(path: string): never {
  throw new Error(`Invalid cost payload at ${path}`);
}

function validateEvidence(value: unknown, amount: unknown, path: string): void {
  if (!isRecord(value)) invalid(path);
  const subtotal = value.known_subtotal_usd;
  if (value.state === 'observed') {
    if (!isAmount(amount) || subtotal !== amount) invalid(path);
  } else if (value.state === 'partial') {
    if (amount !== null || !isAmount(subtotal)) invalid(path);
  } else if (value.state === 'unavailable') {
    if (amount !== null || subtotal !== null) invalid(path);
  } else invalid(path);
}

function validateSummary(value: unknown, amount: unknown, path: string): void {
  if (!isRecord(value) || !isNullableAmount(amount) || !isNullableAmount(value.known_subtotal_usd)) invalid(path);
  for (const key of ['fully_observed', 'partial', 'unavailable', 'total']) {
    if (typeof value[key] !== 'number' || !Number.isSafeInteger(value[key]) || value[key] < 0) invalid(path);
  }
  const observed = value.fully_observed as number;
  const partial = value.partial as number;
  const unavailable = value.unavailable as number;
  const total = value.total as number;
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

/** Validate changed financial fields before existing API-specific projections. */
export function validateCostPayload(value: unknown, path = 'response', requireEvidence = false): void {
  if (Array.isArray(value)) {
    value.forEach((entry: unknown, index) => validateCostPayload(entry, `${path}[${index}]`, requireEvidence));
    return;
  }
  if (!isRecord(value)) return;
  const strictFacts = requireEvidence || value.schema_version === 5;
  for (const field of ['cost_usd', 'total_cost_usd', 'cumulative_cost_usd']) {
    if (field in value && !isNullableAmount(value[field])) invalid(`${path}.${field}`);
  }
  if ('cost_evidence' in value) validateEvidence(value.cost_evidence, value.cost_usd, `${path}.cost_evidence`);
  if ('cost_summary' in value) validateSummary(value.cost_summary, value.total_cost_usd ?? value.cost_usd ?? null, `${path}.cost_summary`);
  if ('cumulative_cost_summary' in value) validateSummary(value.cumulative_cost_summary, value.cumulative_cost_usd, `${path}.cumulative_cost_summary`);
  if ((value.total_cost_usd === null || (value.cost_usd === null && !('tokens' in value) && !('subagent_type' in value))) && !('cost_summary' in value)) invalid(`${path}.cost_summary`);
  if (value.cumulative_cost_usd === null && !('cumulative_cost_summary' in value)) invalid(`${path}.cumulative_cost_summary`);
  if (strictFacts && 'tokens' in value && 'model' in value && !('cost_evidence' in value)) invalid(`${path}.cost_evidence`);
  for (const [amountMap, summaryMap] of [['role_costs', 'role_cost_summaries'], ['model_costs', 'model_cost_summaries']] as const) {
    const amounts = value[amountMap];
    const summaries = value[summaryMap];
    if (summaries !== undefined) {
      if (!isRecord(summaries) || !isRecord(amounts)) invalid(`${path}.${summaryMap}`);
      for (const key of Object.keys(summaries)) {
        if (!Object.prototype.hasOwnProperty.call(amounts, key)) invalid(`${path}.${summaryMap}.${key}`);
      }
    }
    if (amounts !== undefined) {
      if (!isRecord(amounts)) invalid(`${path}.${amountMap}`);
      for (const [key, amount] of Object.entries(amounts)) {
        if (!isNullableAmount(amount)) invalid(`${path}.${amountMap}.${key}`);
        if (amount === null && summaries === undefined) invalid(`${path}.${summaryMap}.${key}`);
        if (summaries !== undefined) {
          if (!isRecord(summaries)) invalid(`${path}.${summaryMap}`);
          validateSummary(summaries[key], amount, `${path}.${summaryMap}.${key}`);
        }
      }
    }
  }
  for (const [key, entry] of Object.entries(value)) validateCostPayload(entry, `${path}.${key}`, strictFacts);
}

export async function readCostResponse<T>(response: Response): Promise<T> {
  const value: unknown = await response.json();
  validateCostPayload(value);
  return value as T;
}
