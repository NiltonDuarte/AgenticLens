import { compareText } from './tickets';

export function compareRoles(left: string, right: string, order: readonly string[]): number {
  const leftIndex = order.indexOf(left);
  const rightIndex = order.indexOf(right);
  if (leftIndex >= 0 && rightIndex >= 0) return leftIndex - rightIndex;
  if (leftIndex >= 0 || rightIndex >= 0) return leftIndex >= 0 ? -1 : 1;
  return compareText(left, right);
}

export function projectedRole(role: string, mappings: Readonly<Record<string, string>>): string {
  return Object.hasOwn(mappings, role) ? mappings[role]! : role;
}
