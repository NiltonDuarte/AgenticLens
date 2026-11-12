import { useState } from 'react';
import { describe, expect, it, vi } from 'vitest';
import userEvent from '@testing-library/user-event';
import { fireEvent, render, screen, within } from '../test/render';
import { TagFilters } from './TagFilters';
import type { TagChoice, TagFilter } from '../domain/tags';

const choices: readonly TagChoice[] = [
  { key: 'team', values: ['core', 'Missing'] },
  { key: 'environment', values: ['dev', 'prod'] },
];

function Harness({ available = choices, initial = [], changed = () => undefined }: {
  readonly available?: readonly TagChoice[];
  readonly initial?: readonly TagFilter[];
  readonly changed?: (filters: readonly TagFilter[]) => void;
}) {
  const [filters, setFilters] = useState(initial);
  return <><TagFilters choices={available} filters={filters} onChange={(next) => { setFilters(next); changed(next); }} /><button type="button">Outside control</button></>;
}

describe('compact ticket tag filter', () => {
  it('lists active clauses and removes a nonedited key independently before clearing all', async () => {
    const user = userEvent.setup();
    const changed = vi.fn();
    render(<Harness changed={changed} />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    expect(trigger).toHaveTextContent('All');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(screen.queryByRole('checkbox')).not.toBeInTheDocument();
    await user.click(trigger);
    const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
    const key = within(editor).getByRole('combobox', { name: 'Tag key' });
    expect(key).toHaveFocus();
    expect(changed).not.toHaveBeenCalled();
    await user.click(within(editor).getByRole('checkbox', { name: 'Value: core' }));
    await user.click(within(editor).getByRole('checkbox', { name: 'Value: Missing' }));
    await user.click(within(editor).getByRole('checkbox', { name: 'Missing (key absent)' }));
    expect(changed).toHaveBeenLastCalledWith([{ key: 'team', values: ['core', 'Missing'], missing: true }]);
    await user.selectOptions(key, 'environment');
    await user.click(within(editor).getByRole('checkbox', { name: 'Value: prod' }));
    expect(changed).toHaveBeenLastCalledWith([
      { key: 'team', values: ['core', 'Missing'], missing: true },
      { key: 'environment', values: ['prod'], missing: false },
    ]);
    expect(trigger).toHaveTextContent('2 keys');
    const active = within(editor).getByRole('list', { name: 'Active ticket filters' });
    expect(within(active).getAllByRole('listitem')).toHaveLength(2);
    expect(active).toHaveTextContent('team');
    expect(active).toHaveTextContent('Value: core or Value: Missing or Missing (key absent)');
    expect(active).toHaveTextContent('environment');
    expect(active).toHaveTextContent('Value: prod');
    expect(key).toHaveValue('environment');
    expect(within(editor).queryByRole('button', { name: 'Clear key' })).not.toBeInTheDocument();
    expect(within(editor).queryByRole('button', { name: 'Clear tag filters' })).not.toBeInTheDocument();
    await user.click(within(active).getByRole('button', { name: 'Remove filter for team' }));
    expect(changed).toHaveBeenLastCalledWith([{ key: 'environment', values: ['prod'], missing: false }]);
    expect(key).toHaveValue('environment');
    expect(within(editor).getByRole('checkbox', { name: 'Value: prod' })).toBeChecked();
    expect(within(active).getAllByRole('listitem')).toHaveLength(1);
    expect(editor.contains(document.activeElement)).toBe(true);
    await user.selectOptions(key, 'team');
    await user.click(within(editor).getByRole('checkbox', { name: 'Value: core' }));
    expect(within(active).getAllByRole('listitem')).toHaveLength(2);
    await user.click(within(editor).getByRole('button', { name: 'Clear all' }));
    expect(changed).toHaveBeenLastCalledWith([]);
    expect(trigger).toHaveTextContent('All');
    expect(within(editor).getByText('No active filters.')).toBeInTheDocument();
    expect(within(editor).queryByRole('button', { name: 'Clear all' })).not.toBeInTheDocument();
    expect(editor.contains(document.activeElement)).toBe(true);
    await user.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('dismisses with Escape or an outside control without trapping or stealing focus', async () => {
    const user = userEvent.setup();
    render(<Harness />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    trigger.focus();
    await user.keyboard('{Enter}');
    expect(screen.getByRole('combobox', { name: 'Tag key' })).toHaveFocus();
    await user.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
    await user.keyboard(' ');
    await user.click(screen.getByRole('button', { name: 'Outside control' }));
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Outside control' })).toHaveFocus();
    await user.keyboard('{Escape}');
    expect(screen.getByRole('button', { name: 'Outside control' })).toHaveFocus();
  });

  it('keeps stale selected keys and values editable when the catalog becomes empty', async () => {
    const user = userEvent.setup();
    const changed = vi.fn();
    render(<Harness available={[]} initial={[{ key: '__proto__', values: ['old=value'], missing: true }]} changed={changed} />);
    await user.click(screen.getByRole('button', { name: 'Ticket tags filter' }));
    expect(screen.getByRole('combobox', { name: 'Tag key' })).toHaveValue('__proto__');
    expect(screen.getByRole('checkbox', { name: 'Value: old=value' })).toBeChecked();
    const active = screen.getByRole('list', { name: 'Active ticket filters' });
    expect(active).toHaveTextContent('__proto__');
    expect(active).toHaveTextContent('Value: old=value or Missing (key absent)');
    const remove = within(active).getByRole('button', { name: 'Remove filter for __proto__' });
    remove.focus();
    await user.keyboard('{Enter}');
    expect(changed).toHaveBeenLastCalledWith([]);
    expect(screen.getByText('No ticket tags recorded.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Close ticket tags' })).toHaveFocus();
    await user.keyboard('{Escape}');
    expect(screen.getByRole('button', { name: 'Ticket tags filter' })).toHaveFocus();
  });

  it.each([
    { left: 6, top: 400, bottom: 436 },
    { left: 990, top: 700, bottom: 736 },
    { left: 100, top: 8, bottom: 44 },
  ])('keeps the editor within the viewport for trigger bounds $left/$top', async (bounds) => {
    const user = userEvent.setup();
    render(<Harness />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    vi.spyOn(trigger, 'getBoundingClientRect').mockReturnValue({ ...bounds, right: bounds.left + 104, width: 104, height: bounds.bottom - bounds.top, x: bounds.left, y: bounds.top, toJSON: () => ({}) });
    await user.click(trigger);
    const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
    const left = Number.parseFloat(editor.style.left);
    const top = Number.parseFloat(editor.style.top);
    const width = Number.parseFloat(editor.style.width);
    const height = Number.parseFloat(editor.style.maxHeight);
    expect(left).toBeGreaterThanOrEqual(16);
    expect(top).toBeGreaterThanOrEqual(16);
    expect(left + width).toBeLessThanOrEqual(window.innerWidth - 16);
    expect(top + height).toBeLessThanOrEqual(window.innerHeight - 16);
  });

  it('repositions an open editor for a 320px viewport while keeping long key and value text usable', async () => {
    const user = userEvent.setup();
    const originalWidth = Object.getOwnPropertyDescriptor(window, 'innerWidth')!;
    const originalHeight = Object.getOwnPropertyDescriptor(window, 'innerHeight')!;
    const longKey = 'long-ticket-tag-key-'.repeat(12);
    const longValue = 'long-tag-value-'.repeat(18);
    try {
      render(<Harness available={[{ key: longKey, values: [longValue] }]} />);
      const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
      vi.spyOn(trigger, 'getBoundingClientRect').mockReturnValue({ left: 290, right: 394, top: 480, bottom: 516, width: 104, height: 36, x: 290, y: 480, toJSON: () => ({}) });
      await user.click(trigger);
      const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
      expect(Number.parseFloat(editor.style.width)).toBeGreaterThan(288);
      Object.defineProperty(window, 'innerWidth', { configurable: true, value: 320 });
      Object.defineProperty(window, 'innerHeight', { configurable: true, value: 568 });
      fireEvent.resize(window);
      const left = Number.parseFloat(editor.style.left);
      const top = Number.parseFloat(editor.style.top);
      const width = Number.parseFloat(editor.style.width);
      const height = Number.parseFloat(editor.style.maxHeight);
      expect(left).toBeGreaterThanOrEqual(16);
      expect(width).toBeLessThanOrEqual(288);
      expect(left + width).toBeLessThanOrEqual(304);
      expect(top).toBeGreaterThanOrEqual(16);
      expect(top + height).toBeLessThanOrEqual(552);
      expect(within(editor).getByRole('combobox', { name: 'Tag key' })).toHaveValue(longKey);
      await user.click(within(editor).getByRole('checkbox', { name: `Value: ${longValue}` }));
      expect(within(editor).getByRole('checkbox', { name: `Value: ${longValue}` })).toBeChecked();
      await user.click(within(editor).getByRole('button', { name: 'Clear all' }));
      expect(trigger).toHaveTextContent('All');
    } finally {
      Object.defineProperty(window, 'innerWidth', originalWidth);
      Object.defineProperty(window, 'innerHeight', originalHeight);
    }
  });


  it('dismisses an editor whose anchor scrolls outside the viewport without moving focus to its hidden trigger', async () => {
    const user = userEvent.setup();
    render(<Harness />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    const measure = vi.spyOn(trigger, 'getBoundingClientRect');
    await user.click(trigger);
    expect(screen.getByRole('dialog')).toBeInTheDocument();
    measure.mockReturnValue({ left: 100, right: 204, top: 1000, bottom: 1036, width: 104, height: 36, x: 100, y: 1000, toJSON: () => ({}) });
    fireEvent.scroll(window);
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).not.toHaveFocus();
  });


  it('recovers focus when an incoming empty catalog removes the focused key selector', async () => {
    const user = userEvent.setup();
    const { rerender } = render(<Harness />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    await user.click(trigger);
    expect(screen.getByRole('combobox', { name: 'Tag key' })).toHaveFocus();
    rerender(<Harness available={[]} />);
    expect(screen.getByText('No ticket tags recorded.')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: 'Close ticket tags' })).toHaveFocus();
    await user.keyboard('{Escape}');
    expect(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('recovers keyboard dismissal after deselecting a focused stale value while another selection remains', async () => {
    const user = userEvent.setup();
    const changed = vi.fn();
    render(<Harness available={[{ key: 'new', values: ['present'] }]}
      initial={[{ key: 'old', values: ['obsolete', 'still-stale'], missing: false }]} changed={changed} />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    await user.click(trigger);
    const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
    await user.selectOptions(within(editor).getByRole('combobox', { name: 'Tag key' }), 'old');
    const stale = within(editor).getByRole('checkbox', { name: 'Value: obsolete' });
    expect(stale).toBeChecked();
    stale.focus();
    await user.keyboard(' ');
    expect(within(editor).queryByRole('checkbox', { name: 'Value: obsolete' })).not.toBeInTheDocument();
    expect(within(editor).getByRole('checkbox', { name: 'Value: still-stale' })).toBeChecked();
    expect(changed).toHaveBeenLastCalledWith([{ key: 'old', values: ['still-stale'], missing: false }]);
    expect.soft(editor.contains(document.activeElement)).toBe(true);
    await user.keyboard('{Escape}');
    expect.soft(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('recovers keyboard dismissal when a nonempty catalog refresh removes the focused value', async () => {
    const user = userEvent.setup();
    const changed = vi.fn();
    const { rerender } = render(<Harness available={[{ key: 'team', values: ['core', 'obsolete'] }]} changed={changed} />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    await user.click(trigger);
    const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
    const obsolete = within(editor).getByRole('checkbox', { name: 'Value: obsolete' });
    expect(obsolete).not.toBeChecked();
    obsolete.focus();
    expect(obsolete).toHaveFocus();
    rerender(<Harness available={[{ key: 'team', values: ['core'] }]} changed={changed} />);
    expect(within(editor).queryByRole('checkbox', { name: 'Value: obsolete' })).not.toBeInTheDocument();
    expect(within(editor).getByRole('checkbox', { name: 'Value: core' })).not.toBeChecked();
    expect(changed).not.toHaveBeenCalled();
    expect.soft(editor.contains(document.activeElement)).toBe(true);
    await user.keyboard('{Escape}');
    expect.soft(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

  it('dismisses with Escape after clicking nonfocusable text inside the open editor', async () => {
    const user = userEvent.setup();
    render(<Harness />);
    const trigger = screen.getByRole('button', { name: 'Ticket tags filter' });
    await user.click(trigger);
    const editor = screen.getByRole('dialog', { name: 'Ticket tag filters' });
    await user.click(within(editor).getByText('Ticket tags', { selector: 'strong' }));
    expect(editor).toBeInTheDocument();
    await user.keyboard('{Escape}');
    expect.soft(screen.queryByRole('dialog')).not.toBeInTheDocument();
    expect(trigger).toHaveFocus();
  });

});
