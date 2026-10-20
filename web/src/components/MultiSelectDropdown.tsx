import React, { useState, useRef, useEffect, useCallback, useId } from 'react';

export interface MultiSelectOption {
  readonly value: string;
  readonly label: string;
}

export interface MultiSelectDropdownProps {
  readonly id?: string;
  readonly label: string;
  readonly options: readonly MultiSelectOption[];
  readonly selectedValues: readonly string[];
  readonly onChange: (selected: readonly string[]) => void;
  readonly placeholder?: string;
  readonly allLabel?: string;
  readonly ariaLabel?: string;
  readonly testId?: string;
  readonly className?: string;
}

export const MultiSelectDropdown: React.FC<MultiSelectDropdownProps> = ({
  id: customId,
  label,
  options,
  selectedValues,
  onChange,
  placeholder,
  allLabel,
  ariaLabel,
  testId,
  className,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);
  const autoId = useId();
  const id = customId ?? autoId;

  const resolvedAllLabel = allLabel ?? 'All';
  const resolvedPlaceholder = placeholder ?? (allLabel ?? 'All');
  const resolvedAriaLabel = ariaLabel ?? `${label} selector`;
  const resolvedTestId = testId ?? (customId ?? autoId);
  const allCheckboxAriaLabel = allLabel !== undefined ? allLabel : `Select all ${label}`;

  const isAllSelected =
    selectedValues.length === 0 ||
    selectedValues.includes('all') ||
    selectedValues.includes('ALL');

  // Handle clicking outside to close
  useEffect(() => {
    function handleClickOutside(event: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    }
    if (isOpen) {
      document.addEventListener('mousedown', handleClickOutside);
    }
    return () => {
      document.removeEventListener('mousedown', handleClickOutside);
    };
  }, [isOpen]);

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent) => {
      if (e.key === 'Escape') {
        setIsOpen(false);
      }
    },
    []
  );

  const toggleOpen = useCallback(() => {
    setIsOpen((prev) => !prev);
  }, []);

  const handleSelectAll = useCallback(() => {
    onChange(['all']);
  }, [onChange]);

  const handleToggleOption = useCallback(
    (value: string) => {
      if (isAllSelected) {
        onChange([value]);
        return;
      }

      const isCurrentSelected = selectedValues.includes(value);
      if (isCurrentSelected) {
        const next = selectedValues.filter((v) => v !== value);
        if (next.length === 0) {
          onChange(['all']);
        } else {
          onChange(next);
        }
      } else {
        const next = [...selectedValues, value];
        // If all available options are now selected, default back to 'all'
        if (options.length > 0 && next.length === options.length) {
          onChange(['all']);
        } else {
          onChange(next);
        }
      }
    },
    [isAllSelected, selectedValues, onChange, options.length]
  );

  const displaySummary = (): string => {
    if (isAllSelected) {
      return resolvedPlaceholder;
    }
    if (selectedValues.length === 1) {
      const match = options.find((opt) => opt.value === selectedValues[0]);
      return match ? match.label : selectedValues[0]!;
    }
    if (selectedValues.length === 2) {
      const labels = selectedValues.map((v) => {
        const match = options.find((opt) => opt.value === v);
        return match ? match.label : v;
      });
      return labels.join(', ');
    }
    return `${selectedValues.length} selected`;
  };

  const containerClassName = className
    ? `multiselect-container ${className}`
    : 'multiselect-container';

  return (
    <div className={containerClassName} ref={containerRef} onKeyDown={handleKeyDown}>
      <label htmlFor={`${id}-trigger`} className="multiselect-label">
        {label}:
      </label>
      <div className="multiselect-control">
        <button
          type="button"
          role="combobox"
          id={`${id}-trigger`}
          data-testid={resolvedTestId}
          className="multiselect-trigger"
          onClick={toggleOpen}
          aria-haspopup="listbox"
          aria-expanded={isOpen}
          aria-label={resolvedAriaLabel}
        >
          <span className="multiselect-summary">{displaySummary()}</span>
          <span className="multiselect-arrow" aria-hidden="true">
            {isOpen ? '^' : 'v'}
          </span>
        </button>

        {isOpen && (
          <div
            className="multiselect-dropdown"
            role="listbox"
            aria-multiselectable="true"
            aria-label={label}
          >
            <label className="multiselect-option all-option">
              <input
                type="checkbox"
                checked={isAllSelected}
                onChange={handleSelectAll}
                aria-label={allCheckboxAriaLabel}
              />
              <span>{resolvedAllLabel}</span>
            </label>
            <div className="multiselect-divider" />
            {options.map((opt) => {
              const checked = !isAllSelected && selectedValues.includes(opt.value);
              return (
                <label key={opt.value} className="multiselect-option">
                  <input
                    type="checkbox"
                    checked={checked}
                    onChange={() => handleToggleOption(opt.value)}
                    aria-label={opt.label}
                  />
                  <span>{opt.label}</span>
                </label>
              );
            })}
          </div>
        )}
      </div>
    </div>
  );
};
