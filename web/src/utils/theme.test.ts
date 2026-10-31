// @vitest-environment node
import { describe, it, expect } from 'vitest';
import {
  THEME_COLORS,
  STATUS_COLORS,
  CHART_COLORS,
  DYNAMIC_PALETTE,
  getHealthStatusColor,
  getRoleColor,
  getModelColor,
  hashStringToColor,
} from './theme';

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

    for (const color of DYNAMIC_PALETTE) {
      expect(color, 'DYNAMIC_PALETTE item').toMatch(hexRegex);
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

  describe('Dynamic Role and Model Color Hashing', () => {
    it('handles roles case-insensitively with whitespace trimming', () => {
      expect(getRoleColor('  ORCHESTRATOR  ')).toBe(getRoleColor('orchestrator'));
      expect(getRoleColor('Coder')).toBe(getRoleColor('coder'));
    });

    it('returns deterministic hex colors for custom roles', () => {
      const hexRegex = /^#[0-9a-fA-F]{6}$/;
      const color1 = getRoleColor('advisor');
      const color2 = getRoleColor('advisor');
      expect(color1).toMatch(hexRegex);
      expect(color1).toBe(color2);

      const poColor = getRoleColor('po');
      expect(poColor).toMatch(hexRegex);
    });

    it('falls back to slate for empty or whitespace strings', () => {
      expect(hashStringToColor('')).toBe(THEME_COLORS.slate);
      expect(hashStringToColor('   ')).toBe(THEME_COLORS.slate);
    });

    it('handles special characters and complex model tags deterministically', () => {
      const hexRegex = /^#[0-9a-fA-F]{6}$/;
      const complexStrings = [
        'claude-3-7-sonnet',
        'security/auditor_v2',
        'meta-llama/llama-3.3-70b-instruct:free',
        'deepseek-ai/deepseek-r1-distill-qwen-32b',
        'po-analyst@workspace-1',
        'custom:::model$$$',
      ];

      for (const s of complexStrings) {
        const color = getModelColor(s);
        expect(color).toMatch(hexRegex);
        expect(getModelColor(s)).toBe(color);
        expect(DYNAMIC_PALETTE).toContain(color);
      }
    });

  });
});
