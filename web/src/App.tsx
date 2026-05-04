import { useEffect, useState } from 'react';
import { Navbar, Alignment, Button, Classes, NonIdealState, Spinner } from '@blueprintjs/core';
import { useSchema } from './hooks/useSchema';
import { useUrlState } from './hooks/useUrlState';
import { Layout } from './components/Layout';
import { StatsDrawer } from './components/StatsDrawer';

function App() {
  const { data: schema, isLoading, error } = useSchema();
  const { state, updateState } = useUrlState();
  const [isStatsOpen, setIsStatsOpen] = useState(false);

  useEffect(() => {
    if (schema && state.size === 0) {
      updateState({ size: schema.pagination.default_page_size });
    }
  }, [schema, state.size, updateState]);

  if (isLoading) {
    return (
      <div style={{ display: 'flex', height: '100vh', alignItems: 'center', justifyContent: 'center' }}>
        <Spinner size={50} />
      </div>
    );
  }

  if (error) {
    return (
      <NonIdealState
        icon="error"
        title="Failed to load schema"
        description={(error as Error).message}
        action={<Button icon="refresh" onClick={() => window.location.reload()}>Retry</Button>}
      />
    );
  }

  return (
    <div className={Classes.DARK}>
      <Navbar className={Classes.DARK}>
        <Navbar.Group align={Alignment.LEFT}>
          <Navbar.Heading style={{ fontWeight: 'bold', color: 'var(--nord8)' }}>Parq</Navbar.Heading>
          <Navbar.Divider />
          <Button className={Classes.MINIMAL} icon="home" text="Explorer" />
          <Button className={Classes.MINIMAL} icon="info-sign" text="Stats" onClick={() => setIsStatsOpen(true)} />
        </Navbar.Group>
      </Navbar>
      <Layout schema={schema!} />
      <StatsDrawer isOpen={isStatsOpen} onClose={() => setIsStatsOpen(false)} />
    </div>
  );
}

export default App;
