import { useState, useEffect } from 'react';
import { InputGroup, HTMLSelect, ControlGroup } from '@blueprintjs/core';
import type { Config } from '../types';
import { useUrlState } from '../hooks/useUrlState';

export function SearchBar({ schema }: { schema: Config }) {
  const { state, updateState } = useUrlState();
  const [localSearch, setLocalSearch] = useState(state.search);

  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== state.search) {
        updateState({ search: localSearch, page: 1 });
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [localSearch, state.search, updateState]);

  return (
    <ControlGroup fill vertical>
      <HTMLSelect
        value={state.search_col || ''}
        onChange={(e) => updateState({ search_col: e.target.value || undefined, page: 1 })}
      >
        <option value="">All Columns</option>
        {schema.columns.filter(c => c.searchable).map(c => (
          <option key={c.name} value={c.name}>{c.label}</option>
        ))}
      </HTMLSelect>
      <InputGroup
        leftIcon="search"
        placeholder="Search term..."
        value={localSearch}
        onChange={(e) => setLocalSearch(e.target.value)}
        onKeyDown={(e) => {
          if (e.key === 'Enter') {
            updateState({ search: localSearch, page: 1 });
          }
        }}
      />
    </ControlGroup>
  );
}
