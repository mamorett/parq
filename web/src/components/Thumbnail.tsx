import { useState } from 'react';
import { Spinner, Icon } from '@blueprintjs/core';
import { getThumbnailUrl } from '../api';

export function Thumbnail({ index, column, parquetName }: { index: number; column?: string; parquetName?: string }) {
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState(false);
  const url = getThumbnailUrl(index, column, parquetName);

  return (
    <div style={{
      width: '100%',
      aspectRatio: '1',
      backgroundColor: 'var(--bg-secondary)',
      borderRadius: '4px',
      display: 'flex',
      alignItems: 'center',
      justifyContent: 'center',
      overflow: 'hidden',
      position: 'relative'
    }}>
      {loading && <Spinner size={20} style={{ position: 'absolute' }} />}
      {error ? (
        <Icon icon="error" size={40} style={{ color: 'var(--text-muted)' }} />
      ) : (
        <img
          src={url}
          alt={`Row ${index}`}
          style={{ width: '100%', height: '100%', objectFit: 'cover', opacity: loading ? 0 : 1, transition: 'opacity 0.2s' }}
          onLoad={() => setLoading(false)}
          onError={() => {
            setLoading(false);
            setError(true);
          }}
          loading="lazy"
        />
      )}
    </div>
  );
}
