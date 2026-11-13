import React, { useCallback, useLayoutEffect, useId, useRef, useState } from 'react';
import { createPortal } from 'react-dom';
import type { TagChoice, TagFilter } from '../domain/tags';

interface TagFiltersProps {
  readonly choices: readonly TagChoice[];
  readonly filters: readonly TagFilter[];
  readonly onChange: (filters: readonly TagFilter[]) => void;
}

export const TagFilters: React.FC<TagFiltersProps> = ({ choices, filters, onChange }) => {
  const [isOpen, setIsOpen] = useState(false);
  const [selectedKey, setSelectedKey] = useState('');
  const [position, setPosition] = useState<React.CSSProperties>({});
  const triggerRef = useRef<HTMLButtonElement>(null);
  const editorRef = useRef<HTMLDivElement>(null);
  const keyRef = useRef<HTMLSelectElement>(null);
  const closeRef = useRef<HTMLButtonElement>(null);
  const wasOpenRef = useRef(false);
  const focusedEditorElementRef = useRef<HTMLElement | null>(null);
  const id = useId();
  const catalog = new Map(choices.map((choice) => [choice.key, [...choice.values]]));
  for (const filter of filters) {
    catalog.set(filter.key, [...new Set([...(catalog.get(filter.key) ?? []), ...filter.values])]);
  }
  const keys = [...catalog.keys()];
  const key = catalog.has(selectedKey) ? selectedKey : keys[0] ?? '';
  const values = catalog.get(key) ?? [];
  const selected = filters.find((filter) => filter.key === key);

  const placeEditor = useCallback(() => {
    const bounds = triggerRef.current?.getBoundingClientRect();
    if (!bounds) return;
    if (bounds.bottom < 0 || bounds.top > window.innerHeight || bounds.right < 0 || bounds.left > window.innerWidth) {
      setIsOpen(false);
      return;
    }
    const width = Math.min(352, Math.max(0, window.innerWidth - 32));
    const below = Math.max(0, window.innerHeight - bounds.bottom - 20);
    const above = Math.max(0, bounds.top - 20);
    const placeBelow = below >= Math.min(260, above);
    const maxHeight = Math.min(420, placeBelow ? below : above, Math.max(0, window.innerHeight - 32));
    const top = Math.max(16, Math.min(placeBelow ? bounds.bottom + 4 : bounds.top - maxHeight - 4, window.innerHeight - maxHeight - 16));
    setPosition({ width, maxHeight, top, left: Math.max(16, Math.min(bounds.left, window.innerWidth - width - 16)) });
  }, []);

  const closeEditor = useCallback(() => {
    setIsOpen(false);
    triggerRef.current?.focus();
  }, []);

  useLayoutEffect(() => {
    if (!isOpen) return;
    placeEditor();
    const dismissOutside = (event: PointerEvent | FocusEvent) => {
      const target = event.target;
      if (target instanceof Node && !editorRef.current?.contains(target) && !triggerRef.current?.contains(target)) {
        setIsOpen(false);
      }
    };
    const dismissOnEscape = (event: KeyboardEvent) => {
      if (event.key === 'Escape' && !event.defaultPrevented) {
        event.preventDefault();
        event.stopPropagation();
        closeEditor();
      }
    };
    document.addEventListener('keydown', dismissOnEscape);
    document.addEventListener('pointerdown', dismissOutside);
    document.addEventListener('focusin', dismissOutside);
    window.addEventListener('resize', placeEditor);
    window.addEventListener('scroll', placeEditor, true);
    return () => {
      document.removeEventListener('keydown', dismissOnEscape);
      document.removeEventListener('pointerdown', dismissOutside);
      document.removeEventListener('focusin', dismissOutside);
      window.removeEventListener('resize', placeEditor);
      window.removeEventListener('scroll', placeEditor, true);
    };
  }, [isOpen, placeEditor, closeEditor]);

  useLayoutEffect(() => {
    const justOpened = isOpen && !wasOpenRef.current;
    wasOpenRef.current = isOpen;
    if (!isOpen) {
      focusedEditorElementRef.current = null;
      return;
    }
    const previous = focusedEditorElementRef.current;
    const focusedElementWasRemoved = previous !== null && !editorRef.current?.contains(previous);
    const active = document.activeElement;
    const focusFellToDocument = active === null || active === document.body || active === document.documentElement;
    if (justOpened || (focusedElementWasRemoved && focusFellToDocument)) {
      (keyRef.current ?? closeRef.current)?.focus();
    }
  });

  const changeKeyFilter = (next: TagFilter | undefined) => {
    const others = filters.filter((filter) => filter.key !== key);
    onChange(next && (next.values.length > 0 || next.missing) ? [...others, next] : others);
  };
  const toggleValue = (value: string) => {
    const next = selected?.values.includes(value) ? selected.values.filter((item) => item !== value) : [...(selected?.values ?? []), value];
    changeKeyFilter({ key, values: next, missing: selected?.missing ?? false });
  };

  return (
    <div className="multiselect-container tag-filters">
      <label htmlFor={`${id}-trigger`} className="multiselect-label">Ticket tags:</label>
      <button ref={triggerRef} id={`${id}-trigger`} type="button" className="multiselect-trigger"
        aria-label="Ticket tags filter" aria-haspopup="dialog" aria-expanded={isOpen} aria-controls={isOpen ? `${id}-editor` : undefined}
        onClick={() => setIsOpen((open) => !open)}>
        <span className="multiselect-summary">{filters.length === 0 ? 'All' : `${filters.length} ${filters.length === 1 ? 'key' : 'keys'}`}</span>
        <span className="multiselect-arrow" aria-hidden="true">{isOpen ? '^' : 'v'}</span>
      </button>
      {isOpen && createPortal(
        <div ref={editorRef} id={`${id}-editor`} className="tag-filters-editor" role="dialog" aria-label="Ticket tag filters" style={position}
          onFocusCapture={(event) => { if (event.target instanceof HTMLElement) focusedEditorElementRef.current = event.target; }}>
          <div className="tag-filters-heading"><strong>Ticket tags</strong><button ref={closeRef} type="button" onClick={closeEditor} aria-label="Close ticket tags">Close</button></div>
          <section className="tag-filters-active">
            <div className="tag-filters-active-heading">
              <strong>Active filters</strong>
              {filters.length > 0 && <button type="button" className="tag-filters-clear" onClick={() => onChange([])}>Clear all</button>}
            </div>
            {filters.length === 0 ? <p className="tag-filters-none">No active filters.</p> : <ul className="tag-filters-active-list" aria-label="Active ticket filters">
              {filters.map((filter) => <li key={filter.key}>
                <span className="tag-filters-active-label">
                  <strong>{filter.key}</strong>
                  <span>{[...filter.values.map((value) => `Value: ${value}`), ...(filter.missing ? ['Missing (key absent)'] : [])].join(' or ')}</span>
                </span>
                <button type="button" className="tag-filters-remove" aria-label={`Remove filter for ${filter.key}`}
                  onClick={() => onChange(filters.filter((item) => item.key !== filter.key))}><span aria-hidden="true">X</span></button>
              </li>)}
            </ul>}
          </section>
          {keys.length === 0 ? <p>No ticket tags recorded.</p> : <>
            <label className="tag-filters-key-label" htmlFor={`${id}-key`}>Tag key</label>
            <select ref={keyRef} id={`${id}-key`} value={key} onChange={(event) => setSelectedKey(event.target.value)}>
              {keys.map((item) => <option key={item} value={item}>{item}{filters.some((filter) => filter.key === item) ? ' (filtered)' : ''}</option>)}
            </select>
            <fieldset className="tag-filters-values">
              <legend>Values for {key}</legend>
              <label className="multiselect-option"><input type="checkbox" checked={selected === undefined} onChange={() => changeKeyFilter(undefined)} />All values</label>
              <label className="multiselect-option"><input type="checkbox" checked={selected?.missing ?? false} onChange={() => changeKeyFilter({ key, values: selected?.values ?? [], missing: !(selected?.missing ?? false) })} />Missing (key absent)</label>
              {values.map((value) => <label key={value} className="multiselect-option"><input type="checkbox" checked={selected?.values.includes(value) ?? false} onChange={() => toggleValue(value)} /><span>Value: {value}</span></label>)}
            </fieldset>
          </>}
        </div>, document.body)}
    </div>
  );
};
