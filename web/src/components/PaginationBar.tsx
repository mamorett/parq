import { Button, ButtonGroup, Text } from '@blueprintjs/core';
import { useUrlState } from '../hooks/useUrlState';

export function PaginationBar({ total }: { total: number }) {
  const { state, updateState } = useUrlState();
  const totalPages = Math.ceil(total / state.size);
  const currentPage = state.page;

  return (
    <div style={{ display: 'flex', alignItems: 'center', justifyContent: 'space-between', padding: '0.5rem', backgroundColor: 'var(--nord1)', borderRadius: '4px' }}>
      <Text>Total: <b>{total}</b> rows</Text>
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
        <Button className="bp5-minimal" disabled>
          Page {currentPage} of {totalPages || 1}
        </Button>
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
