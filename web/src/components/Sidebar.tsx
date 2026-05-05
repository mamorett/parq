import { FormGroup, HTMLSelect, H6, Divider } from '@blueprintjs/core';
import type { Config } from '../types';
import { SearchBar } from './SearchBar';
import { useUrlState } from '../hooks/useUrlState';

export function Sidebar({ schema }: { schema: Config }) {
  const { state, updateState } = useUrlState();

  return (
    <div>
      <H6 style={{ color: 'var(--accent-primary)' }}>Search</H6>
      <SearchBar schema={schema} />

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Sort By" labelFor="sort-select">
        <HTMLSelect
          id="sort-select"
          fill
          value={state.sort || schema.default_sort.column}
          onChange={(e) => updateState({ sort: e.target.value })}
        >
          {schema.columns.filter(c => c.sortable).map(c => (
            <option key={c.name} value={c.name}>{c.label}</option>
          ))}
        </HTMLSelect>
      </FormGroup>

      <FormGroup label="Order">
        <HTMLSelect
          fill
          value={state.order || schema.default_sort.order}
          onChange={(e) => updateState({ order: e.target.value as 'asc' | 'desc' })}
        >
          <option value="asc">Ascending</option>
          <option value="desc">Descending</option>
        </HTMLSelect>
      </FormGroup>

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Page Size">
        <HTMLSelect
          fill
          value={state.size}
          onChange={(e) => updateState({ size: parseInt(e.target.value, 10), page: 1 })}
        >
          {schema.pagination.page_size_options.map(s => (
            <option key={s} value={s}>{s}</option>
          ))}
        </HTMLSelect>
      </FormGroup>
    </div>
  );
}
