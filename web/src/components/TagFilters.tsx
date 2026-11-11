import React from 'react';
import { MultiSelectDropdown } from './MultiSelectDropdown';
import type { TagChoice, TagFilter } from '../domain/tags';

interface TagFiltersProps {
  readonly choices: readonly TagChoice[];
  readonly filters: readonly TagFilter[];
  readonly onChange: (filters: readonly TagFilter[]) => void;
}
export const TagFilters: React.FC<TagFiltersProps> = ({ choices, filters, onChange }) => (
  <fieldset className="tag-filters">
    <legend>Ticket tags</legend>
    {choices.length === 0 && <span>No ticket tags recorded.</span>}
    {choices.map(({ key, values }) => {
      const selected = filters.find((filter) => filter.key === key);
      return <MultiSelectDropdown
        key={key}
        label={`Tag: ${key}`}
        options={[{ value: 'missing', label: 'Missing (key absent)' }, ...values.map((value) => ({ value: JSON.stringify(['value', value]), label: `Value: ${value}` }))]}
        selectedValues={selected ? [...(selected.missing ? ['missing'] : []), ...selected.values.map((value) => JSON.stringify(['value', value]))] : []}
        collapseCompleteSelection={false}
        onChange={(tokens) => {
          const remaining = filters.filter((filter) => filter.key !== key);
          if (tokens.includes('all') || tokens.length === 0) { onChange(remaining); return; }
          onChange([...remaining, { key, missing: tokens.includes('missing'), values: values.filter((value) => tokens.includes(JSON.stringify(['value', value]))) }]);
        }}
      />;
    })}
    <button type="button" disabled={filters.length === 0} onClick={() => onChange([])}>Clear tag filters</button>
  </fieldset>
);
