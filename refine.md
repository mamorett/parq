# Refine Plan — Parq UI Fixes

## Issue 1: List mode must be as wide as the browser

**Root cause**: `web/src/index.css:53-57` constrains `#root` to `width: 1126px`. Since the entire React app mounts inside `#root`, everything including the list view is capped at 1126px.

**Fix** — `web/src/index.css`:
- Change `#root` from `width: 1126px` to `width: 100%`
- Remove `border-inline: 1px solid var(--border)` (adds unwanted vertical borders)
- Remove `text-align: center` (conflicts with card content alignment)
- Remove `margin: 0 auto` (no longer needed when full-width)

---

## Issue 2: Date format in human-readable format, not UNIX timestamp

**Root cause A** (backend): `internal/config/discover.go:116-130` only detects datetime from **string** columns (BYTE_ARRAY with text date patterns). Parquet commonly stores timestamps as `INT64` with a `TIMESTAMP_MILLIS` or `TIMESTAMP_MICROS` logical type, but `detectType()` classifies those as `type: "int"` with no `format: "datetime"`. So `formatValue()` in `RowCard.tsx` renders them as `String(val)`, showing the raw UNIX number.

**Root cause B** (frontend): `formatDate()` in `utils.ts` only handles `number` and `string` types, but Parquet logical timestamps from parquet-go may deserialize as `time.Time` structs (which serialize to JSON as ISO date strings) — those work. The problem is plain INT64 columns that contain epoch timestamps: they arrive as JSON numbers but `col.format` is empty (not `"datetime"`).

**Fix A** — `internal/config/discover.go` in `detectType()`:
- After the BYTE_ARRAY heuristics block, add a check for `INT64` / `INT32` columns:
  - If all non-null sample values are in a plausible timestamp range (e.g. between `946684800` and `32503680000`, the years 2000–3000), set `col.Format = "datetime"` and `col.Type = "int"`.

**Fix B** — `web/src/utils.ts` in `formatDate()`:
- Add handling for BigInt values (parquet-go may return big.Int for INT64)
- Widen the timestamp-seconds vs milliseconds heuristic to handle edge cases

**Fix C** — `web/src/components/RowCard.tsx` in `formatValue()`:
- As a fallback, if `formatDate()` produces a different value than `String(val)` and it looks like a date, use it even if `col.format` is not `"datetime"`. This is a safety net.

---

## Issue 3: Download button must download the entire parquet entry, not just filename

**Root cause**: 
- Backend `internal/api/download.go:29-37` reads only the value of a **single column** (`col` query param) and serves it as a `.txt` file.
- Frontend `RowCard.tsx:52` calls `getDownloadUrl(row.index, thumbnailCol)` which passes just the thumbnail column.
- The route `GET /api/rows/{idx}/download` requires `col` query param.

**Fix A** — `internal/api/download.go`:
- Remove the `col` query param requirement.
- Serialize the entire row's `Columns` map as JSON.
- Set `Content-Type: application/json`.
- Set `Content-Disposition: attachment; filename="row_{idx}.json"`.
- Encode with `json.NewEncoder(w).Encode(row.Columns)`.

**Fix B** — `web/src/api.ts`:
- Change `getDownloadUrl()` to not require a `column` parameter:
  ```typescript
  export function getDownloadUrl(index: number): string {
    return `${API_BASE}/rows/${index}/download`;
  }
  ```

**Fix C** — `web/src/components/RowCard.tsx:52`:
- Update the download button to call `getDownloadUrl(row.index)` without the column argument.

---

## Issue 4: Row editing on click is not implemented

**Root cause**: The backend API (`PUT /api/rows/{idx}`) and frontend API function (`updateRow`) both exist and work, but **no frontend component** calls `updateRow`. The detail dialog (opened by clicking a row card) is read-only — it displays all fields but has no edit inputs, no save button, and no mutation logic.

**What exists**:
- Backend: `PUT /api/rows/{idx}` accepts `map[string]any`, updates in-memory store, rewrites entire Parquet file.
- Schema: `ColumnDef.editable` is `true` for string (non-datetime/non-path) columns.
- Frontend: `updateRow()` in `api.ts` is defined but never imported or called.
- TanStack React Query is already installed — just needs `useMutation`.

**Fix A** — New hook `web/src/hooks/useUpdateRow.ts`:
```typescript
import { useMutation, useQueryClient } from '@tanstack/react-query';
import { updateRow } from '../api';

export function useUpdateRow() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: ({ index, columns }: { index: number; columns: Record<string, any> }) =>
      updateRow(index, columns),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['rows'] });
    },
  });
}
```

**Fix B** — `web/src/components/RowCard.tsx` (dialog section, lines 87–153):
- Add state: `editing` (boolean), `editedValues` (Record<string, string>).
- Add an "Edit" / "Done Editing" toggle button in the dialog header/footer.
- When `editing` is **true**:
  - For columns where `col.editable === true`: render an `<input>` or `<textarea>` instead of the static value display.
  - Bind each input to `editedValues[col.name]`, initialized from `row.columns[col.name]`.
  - Show a "Save" button (with loading state from mutation) in the dialog footer.
  - On save: call `updateRowMutation.mutate({ index: row.index, columns: editedValues })`.
  - On success: close dialog or exit edit mode.
- When `editing` is **false**: show existing read-only view (current behavior).
- Disable editing on `path` and `datetime` columns (they are not `editable: true`).
- Use BlueprintJS `<InputGroup>` or `<EditableText>` for consistent styling with the Nord theme.

**Fix C** — `web/src/components/RowCard.tsx` imports:
- Add `useUpdateRow` hook import
- Add needed BlueprintJS components (`InputGroup`, `Intent`)
- Add `useState` for edit state (already imported)

---

## Files to modify

| File | Issue | Changes |
|------|-------|---------|
| `web/src/index.css` | #1 | Change `#root` width, remove border/text-align |
| `internal/config/discover.go` | #2 | Detect INT64/INT32 timestamp columns as datetime |
| `web/src/utils.ts` | #2 | Improve `formatDate()` for BigInt and edge cases |
| `web/src/components/RowCard.tsx` | #2, #3, #4 | Fallback date formatting, download URL, inline editing UI |
| `internal/api/download.go` | #3 | Download entire row as JSON instead of single column |
| `web/src/api.ts` | #3 | Remove `column` param from `getDownloadUrl()` |
| `web/src/hooks/useUpdateRow.ts` | #4 | **New file** — mutation hook for row updates |

---

## Order of implementation

1. **Issue 1** — Width fix (simple CSS change, no dependencies)
2. **Issue 2** — Date formatting (backend + frontend, no dependencies on other issues)
3. **Issue 3** — Download button (backend + frontend, no dependencies)
4. **Issue 4** — Row editing (depends on #2 for proper date display in edit fields)
