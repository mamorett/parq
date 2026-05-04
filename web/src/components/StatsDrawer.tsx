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
      className="theme-editorial"
      style={{ backgroundColor: 'var(--bg-primary)', color: 'var(--text-primary)' }}
    >
      <div className={Classes.DRAWER_BODY} style={{ padding: '2rem' }}>
        {isLoading ? (
          <Text>Loading stats...</Text>
        ) : stats ? (
          <div style={{ display: 'flex', flexDirection: 'column', gap: '1.5rem' }}>
            <div>
              <H5 style={{ color: 'var(--accent-primary)' }}>General</H5>
              <div style={{ backgroundColor: 'var(--bg-secondary)', padding: '1rem', borderRadius: '4px' }}>
                <div style={{ marginBottom: '0.5rem' }}><b>Total Rows:</b> {stats.total_rows.toLocaleString()}</div>
                <div><b>File Size:</b> {(stats.file_size_bytes / 1024 / 1024).toFixed(2)} MB</div>
              </div>
            </div>

            <Divider />

            <div>
              <H5 style={{ color: 'var(--accent-primary)' }}>Images</H5>
              <div style={{ backgroundColor: 'var(--bg-secondary)', padding: '1rem', borderRadius: '4px' }}>
                <div style={{ marginBottom: '0.5rem', color: 'var(--bg-secondary4)' }}><b>Found:</b> {stats.images_found.toLocaleString()}</div>
                <div style={{ color: 'var(--bg-secondary1)' }}><b>Missing:</b> {stats.images_missing.toLocaleString()}</div>
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
