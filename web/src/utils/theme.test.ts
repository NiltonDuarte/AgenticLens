// @vitest-environment node
import { describe, it, expect } from 'vitest';
import { THEME_COLORS, STATUS_COLORS, CHART_COLORS, getHealthStatusColor } from './theme';

describe('Theme and Design Tokens', () => {
  it('defines valid hex colors across all theme tokens', () => {
    const hexRegex = /^#[0-9a-fA-F]{6}$/;

    for (const [key, color] of Object.entries(THEME_COLORS)) {
      expect(color, `THEME_COLORS.${key}`).toMatch(hexRegex);
    }

    for (const [key, color] of Object.entries(STATUS_COLORS)) {
      expect(color, `STATUS_COLORS.${key}`).toMatch(hexRegex);
    }

    for (const [key, color] of Object.entries(CHART_COLORS)) {
      expect(color, `CHART_COLORS.${key}`).toMatch(hexRegex);
    }
  });

  describe('getHealthStatusColor', () => {
    it('returns appropriate color for each status case-insensitively', () => {
      expect(getHealthStatusColor('HEALTHY')).toBe(STATUS_COLORS.HEALTHY);
      expect(getHealthStatusColor('healthy')).toBe(STATUS_COLORS.HEALTHY);
      expect(getHealthStatusColor('ELEVATED')).toBe(STATUS_COLORS.ELEVATED);
      expect(getHealthStatusColor('elevated')).toBe(STATUS_COLORS.ELEVATED);
      expect(getHealthStatusColor('BLOATED')).toBe(STATUS_COLORS.BLOATED);
      expect(getHealthStatusColor('bloated')).toBe(STATUS_COLORS.BLOATED);
      expect(getHealthStatusColor('CRITICAL')).toBe(STATUS_COLORS.CRITICAL);
      expect(getHealthStatusColor('critical')).toBe(STATUS_COLORS.CRITICAL);
    });

    it('returns fallback color for unknown or empty statuses', () => {
      expect(getHealthStatusColor('unknown')).toBe(STATUS_COLORS.UNKNOWN);
      expect(getHealthStatusColor('')).toBe(STATUS_COLORS.UNKNOWN);
      expect(getHealthStatusColor('INVALID_STATUS')).toBe(STATUS_COLORS.UNKNOWN);
    });
  });
});
