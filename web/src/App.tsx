import { useEffect, useState } from 'react';
import { Navbar, NavbarGroup, Alignment, Button, Classes, NonIdealState, Spinner, MenuItem } from '@blueprintjs/core';
import { Select } from '@blueprintjs/select';
import { useSchema, useParquets } from './hooks/useSchema';
import { useUrlState } from './hooks/useUrlState';
import { Layout } from './components/Layout';
import { StatsDrawer } from './components/StatsDrawer';

function App() {
  const { data: parquets, isLoading: loadingParquets } = useParquets();
  const [activeParquet, setActiveParquet] = useState<string>('');
  const { data: schema, isLoading, error } = useSchema(activeParquet || undefined);
  const { state, updateState } = useUrlState();

  // Initialize active parquet from URL or first available
  useEffect(() => {
    const urlParquet = new URLSearchParams(window.location.search).get('parquet');
    if (urlParquet) {
      setActiveParquet(urlParquet);
    } else if (parquets && parquets.length > 0) {
      setActiveParquet(parquets[0]);
    }
  }, [parquets]);

  // Update URL when parquet changes
  useEffect(() => {
    if (activeParquet) {
      const url = new URL(window.location.href);
      url.searchParams.set('parquet', activeParquet);
      window.history.replaceState({}, '', url.toString());
    }
  }, [activeParquet]);

  useEffect(() => {
    if (schema && state.size === 0) {
      updateState({ size: schema.pagination.default_page_size });
    }
  }, [schema, state.size, updateState]);

  if (isLoading || loadingParquets) {
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

  const parquetSelect = (
    <Select
      items={parquets || []}
      itemRenderer={(item, { handleClick, modifiers }) => {
        if (!modifiers.matchesPredicate) return null;
        return (
          <MenuItem
            key={item}
            text={item}
            active={item === activeParquet}
            onClick={handleClick}
          />
        );
      }}
      onItemSelect={(item) => setActiveParquet(item as string)}
      activeItem={activeParquet}
      fill={false}
    />
  );

  return (
    <div className="theme-editorial">
      <Navbar className="theme-editorial">
        <NavbarGroup align={Alignment.LEFT}>
          <Navbar.Heading style={{ fontWeight: 'bold', color: 'var(--accent-primary)' }}>Parq</Navbar.Heading>
          <Navbar.Divider />
          {parquets && parquets.length > 1 && (
            <>
              <span style={{ color: 'var(--text-secondary)', fontSize: '14px' }}>File:</span>
              {parquetSelect}
              <Navbar.Divider />
            </>
          )}
          <Button className={Classes.MINIMAL} icon="home" text="Explorer" />
          <Button className={Classes.MINIMAL} icon="info-sign" text="Stats" onClick={() => updateState({ showStats: true })} />
        </NavbarGroup>
      </Navbar>
      <Layout schema={schema!} parquetName={activeParquet} />
      <StatsDrawer isOpen={state.showStats} onClose={() => updateState({ showStats: false })} parquetName={activeParquet} />
    </div>
  );
}

export default App;
