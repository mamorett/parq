import type { Config } from '../types';
import { useRows } from '../hooks/useRows';
import { useUrlState } from '../hooks/useUrlState';
import { RowCard } from './RowCard';
import { PaginationBar } from './PaginationBar';
import { NonIdealState, Spinner } from '@blueprintjs/core';

export function RowList({ schema, parquetName }: { schema: Config; parquetName?: string }) {
  const { state } = useUrlState();
  const { data, isLoading, isRefetching, error } = useRows({ ...state, parquet: parquetName || undefined });

  if (isLoading) {
    return <div style={{ padding: '2rem', textAlign: 'center' }}><Spinner /></div>;
  }

  if (error) {
    return <NonIdealState icon="error" title="Error loading rows" description={(error as Error).message} />;
  }

  if (!data || data.rows.length === 0) {
    return <NonIdealState icon="search" title="No results found" description="Try adjusting your filters or search term." />;
  }

  return (
    <div>
      {isRefetching && (
        <div style={{
          height: '3px', backgroundColor: 'var(--accent-primary)',
          marginBottom: '0.5rem',
          animation: 'refetch-pulse 0.8s ease-in-out infinite',
        }} />
      )}
      <PaginationBar total={data.total} />
      <div className="row-card-list" style={{ margin: '1rem 0' }}>
        {data.rows.map(row => (
          <RowCard key={row.index} row={row} schema={schema} parquetName={parquetName} />
        ))}
      </div>
      <PaginationBar total={data.total} />
    </div>
  );
}
