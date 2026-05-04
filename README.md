# Parq — The Extensible Parquet Explorer & Editor

![Go](https://img.shields.io/badge/Go-00ADD8?style=for-the-badge&logo=go&logoColor=white)
![React](https://img.shields.io/badge/React-20232A?style=for-the-badge&logo=react&logoColor=61DAFB)
![BlueprintJS](https://img.shields.io/badge/BlueprintJS-137CBD?style=for-the-badge&logo=blueprintjs&logoColor=white)
![Docker](https://img.shields.io/badge/Docker-2496ED?style=for-the-badge&logo=docker&logoColor=white)
![License](https://img.shields.io/badge/License-MIT-green?style=for-the-badge)

**Parq** is a high-performance, modern web application designed for exploring, searching, and editing Parquet datasets—specifically optimized for AI/Vision datasets containing image paths and rich metadata.

## 🚀 Key Features

- **🔍 Smart Autodiscovery**: Point it at any Parquet file and it will automatically guess column types, identify path columns, and detect datetime formats.
- **🖼️ Media Probing**: Automatically extracts image dimensions, aspect ratios, and file sizes from path columns.
- **⚡ Fast Search & Filter**: Substring search across all columns, exact filters, and subdirectory-based path filtering.
- **📝 Inline Editing**: Edit string columns directly in the UI and persist changes back to the original Parquet file.
- **🔗 Path Remapping**: Use regex rules to reconnect file paths inside your Parquet to different mount points at runtime without mutating the data.
- **🌑 Nord Dark Theme**: A beautiful, eye-friendly interface built with BlueprintJS and the Nord color palette.
- **📦 Zero-Config Docker**: Run as a single container with minimal setup.

## 🛠️ Quick Start (Docker)

The fastest way to explore a Parquet file:

```bash
docker run -p 8080:8080 \
  -v /path/to/your/data:/data \
  trithemius/parq \
  -parquet /data/your_file.parquet
```

Visit `http://localhost:8080` and you're ready to go!

## ⚙️ Configuration (`schema.json`)

While Parq can autodiscover your schema, you can customize the experience with a `schema.json` file.

```json
{
  "parquet_file": "/data/vision_ai.parquet",
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
  "default_sort": { "column": "created_at", "order": "desc" }
}
```

## 🏗️ Architecture

- **Backend**: Go 1.25 REST API. Uses an in-memory `RowStore` for blazing-fast queries and `fsnotify` for hot-reloading when the Parquet file changes.
- **Frontend**: React 18 SPA using BlueprintJS for a dense, professional UI and TanStack Query for efficient data synchronization.
- **Persistence**: Edits are persisted back to the Parquet file using the `parquet-go` library.

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
3. Run the backend:
   ```bash
   go run main.go -parquet ./data.parquet -auto-discover
   ```
4. Run the frontend:
   ```bash
   cd web && npm run dev
   ```

## 📄 License

MIT © [Trithemius](https://github.com/trithemius)
