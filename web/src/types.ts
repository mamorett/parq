export interface Config {
  parquet_file: string;
  columns: ColumnDef[];
  default_sort: SortDef;
  pagination: Pagination;
  thumbnail: ThumbnailConfig;
}

export interface ColumnDef {
  name: string;
  type: 'string' | 'int' | 'blob' | 'path';
  label: string;
  searchable: boolean;
  editable: boolean;
  sortable: boolean;
  copyable: boolean;
  hidden: boolean;
  format?: 'datetime';
  remap: Remap[];
  probe_dimensions: boolean;
}

export interface SortDef {
  column: string;
  order: 'asc' | 'desc';
}

export interface Pagination {
  default_page_size: number;
  page_size_options: number[];
}

export interface ThumbnailConfig {
  column: string;
  max_size: number;
  format: 'webp' | 'jpeg';
}

export interface Remap {
  pattern: string;
  replace: string;
}

export interface Row {
  index: number;
  columns: Record<string, any>;
  image_meta?: ImageMeta;
}

export interface ImageMeta {
  width: number;
  height: number;
  aspect: string;
  megapixels: number;
  file_size_kb: number;
}

export interface Stats {
  total_rows: number;
  images_found: number;
  images_missing: number;
  date_range: Record<string, any>;
  file_size_bytes: number;
}

export interface RowsResponse {
  total: number;
  page: number;
  size: number;
  rows: Row[];
}
