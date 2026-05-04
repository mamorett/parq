import type { Config } from '../types';
import { Sidebar } from './Sidebar';
import { RowList } from './RowList';
import { useUrlState } from '../hooks/useUrlState';

export function Layout({ schema }: { schema: Config }) {
  const { state } = useUrlState();

  return (
    <div className="app-layout">
      <div className="sidebar">
        <Sidebar schema={schema} />
      </div>
      <div className="main-content">
        {state.size > 0 && <RowList schema={schema} />}
      </div>
    </div>
  );
}
