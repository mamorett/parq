import { FormGroup, H6, Divider, MenuItem } from '@blueprintjs/core';
import { Select } from '@blueprintjs/select';
import type { Config } from '../types';
import { SearchBar } from './SearchBar';
import { useUrlState } from '../hooks/useUrlState';

type SelectItem = { value: string; label: string };

function createSelect(items: SelectItem[], value: string, onChange: (v: string) => void) {
  return (
    <Select<SelectItem>
      items={items}
      itemRenderer={(item, { handleClick, modifiers }) => {
        if (!modifiers.matchesPredicate) return null;
        return <MenuItem key={item.value} text={item.label} active={item.value === value} onClick={handleClick} />;
      }}
      onItemSelect={(item) => onChange(item.value)}
      popoverProps={{ minimal: true }}
    >
      <button className="bp5-select bp5-button bp5-fill" type="button" style={{ textAlign: 'left', justifyContent: 'flex-start' }}>
        {items.find(i => i.value === value)?.label || items[0]?.label}
      </button>
    </Select>
  );
}

export function Sidebar({ schema }: { schema: Config }) {
  const { state, updateState } = useUrlState();

  return (
    <div>
      <H6 style={{ color: 'var(--accent-primary)' }}>Search</H6>
      <SearchBar schema={schema} />

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Sort By" labelFor="sort-select">
        {createSelect(
          schema.columns.filter(c => c.sortable).map(c => ({ value: c.name, label: c.label })),
          state.sort || schema.default_sort.column,
          (v) => updateState({ sort: v })
        )}
      </FormGroup>

      <FormGroup label="Order">
        {createSelect(
          [{ value: 'asc', label: 'Ascending' }, { value: 'desc', label: 'Descending' }],
          state.order || schema.default_sort.order,
          (v) => updateState({ order: v as 'asc' | 'desc' })
        )}
      </FormGroup>

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Page Size">
        {createSelect(
          schema.pagination.page_size_options.map(s => ({ value: String(s), label: String(s) })),
          String(state.size),
          (v) => updateState({ size: parseInt(v, 10), page: 1 })
        )}
      </FormGroup>
    </div>
  );
}
