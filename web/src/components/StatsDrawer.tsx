import { Drawer, H5, Divider, Classes, Text } from '@blueprintjs/core';
import { useQuery } from '@tanstack/react-query';
import { fetchStats } from '../api';

export function StatsDrawer({ isOpen, onClose }: { isOpen: boolean, onClose: () => void }) {
  const { data: stats, isLoading } = useQuery({
    queryKey: ['stats'],
    queryFn: fetchStats,
    enabled: isOpen,
  });

  return (
    <Drawer
      isOpen={isOpen}
      onClose={onClose}
      title="Dataset Statistics"
      icon="info-sign"
      size="400px"
      className={Classes.DARK}
      style={{ backgroundColor: 'var(--nord0)', color: 'var(--nord6)' }}
    >
      <div className={Classes.DRAWER_BODY} style={{ padding: '2rem' }}>
        {isLoading ? (
          <Text>Loading stats...</Text>
        ) : stats ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            <div>
              <H5 style={{ color: 'var(--nord8)' }}>General</H5>
              <div style={{ backgroundColor: 'var(--nord1)', padding: '1rem', borderRadius: '4px' }}>
                <div style={{ marginBottom: '0.5rem' }}><b>Total Rows:</b> {stats.total_rows.toLocaleString()}</div>
                <div><b>File Size:</b> {(stats.file_size_bytes / 1024 / 1024).toFixed(2)} MB</div>
              </div>
            </div>

            <Divider />

            <div>
              <H5 style={{ color: 'var(--nord8)' }}>Images</H5>
              <div style={{ backgroundColor: 'var(--nord1)', padding: '1rem', borderRadius: '4px' }}>
                <div style={{ marginBottom: '0.5rem', color: 'var(--nord14)' }}><b>Found:</b> {stats.images_found.toLocaleString()}</div>
                <div style={{ color: 'var(--nord11)' }}><b>Missing:</b> {stats.images_missing.toLocaleString()}</div>
              </div>
            </div>
          </div>
        ) : (
          <Text>No statistics available.</Text>
        )}
      </div>
    </Drawer>
  );
}
