import { useState, useEffect } from 'react';
import { TextArea, Button, MenuItem, Tooltip } from '@blueprintjs/core';
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
        <Tooltip
          content={
            <div style={{ maxWidth: '300px' }}>
              <div><strong>Operators:</strong></div>
              <div><code>word1 word2</code> — AND (both match)</div>
              <div><code>word1 OR word2</code> — OR (either matches)</div>
              <div><code>word -exclude</code> — NOT (exclude term)</div>
              <div><code>(group)</code> — Parentheses for grouping</div>
              <div style={{ marginTop: '0.5rem' }}><em>Examples: cat dog, cat OR dog, cat -dog</em></div>
            </div>
          }
          hoverOpenDelay={200}
        >
          <TextArea
            placeholder="Search…"
            value={localSearch}
            onChange={(e) => setLocalSearch(e.target.value)}
            rows={4}
            fill
            style={{ fontFamily: 'var(--font-mono)', fontSize: '0.9rem', paddingRight: '36px' }}
          />
        </Tooltip>
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