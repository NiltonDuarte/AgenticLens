import { describe, it, expect, vi } from 'vitest';
import { render, screen, fireEvent, within } from '@testing-library/react';
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
    const arrow = trigger.querySelector('.multiselect-arrow')!;
    expect(arrow).toBeInTheDocument();
    expect(arrow.getAttribute('aria-hidden')).toBe('true');
    expect(arrow.textContent).toBe('v');
    fireEvent.click(trigger);
    expect(trigger.getAttribute('aria-expanded')).toBe('true');
    expect(arrow.textContent).toBe('^');

    const listbox = screen.getByRole('listbox', { name: 'Ticket Type' });
    expect(listbox).toBeDefined();
    expect(screen.getByText('Feature')).toBeDefined();
    expect(screen.getByText('Bug')).toBeDefined();
    expect(screen.getByText('Docs')).toBeDefined();
    fireEvent.click(trigger);
    expect(arrow.textContent).toBe('v');
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

  it('renders allLabel as "No filter" in trigger summary and checkbox when provided', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="tool-select"
        testId="custom-tool-select"
        label="Tool"
        allLabel="No filter"
        ariaLabel="Filter by tool"
        className="custom-filter-group"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    // Trigger should have data-testid, aria-label, and display "No filter"
    const trigger = screen.getByTestId('custom-tool-select');
    expect(trigger).toHaveAttribute('aria-label', 'Filter by tool');
    expect(trigger.textContent).toContain('No filter');

    // Root container should include className
    const container = trigger.closest('.multiselect-container');
    expect(container).toHaveClass('custom-filter-group');

    // Open dropdown and inspect all-option checkbox
    fireEvent.click(trigger);
    const noFilterCheckbox = screen.getByRole('checkbox', { name: 'No filter' });
    expect(noFilterCheckbox).toBeDefined();
    expect(noFilterCheckbox).toBeChecked();
    expect(screen.getAllByText('No filter').length).toBe(2);
  });

  it('respects explicit placeholder over allLabel when both are provided', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="role-select"
        label="Role"
        allLabel="No filter"
        placeholder="Custom Placeholder"
        options={options}
        selectedValues={['all']}
        onChange={handleChange}
      />
    );

    const trigger = screen.getByRole('combobox', { name: 'Role selector' });
    expect(trigger.textContent).toContain('Custom Placeholder');
  });

  it('selects all using allLabel checkbox click when single option was selected', () => {
    const handleChange = vi.fn();
    render(
      <MultiSelectDropdown
        id="tool-select"
        label="Tool"
        allLabel="No filter"
        options={options}
        selectedValues={['feature']}
        onChange={handleChange}
      />
    );

    fireEvent.click(screen.getByRole('combobox', { name: 'Tool selector' }));
    const noFilterCheckbox = screen.getByRole('checkbox', { name: 'No filter' });
    expect(noFilterCheckbox).not.toBeChecked();
    fireEvent.click(noFilterCheckbox);

    expect(handleChange).toHaveBeenCalledWith(['all']);
  });

  describe('Adversarial & Boundary Edge Cases', () => {
    it('handles empty selectedValues array as all-selected state', () => {
      const handleChange = vi.fn();
      render(
        <MultiSelectDropdown
          id="role-select"
          label="Role"
          allLabel="No filter"
          options={options}
          selectedValues={[]}
          onChange={handleChange}
        />
      );

      const trigger = screen.getByRole('combobox', { name: 'Role selector' });
      expect(trigger.textContent).toContain('No filter');

      fireEvent.click(trigger);
      const noFilterCheckbox = screen.getByRole('checkbox', { name: 'No filter' });
      expect(noFilterCheckbox).toBeChecked();
    });

    it('handles uppercase ALL in selectedValues as all-selected state', () => {
      const handleChange = vi.fn();
      render(
        <MultiSelectDropdown
          id="role-select"
          label="Role"
          allLabel="No filter"
          options={options}
          selectedValues={['ALL']}
          onChange={handleChange}
        />
      );

      const trigger = screen.getByRole('combobox', { name: 'Role selector' });
      expect(trigger.textContent).toContain('No filter');

      fireEvent.click(trigger);
      const noFilterCheckbox = screen.getByRole('checkbox', { name: 'No filter' });
      expect(noFilterCheckbox).toBeChecked();
    });

    it('handles unknown or unmapped values in selectedValues gracefully', () => {
      const handleChange = vi.fn();
      const { rerender } = render(
        <MultiSelectDropdown
          id="tool-select"
          label="Tool"
          options={options}
          selectedValues={['custom-unmapped-tool']}
          onChange={handleChange}
        />
      );

      // Single unmapped value displays raw value string
      const trigger = screen.getByRole('combobox', { name: 'Tool selector' });
      expect(trigger.textContent).toContain('custom-unmapped-tool');

      // Two unmapped values display comma-separated raw value strings
      rerender(
        <MultiSelectDropdown
          id="tool-select"
          label="Tool"
          options={options}
          selectedValues={['custom-tool-1', 'custom-tool-2']}
          onChange={handleChange}
        />
      );
      expect(trigger.textContent).toContain('custom-tool-1, custom-tool-2');
    });

    it('handles empty options list without crashing', () => {
      const handleChange = vi.fn();
      render(
        <MultiSelectDropdown
          id="empty-select"
          label="Empty Filter"
          allLabel="No filter"
          options={[]}
          selectedValues={['all']}
          onChange={handleChange}
        />
      );

      const trigger = screen.getByRole('combobox', { name: 'Empty Filter selector' });
      expect(trigger.textContent).toContain('No filter');

      fireEvent.click(trigger);
      expect(trigger.getAttribute('aria-expanded')).toBe('true');

      const listbox = screen.getByRole('listbox', { name: 'Empty Filter' });
      expect(listbox).toBeInTheDocument();

      const checkboxes = within(listbox).getAllByRole('checkbox');
      expect(checkboxes.length).toBe(1);
      expect(checkboxes[0]).toHaveAttribute('aria-label', 'No filter');
    });

    it('reverts to all when user manually selects each available option one by one', () => {
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

      // Since all 3 options are now selected, it automatically reverts to ['all']
      expect(handleChange).toHaveBeenCalledWith(['all']);
    });

    it('renders accessibility attributes correctly on combobox and listbox', () => {
      const handleChange = vi.fn();
      render(
        <MultiSelectDropdown
          id="custom-id"
          testId="custom-testid"
          label="Accessible Filter"
          ariaLabel="Custom Accessible Aria Label"
          options={options}
          selectedValues={['feature']}
          onChange={handleChange}
        />
      );

      const trigger = screen.getByTestId('custom-testid');
      expect(trigger).toHaveAttribute('role', 'combobox');
      expect(trigger).toHaveAttribute('aria-haspopup', 'listbox');
      expect(trigger).toHaveAttribute('aria-expanded', 'false');
      expect(trigger).toHaveAttribute('aria-label', 'Custom Accessible Aria Label');

      fireEvent.click(trigger);
      expect(trigger).toHaveAttribute('aria-expanded', 'true');

      const listbox = screen.getByRole('listbox', { name: 'Accessible Filter' });
      expect(listbox).toHaveAttribute('aria-multiselectable', 'true');
    });

    it('closes on Escape key when focus is inside the dropdown list', () => {
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

      const featureCheckbox = screen.getByRole('checkbox', { name: 'Feature' });
      fireEvent.keyDown(featureCheckbox, { key: 'Escape' });
      expect(trigger.getAttribute('aria-expanded')).toBe('false');
    });

    it('does not close dropdown on non-Escape key events', () => {
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

      fireEvent.keyDown(trigger, { key: 'ArrowDown' });
      expect(trigger.getAttribute('aria-expanded')).toBe('true');

      fireEvent.keyDown(trigger, { key: 'Enter' });
      expect(trigger.getAttribute('aria-expanded')).toBe('true');
    });

    it('does not close dropdown when clicking inside the dropdown menu', () => {
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
      fireEvent.mouseDown(listbox);
      expect(trigger.getAttribute('aria-expanded')).toBe('true');
    });

    it('links label to trigger button via id and htmlFor', () => {
      const handleChange = vi.fn();
      const { container } = render(
        <MultiSelectDropdown
          id="linked-id"
          label="Linked Label"
          options={options}
          selectedValues={['all']}
          onChange={handleChange}
        />
      );

      const label = container.querySelector('label.multiselect-label')!;
      const trigger = screen.getByRole('combobox', { name: 'Linked Label selector' });

      expect(label.getAttribute('for')).toBe('linked-id-trigger');
      expect(trigger.getAttribute('id')).toBe('linked-id-trigger');
    });

    it('displays summary count for 4 or more selections', () => {
      const fourOptions: readonly MultiSelectOption[] = [
        ...options,
        { value: 'chore', label: 'Chore' },
        { value: 'perf', label: 'Perf' },
      ];
      const handleChange = vi.fn();
      render(
        <MultiSelectDropdown
          id="type-select"
          label="Ticket Type"
          options={fourOptions}
          selectedValues={['feature', 'bug', 'docs', 'chore']}
          onChange={handleChange}
        />
      );

      const trigger = screen.getByRole('combobox', { name: 'Ticket Type selector' });
      expect(trigger.textContent).toContain('4 selected');
    });
  });
});
