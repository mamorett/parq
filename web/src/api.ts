import type { Config, RowsResponse, Stats } from './types';

const API_BASE = '/api';

export async function fetchSchema(): Promise<Config> {
  const res = await fetch(`${API_BASE}/schema`);
  if (!res.ok) throw new Error('Failed to fetch schema');
  return res.json();
}

export async function fetchStats(): Promise<Stats> {
  const res = await fetch(`${API_BASE}/stats`);
  if (!res.ok) throw new Error('Failed to fetch stats');
  return res.json();
}

export async function fetchRows(params: {
  page: number;
  size: number;
  sort?: string;
  order?: string;
  search?: string;
  search_col?: string;
  subdirs?: string[];
}): Promise<RowsResponse> {
  const url = new URL(`${window.location.origin}${API_BASE}/rows`);
  url.searchParams.set('page', params.page.toString());
  url.searchParams.set('size', params.size.toString());
  if (params.sort) url.searchParams.set('sort', params.sort);
  if (params.order) url.searchParams.set('order', params.order);
  if (params.search) url.searchParams.set('search', params.search);
  if (params.search_col) url.searchParams.set('search_col', params.search_col);
  if (params.subdirs) {
    params.subdirs.forEach(s => url.searchParams.append('subdir', s));
  }

  const res = await fetch(url.toString());
  if (!res.ok) throw new Error('Failed to fetch rows');
  return res.json();
}

export async function updateRow(index: number, columns: Record<string, any>): Promise<void> {
  const res = await fetch(`${API_BASE}/rows/${index}`, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(columns),
  });
  if (!res.ok) throw new Error('Failed to update row');
}

export async function fetchSubdirs(column: string): Promise<string[]> {
  const res = await fetch(`${API_BASE}/subdirs?col=${column}`);
  if (!res.ok) throw new Error('Failed to fetch subdirs');
  const data = await res.json();
  return data.subdirs;
}

export function getThumbnailUrl(index: number, column?: string): string {
  const url = new URL(`${window.location.origin}${API_BASE}/thumbnail`);
  url.searchParams.set('idx', index.toString());
  if (column) url.searchParams.set('col', column);
  return url.toString();
}

export function getFileUrl(index: number, column?: string): string {
  const url = new URL(`${window.location.origin}${API_BASE}/file`);
  url.searchParams.set('idx', index.toString());
  if (column) url.searchParams.set('col', column);
  return url.toString();
}

export function getDownloadUrl(index: number, column: string): string {
  return `${API_BASE}/rows/${index}/download?col=${column}`;
}
