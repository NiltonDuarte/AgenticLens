// A missing provisional amount is not proof of completed source activity.
export function durationDescription(provisional: number | undefined): string {
  return provisional !== undefined && provisional > 0
    ? `Recorded duration including ${provisional}s provisional`
    : 'Recorded duration';
}

export function durationLabel(value: string, provisional: number | undefined): string {
  return provisional !== undefined && provisional > 0
    ? `${value} (includes ${provisional}s provisional)`
    : value;
}

export function validDuration(total: unknown, provisional: unknown): boolean {
  return typeof total === 'number' && Number.isSafeInteger(total) && total >= 0 &&
    (provisional === undefined || (typeof provisional === 'number' && Number.isSafeInteger(provisional) && provisional >= 0 && provisional <= total));
}
