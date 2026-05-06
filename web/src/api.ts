import type { Config, RowsResponse, Stats } from './types';

const API_BASE = '/api';

export async function fetchSchema(parquetName?: string): Promise<Config> {
  const url = new URL(`${API_BASE}/schema`, window.location.origin);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  const res = await fetch(url.toString());
  if (!res.ok) throw new Error('Failed to fetch schema');
  return res.json();
}

export async function fetchParquets(): Promise<string[]> {
  const res = await fetch(`${API_BASE}/parquets`);
  if (!res.ok) throw new Error('Failed to fetch parquets');
  const data = await res.json();
  return data.parquets || [];
}

export async function fetchStats(parquetName?: string): Promise<Stats> {
  const url = new URL(`${API_BASE}/stats`, window.location.origin);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  const res = await fetch(url.toString());
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
  parquet?: string;
}): Promise<RowsResponse> {
  const url = new URL(`${window.location.origin}${API_BASE}/rows`);
  url.searchParams.set('page', params.page.toString());
  url.searchParams.set('size', params.size.toString());
  if (params.sort) url.searchParams.set('sort', params.sort);
  if (params.order) url.searchParams.set('order', params.order);
  if (params.search) url.searchParams.set('search', params.search);
  if (params.search_col) url.searchParams.set('search_col', params.search_col);
  if (params.parquet) url.searchParams.set('parquet', params.parquet);
  if (params.subdirs) {
    params.subdirs.forEach(s => url.searchParams.append('subdir', s));
  }

  const res = await fetch(url.toString());
  if (!res.ok) throw new Error('Failed to fetch rows');
  return res.json();
}

export async function updateRow(index: number, columns: Record<string, any>, parquetName?: string): Promise<void> {
  const url = new URL(`${API_BASE}/rows/${index}`, window.location.origin);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  const res = await fetch(url.toString(), {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(columns),
  });
  if (!res.ok) throw new Error('Failed to update row');
}

export async function deleteRow(index: number, parquetName?: string): Promise<void> {
  const url = new URL(`${API_BASE}/rows/${index}`, window.location.origin);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  const res = await fetch(url.toString(), {
    method: 'DELETE',
  });
  if (!res.ok) throw new Error('Failed to delete row');
}

export async function fetchSubdirs(column: string, parquetName?: string): Promise<string[]> {
  const url = new URL(`${API_BASE}/subdirs`, window.location.origin);
  url.searchParams.set('col', column);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  const res = await fetch(url.toString());
  if (!res.ok) throw new Error('Failed to fetch subdirs');
  const data = await res.json();
  return data.subdirs;
}

export function getThumbnailUrl(index: number, column?: string, parquetName?: string): string {
  const url = new URL(`${window.location.origin}${API_BASE}/thumbnail`);
  url.searchParams.set('idx', index.toString());
  if (column) url.searchParams.set('col', column);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  return url.toString();
}

export function getFileUrl(index: number, column?: string, parquetName?: string): string {
  const url = new URL(`${window.location.origin}${API_BASE}/file`);
  url.searchParams.set('idx', index.toString());
  if (column) url.searchParams.set('col', column);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  return url.toString();
}

export function getFileDownloadUrl(index: number, column?: string, parquetName?: string): string {
  const url = new URL(`${window.location.origin}${API_BASE}/file`);
  url.searchParams.set('idx', index.toString());
  if (column) url.searchParams.set('col', column);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  url.searchParams.set('dl', '1');
  return url.toString();
}

export function getDownloadUrl(index: number, parquetName?: string): string {
  const url = new URL(`${API_BASE}/rows/${index}/download`, window.location.origin);
  if (parquetName) url.searchParams.set('parquet', parquetName);
  return url.toString();
}
