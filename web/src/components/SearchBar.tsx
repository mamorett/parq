import { useState, useEffect } from 'react';
import { TextArea, Button, MenuItem } from '@blueprintjs/core';
import { Select } from '@blueprintjs/select';
import type { Config } from '../types';
import { useUrlState } from '../hooks/useUrlState';

interface SearchColumn { name: string; label: string }

const ALL_COLUMNS: SearchColumn = { name: '', label: 'All Columns' };

export function SearchBar({ schema }: { schema: Config }) {
  const { state, updateState } = useUrlState();
  const [localSearch, setLocalSearch] = useState(state.search);

  const columns: SearchColumn[] = [
    ALL_COLUMNS,
    ...schema.columns.filter(c => c.searchable).map(c => ({ name: c.name, label: c.label })),
  ];

  const currentCol = state.search_col || '';

  useEffect(() => {
    const timer = setTimeout(() => {
      if (localSearch !== state.search) {
        updateState({ search: localSearch, page: 1 });
      }
    }, 300);
    return () => clearTimeout(timer);
  }, [localSearch, state.search, updateState]);

  const handleClearSearch = () => {
    setLocalSearch('');
    updateState({ search: '', page: 1 });
  };

  return (
    <div style={{ display: 'flex', flexDirection: 'column', gap: '0.5rem' }}>
      <Select<SearchColumn>
        items={columns}
        itemRenderer={(item, { handleClick, modifiers }) => {
          if (!modifiers.matchesPredicate) return null;
          return <MenuItem key={item.name} text={item.label} active={item.name === currentCol} onClick={handleClick} />;
        }}
        onItemSelect={(item) => updateState({ search_col: item.name || undefined, page: 1 })}
        popoverProps={{ minimal: true }}
      >
        <Button minimal small text={columns.find(c => c.name === currentCol)?.label || 'All Columns'} icon="column-layout" />
      </Select>
      <div style={{ position: 'relative' }}>
        <TextArea
          placeholder="Search (use AND, OR, - for NOT, ( ) for grouping)…&#10;Examples: cat dog (AND), cat OR dog, cat -dog"
          value={localSearch}
          onChange={(e) => setLocalSearch(e.target.value)}
          rows={4}
          fill
          style={{ fontFamily: 'var(--font-mono)', fontSize: '0.9rem', paddingRight: '36px' }}
        />
        {localSearch && (
          <Button
            minimal
            small
            icon="cross"
            onClick={handleClearSearch}
            style={{
              position: 'absolute',
              right: '8px',
              top: '8px',
              zIndex: 1,
            }}
          />
        )}
      </div>
    </div>
  );
}