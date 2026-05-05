import { useState, useEffect } from 'react';
import { InputGroup, ControlGroup, MenuItem } from '@blueprintjs/core';
import { Select } from '@blueprintjs/select';
import type { Config } from '../types';
import { useUrlState } from '../hooks/useUrlState';

type SelectItem = { value: string; label: string };

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

  const searchColItems: SelectItem[] = [
    { value: '', label: 'All Columns' },
    ...schema.columns.filter(c => c.searchable).map(c => ({ value: c.name, label: c.label }))
  ];

  return (
    <ControlGroup fill vertical>
      <Select<SelectItem>
        items={searchColItems}
        itemRenderer={(item, { handleClick, modifiers }) => {
          if (!modifiers.matchesPredicate) return null;
          return <MenuItem key={item.value} text={item.label} active={item.value === state.search_col} onClick={handleClick} />;
        }}
        onItemSelect={(item) => updateState({ search_col: item?.value || undefined, page: 1 })}
        popoverProps={{ minimal: true }}
      >
        <button className="bp5-select bp5-button bp5-fill" type="button" style={{ textAlign: 'left', justifyContent: 'flex-start' }}>
          {searchColItems.find(i => i.value === state.search_col)?.label || 'All Columns'}
        </button>
      </Select>
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
