import { Button, ButtonGroup, Text, InputGroup } from '@blueprintjs/core';
import { useState, useEffect } from 'react';
import { useUrlState } from '../hooks/useUrlState';

export function PaginationBar({ total }: { total: number }) {
  const { state, updateState } = useUrlState();
  const totalPages = Math.ceil(total / state.size);
  const currentPage = state.page;

  const [inputPage, setInputPage] = useState(currentPage.toString());

  useEffect(() => {
    setInputPage(currentPage.toString());
  }, [currentPage]);

  const handlePageSubmit = () => {
    const p = parseInt(inputPage, 10);
    if (!isNaN(p) && p >= 1 && p <= totalPages) {
      updateState({ page: p });
    } else {
      setInputPage(currentPage.toString());
    }
  };

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: 'var(--bg-secondary)', borderRadius: '4px' }}>
      <Text className="pagination-text">Total: <b>{total}</b> rows</Text>
      <ButtonGroup>
        <Button
          icon="double-chevron-left"
          disabled={currentPage === 1}
          onClick={() => updateState({ page: 1 })}
        />
        <Button
          icon="chevron-left"
          disabled={currentPage === 1}
          onClick={() => updateState({ page: currentPage - 1 })}
        />
        <div style={{ display: 'flex', alignItems: 'center', padding: '0 0.5rem' }} className="pagination-text">
          Page 
          <InputGroup
            style={{ width: '60px', margin: '0 0.5rem' }}
            value={inputPage}
            onChange={(e: React.ChangeEvent<HTMLInputElement>) => setInputPage(e.target.value)}
            onKeyDown={(e: React.KeyboardEvent<HTMLInputElement>) => {
              if (e.key === 'Enter') {
                handlePageSubmit();
              }
            }}
            onBlur={handlePageSubmit}
          />
          of {totalPages || 1}
        </div>
        <Button
          icon="chevron-right"
          disabled={currentPage === totalPages || totalPages === 0}
          onClick={() => updateState({ page: currentPage + 1 })}
        />
        <Button
          icon="double-chevron-right"
          disabled={currentPage === totalPages || totalPages === 0}
          onClick={() => updateState({ page: totalPages })}
        />
      </ButtonGroup>
    </div>
  );
}
