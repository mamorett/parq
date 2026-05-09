# Parq — The Extensible Parquet Explorer & Editor

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![BlueprintJS](https://img.shields.io/badge/BlueprintJS-137CBD?style=for-the-badge&logo=blueprintjs&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**Parq** is a high-performance, modern web application designed for exploring, searching, and editing Parquet datasets—specifically optimized for AI/Vision datasets containing image paths and rich metadata.

## 🚀 Key Features

- **📂 Multi-File Support**: Browse multiple Parquet files from a single server instance with a file switcher in the UI.
- **🛠️ Batch Config Generator**: `parq discover --dir /data/` scans a directory and generates `parqs.json` for all Parquet files at once.
- **🔍 Smart Autodiscovery**: Point it at any Parquet file and it will automatically guess column types, identify path columns, and detect datetime formats.
- **📁 Directory Autodiscovery**: Run without a config file and Parq will automatically scan the current directory for all `.parquet` files.
- **🖼️ Media Probing**: Automatically extracts image dimensions, aspect ratios, and file sizes from path columns.
- **⚡ Advanced Boolean Search**: Full-text search with AND (space-separated), OR, NOT (`-` prefix), and parentheses for grouping. Examples: `cat dog` (both), `cat OR dog`, `cat -dog`.
- **📝 Inline Editing**: Edit string columns directly in the UI and persist changes back to the original Parquet file.
- **🔗 Path Remapping**: Use regex rules to reconnect file paths inside your Parquet to different mount points at runtime without mutating the data.
- **🔄 Live Refresh**: Refresh the parquet file list without a full page reload using the Refresh button in the navbar.
- **🎨 Editorial Theme**: A beautiful, professional interface with serif/sans-serif typography and custom color palette.
- **📦 Zero-Config Docker**: Run as a single container with minimal setup.

## 🛠️ Quick Start (Docker)

Point Parq at a directory of Parquet files — auto-discovery handles the rest:

```bash
docker run -p 8080:8080 \
  -v /path/to/your/data:/data \
  trithemius/parq \
  -config /data/parqs.json
```

If `parqs.json` doesn't exist, Parq will automatically scan `/data` for all `.parquet` files:

```bash
docker run -p 8080:8080 \
  -v /path/to/your/data:/data \
  trithemius/parq
```

Or use `-parquet-dir` to specify a different scan directory:

```bash
docker run -p 8080:8080 \
  -v /path/to/your/data:/data \
  trithemius/parq \
  -parquet-dir /data
```

Visit `http://localhost:8080` and you're ready to go!

## ⚙️ Configuration (`parqs.json`)

Define how each Parquet file should be displayed. If a file has no explicit `columns`, Parq auto-discovers its schema at startup.

```jsonc
{
  "parquets": [
    // Auto-discovered — no manual config needed
    { "path": "/data/vision_ai.parquet" },

    // Explicit config for fine-grained control
    {
      "path": "/data/metadata.parquet",
      "columns": [
        {
          "name": "image_path",
          "type": "path",
          "label": "Source Image",
          "probe_dimensions": true,
          "remap": [
            { "pattern": "^/old/mnt/(.+)$", "replace": "/data/new/$1" }
          ]
        },
        {
          "name": "prompt",
          "type": "string",
          "label": "AI Prompt",
          "searchable": true,
          "editable": true
        }
      ],
      "default_sort": { "column": "created_at", "order": "desc" },
      "thumbnail": { "column": "image_path", "max_size": 300, "format": "webp" }
    }
  ]
}
```

| Field | Description |
|-------|-------------|
| `path` | Absolute path to the `.parquet` file |
| `name` | (optional) Display name in the UI file switcher; defaults to filename without extension |
| `columns` | Column definitions (omit to auto-discover) |
| `default_sort` | Initial sort column and direction |
| `pagination` | `default_page_size` and `page_size_options` |
| `thumbnail` | Which column to use for thumbnails, max size, and format (`webp` or `jpeg`) |

## 🔍 Search Syntax

The search bar supports boolean operators for powerful filtering:

| Syntax | Description | Example |
|--------|-------------|---------|
| `word1 word2` | AND: both words must match (implicit) | `cat dog` → rows with both "cat" AND "dog" |
| `word1 OR word2` | OR: either word matches | `cat OR dog` → rows with "cat" OR "dog" |
| `word -exclusion` | NOT: exclude terms | `cat -dog` → "cat" but NOT "dog" |
| `(group)` | Parentheses for grouping | `cat OR (dog bird)` |
| Column selector | Filter search to specific column | Select column from dropdown, then search |

Searches are case-insensitive and match substrings within the selected column or all searchable columns.

### Column definition reference

| Field | Type | Description |
|-------|------|-------------|
| `name` | string | Column name in the Parquet file |
| `type` | `string` \| `int` \| `blob` \| `path` | How to render and interact with values |
| `label` | string | Display label in the table header |
| `searchable` | bool | Include in full-text search |
| `editable` | bool | Allow inline editing in the UI |
| `sortable` | bool | Column supports sorting |
| `copyable` | bool | Show a copy-to-clipboard button |
| `hidden` | bool | Hide from the table view |
| `format` | `"datetime"` or empty | Special rendering (timestamps) |
| `remap` | `[{pattern, replace}]` | Regex path remapping rules |
| `probe_dimensions` | bool | Extract image width/height for path columns |

## 🛠️ CLI Tool: `parq discover`

Generate `parqs.json` config from one or many Parquet files without writing any JSON by hand.

### Single file

```bash
parq discover --parquet /data/file.parquet
# Prints auto-discovered config to stdout
```

### Batch (directory)

```bash
parq discover --dir /data/
# Scans /data/*.parquet, prints MultiConfig JSON

parq discover --dir /data/ --output parqs.json
# Same, written to file
```

### Multiple specific files

```bash
parq discover --parquet /data/a.parquet --parquet /data/b.parquet
```

## 🏗️ Architecture

- **Backend**: Go REST API. Uses a `MultiStore` holding one in-memory `MemoryStore` per Parquet file. Watches all files via `fsnotify` for hot-reload.
- **Frontend**: React 18 SPA using BlueprintJS for a dense, professional UI and TanStack Query for efficient data synchronization.
- **Persistence**: Edits are persisted back to the Parquet file using the `parquet-go` library.
- **Search Engine**: Custom boolean expression parser supporting AND, OR, NOT, and parentheses for complex queries.

## 👩‍💻 Development

### Prerequisites
- Go 1.25+
- Node.js 22+

### Setup
1. Clone the repository.
2. Install dependencies:
   ```bash
   go mod tidy
   cd web && npm install
   ```
3. Generate config:
   ```bash
   go run main.go discover --dir ./testdata/ --output parqs.json
   ```
4. Run the backend:
   ```bash
   go run main.go --config ./parqs.json
   ```
5. Run the frontend:
   ```bash
   cd web && npm run dev
   ```

### Server flags

| Flag | Default | Description |
|------|---------|-------------|
| `-config` | `./parqs.json` | Path to multi-parquet config |
| `-addr` | `:8080` | Listen address |
| `-base-path` | `/` | URL prefix for reverse-proxy |
| `-static-dir` | `./web/dist` | React build directory |
| `-cors-origins` | `*` | Allowed CORS origins |
| `-auto-discover` | `false` | Auto-discover parquet files if config missing |
| `-parquet` | — | Single parquet file path (with `-auto-discover`) |
| `-parquet-dir` | — | Directory to scan for all `.parquet` files |

### Environment variables

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Alternative to `-addr` |
| `CONFIG_PATH` | `./parqs.json` | Alternative to `-config` |

## 📄 License

See [LICENSE](LICENSE) for details.
