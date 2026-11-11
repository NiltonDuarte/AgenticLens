/** Shared literal ticket grammar; recorded identifiers are never rewritten. */
export interface TicketRecognitionPolicy {
  readonly prefixes: readonly string[];
  readonly caseInsensitive: boolean;
}

/** Unicode code point ordering agrees with Go string ordering for valid UTF-8. */
export function compareText(left: string, right: string): number {
  const a = Array.from(left, (char) => char.codePointAt(0)!);
  const b = Array.from(right, (char) => char.codePointAt(0)!);
  for (let i = 0; i < Math.min(a.length, b.length); i++) {
    if (a[i] !== b[i]) return a[i]! < b[i]! ? -1 : 1;
  }
  return a.length === b.length ? 0 : a.length < b.length ? -1 : 1;
}

/** Match Go's per-code-point Unicode lowercase operation without contextual folds. */
export function lowercaseTicketText(value: string): string {
  return Array.from(value, (char) => Array.from(char.toLowerCase())[0]!).join('');
}

export function ticketNumber(id: string, prefix: string, caseInsensitive: boolean): string | null {
  const selected = prefix.trim();
  if (!selected) return null;
  const candidate = caseInsensitive ? lowercaseTicketText(id) : id;
  const literal = caseInsensitive ? lowercaseTicketText(selected) : selected;
  if (!candidate.startsWith(literal)) return null;
  const suffix = candidate.slice(literal.length);
  return /^[-_]?[0-9]+$/.test(suffix) ? suffix.replace(/^[-_]/, '') : null;
}

function recognizedTicket(id: string, policy: TicketRecognitionPolicy): {prefix: string; number: string} | null {
  const prefixes = [...policy.prefixes].map((prefix) => prefix.trim())
    .sort((a, b) => b.length - a.length || compareText(a, b));
  for (const prefix of prefixes) {
    const number = ticketNumber(id, prefix, policy.caseInsensitive);
    if (number !== null) return {prefix: policy.caseInsensitive ? lowercaseTicketText(prefix) : prefix, number};
  }
  return null;
}

export function compareConfiguredTicketIds(a: string, b: string, policy?: TicketRecognitionPolicy): number {
  if (a === b) return 0;
  const left = policy ? recognizedTicket(a, policy) : null;
  const right = policy ? recognizedTicket(b, policy) : null;
  if (left && right) {
    const prefixOrder = compareText(left.prefix, right.prefix);
    if (prefixOrder !== 0) return prefixOrder;
    const numberA = left.number.replace(/^0+/, '') || '0';
    const numberB = right.number.replace(/^0+/, '') || '0';
    if (numberA.length !== numberB.length) return numberA.length < numberB.length ? -1 : 1;
    const numericOrder = compareText(numberA, numberB);
    if (numericOrder !== 0) return numericOrder;
  } else if (left || right) {
    return left ? -1 : 1;
  }
  return compareText(a, b);
}
