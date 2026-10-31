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

function validateFactCost(value: unknown, requireEvidence: boolean, path: string): void {
  const fact = requireRecord(value, path);
  if (requireEvidence && (!hasOwn(fact, 'cost_usd') || !hasOwn(fact, 'cost_evidence'))) invalid(path);
  if (hasOwn(fact, 'cost_usd') && !isNullableAmount(fact.cost_usd)) invalid(`${path}.cost_usd`);
  if (hasOwn(fact, 'cost_evidence')) validateEvidence(fact.cost_evidence, fact.cost_usd, `${path}.cost_evidence`);
}

function validateFactLedger(value: unknown, path: string): void {
  const facts = requireRecord(value, path);
  const version = facts.schema_version;
  if (version !== undefined && version !== 4 && version !== 5) invalid(`${path}.schema_version`);
  const strictFacts = version === 5;
  if (!Array.isArray(facts.sessions)) invalid(`${path}.sessions`);
  facts.sessions.forEach((value: unknown, index) => {
    const session = requireRecord(value, `${path}.sessions[${index}]`);
    if (!Array.isArray(session.invocations)) invalid(`${path}.sessions[${index}].invocations`);
    session.invocations.forEach((invocation: unknown, row) => validateFactCost(invocation, strictFacts, `${path}.sessions[${index}].invocations[${row}]`));
  });
  if (facts.task_links != null) {
    if (!Array.isArray(facts.task_links)) invalid(`${path}.task_links`);
    facts.task_links.forEach((task: unknown, index) => validateFactCost(task, strictFacts, `${path}.task_links[${index}]`));
  }
}

function validateMetric(value: unknown, path: string): void {
  const metric = requireRecord(value, path);
  validateAmountSummary(metric.cost_usd, metric.cost_summary, path);
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
    ledger.models.forEach((model: unknown, index) => validateMetric(model, `${path}.models[${index}]`));
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
        role.models.forEach((model: unknown, modelIndex) => validateMetric(model, `${rolePath}.models[${modelIndex}]`));
      });
    });
  }
}

export function validateFleetCosts(value: unknown): void {
  const fleet = requireRecord(value, 'response');
  validateAmountSummary(fleet.total_cost_usd, fleet.cost_summary, 'response.total_cost_usd');
  validateCostMap(fleet.role_costs, fleet.role_cost_summaries, 'response.role_costs');
  validateCostMap(fleet.model_costs, fleet.model_cost_summaries, 'response.model_costs');
}
