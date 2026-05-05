import { FormGroup, H6, Divider, MenuItem, Button } from '@blueprintjs/core';
import { Select } from '@blueprintjs/select';
import type { Config } from '../types';
import { SearchBar } from './SearchBar';
import { useUrlState } from '../hooks/useUrlState';

interface SortColumn { name: string; label: string }
interface OrderOption { value: string; label: string }

const ORDER_OPTIONS: OrderOption[] = [
  { value: 'asc', label: 'Ascending' },
  { value: 'desc', label: 'Descending' },
];

export function Sidebar({ schema }: { schema: Config }) {
  const { state, updateState } = useUrlState();

  const sortColumns: SortColumn[] = schema.columns.filter(c => c.sortable).map(c => ({
    name: c.name,
    label: c.label,
  }));

  const currentSort = state.sort || schema.default_sort.column;
  const currentOrder = state.order || schema.default_sort.order;
  const currentSize = state.size;

  return (
    <div>
      <H6 style={{ color: 'var(--accent-primary)' }}>Search</H6>
      <SearchBar schema={schema} />

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Sort By">
        <Select<SortColumn>
          items={sortColumns}
          itemRenderer={(item, { handleClick, modifiers }) => {
            if (!modifiers.matchesPredicate) return null;
            return <MenuItem key={item.name} text={item.label} active={item.name === currentSort} onClick={handleClick} />;
          }}
          onItemSelect={(item) => updateState({ sort: item.name })}
          popoverProps={{ minimal: true }}
        >
          <Button minimal small text={sortColumns.find(c => c.name === currentSort)?.label || currentSort} icon="sort" />
        </Select>
      </FormGroup>

      <FormGroup label="Order">
        <Select<OrderOption>
          items={ORDER_OPTIONS}
          itemRenderer={(item, { handleClick, modifiers }) => {
            if (!modifiers.matchesPredicate) return null;
            return <MenuItem key={item.value} text={item.label} active={item.value === currentOrder} onClick={handleClick} />;
          }}
          onItemSelect={(item) => updateState({ order: item.value as 'asc' | 'desc' })}
          popoverProps={{ minimal: true }}
        >
          <Button minimal small text={ORDER_OPTIONS.find(o => o.value === currentOrder)?.label || 'Ascending'} icon="arrows-vertical" />
        </Select>
      </FormGroup>

      <Divider style={{ margin: '1rem 0' }} />

      <FormGroup label="Page Size">
        <Select<number>
          items={schema.pagination.page_size_options}
          itemRenderer={(item, { handleClick, modifiers }) => {
            if (!modifiers.matchesPredicate) return null;
            return <MenuItem key={item} text={`${item}`} active={item === currentSize} onClick={handleClick} />;
          }}
          onItemSelect={(item) => updateState({ size: item, page: 1 })}
          popoverProps={{ minimal: true }}
        >
          <Button minimal small text={`${currentSize}`} icon="numbered-list" />
        </Select>
      </FormGroup>
    </div>
  );
}