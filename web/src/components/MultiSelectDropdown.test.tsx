import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent } from '@testing-library/react';
import { MultiSelectDropdown, type MultiSelectOption } from './MultiSelectDropdown';

describe('MultiSelectDropdown Component', () => {
  const options: readonly MultiSelectOption[] = [
    { value: 'feature', label: 'Feature' },
    { value: 'bug', label: 'Bug' },
    { value: 'docs', label: 'Docs' },
  ];

  it('renders trigger button displaying placeholder when all is selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    expect(trigger).toBeDefined();
    expect(trigger.textContent).toContain('All');
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('opens dropdown on click and displays options', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(trigger);
    expect(trigger.getAttribute('aria-expanded')).toBe('true');

    const listbox = screen.getByRole('listbox', { name: 'Ticket Type' });
    expect(listbox).toBeDefined();
    expect(screen.getByText('Feature')).toBeDefined();
    expect(screen.getByText('Bug')).toBeDefined();
    expect(screen.getByText('Docs')).toBeDefined();
  });

  it('selects single option when initially all was selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const featureCheckbox = screen.getByRole('checkbox', { name: 'Feature' });
    fireEvent.click(featureCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['feature']);
  });

  it('toggles option and reverts to all if last item unchecked', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const featureCheckbox = screen.getByRole('checkbox', { name: 'Feature' });
    fireEvent.click(featureCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['all']);
  });

  it('adds option when another option is already selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['feature', 'bug']);
  });

  it('reverts to all when all options become selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature', 'bug']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const docsCheckbox = screen.getByRole('checkbox', { name: 'Docs' });
    fireEvent.click(docsCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['all']);
  });

  it('selects all when All checkbox is clicked', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const allCheckbox = screen.getByRole('checkbox', { name: 'Select all Ticket Type' });
    fireEvent.click(allCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['all']);
  });

  it('closes dropdown when Escape key is pressed', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(trigger);
    expect(trigger.getAttribute('aria-expanded')).toBe('true');

    fireEvent.keyDown(trigger, { key: 'Escape' });
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('displays summary text for two selections', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature', 'bug']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    expect(trigger.textContent).toContain('Feature, Bug');
  });

  it('displays summary count for three selections with larger list', () => {
    const extendedOptions: readonly MultiSelectOption[] = [
      ...options,
      { value: 'chore', label: 'Chore' },
      { value: 'refactor', label: 'Refactor' },
    ];
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={extendedOptions}
        selectedValues={['feature', 'bug', 'chore']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    expect(trigger.textContent).toContain('3 selected');
  });

  it('deselects single option when multiple are selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature', 'bug']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));
    const bugCheckbox = screen.getByRole('checkbox', { name: 'Bug' });
    fireEvent.click(bugCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['feature']);
  });

  it('closes dropdown when clicking outside', () => {
    const handleChange = vi.fn();
    render(
      <div>
        <div data-testid="outside-area">Outside</div>
        <MultiSelectDropdown
          id="type-select"
          label="Ticket Type"
          options={options}
          selectedValues={['all']}
          onChange={handleChange}
        />
      </div>
    );

    const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
    fireEvent.click(trigger);
    expect(trigger.getAttribute('aria-expanded')).toBe('true');

    // Click outside
    fireEvent.mouseDown(screen.getByTestId('outside-area'));
    expect(trigger.getAttribute('aria-expanded')).toBe('false');
  });

  it('contains zero emojis in the rendered component', () => {
    const handleChange = vi.fn();
    const { container } = render(
      <MultiSelectDropdown
        id="type-select"
        label="Ticket Type"
        options={options}
        selectedValues={['feature']}
        onChange={handleChange}
      />
    );

    // Open dropdown to check all rendered options
    fireEvent.click(screen.getByRole('combobox', { name: 'Ticket Type selector' }));

    const text = container.textContent ?? '';
    const emojiRegex = /[\u{1F300}-\u{1F9FF}\u{2600}-\u{26FF}\u{2700}-\u{27BF}\u{1F1E6}-\u{1F1FF}]/u;
    expect(emojiRegex.test(text)).toBe(false);
  });
});
