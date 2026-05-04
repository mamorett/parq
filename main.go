package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/trithemius/parq/internal/api"
	"github.com/trithemius/parq/internal/config"
	"github.com/trithemius/parq/internal/store"
	"github.com/trithemius/parq/internal/watcher"
)

func main() {
	addr := flag.String("addr", ":8080", "listen address")
	schemaPath := flag.String("schema", "./schema.json", "path to schema.json")
	parquetPath := flag.String("parquet", "", "override parquet_file from schema.json")
	basePath := flag.String("base-path", "/", "URL prefix for reverse-proxy support")
	staticDir := flag.String("static-dir", "./web/dist", "path to React build")
	corsOrigins := flag.String("cors-origins", "*", "allowed CORS origins")
	autoDiscover := flag.Bool("auto-discover", false, "generate schema.json if it doesn't exist")

	flag.Parse()

	// Subcommand handle
	if flag.Arg(0) == "discover" {
		handleDiscover()
		return
	}

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Parq server", "addr", *addr)

	cfg, err := config.Load(*schemaPath, *parquetPath, *autoDiscover)
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	dataStore, err := store.NewMemoryStore(cfg)
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Watch for changes
	if err := watcher.Watch(cfg.ParquetFile, func() {
		if err := dataStore.Reload(); err != nil {
			slog.Error("Failed to reload store", "error", err)
		}
	}); err != nil {
		slog.Error("Failed to start watcher", "error", err)
	}

	router := api.NewRouter(dataStore, cfg, *staticDir, *corsOrigins, *basePath)

	server := &http.Server{
		Addr:    *addr,
		Handler: router,
	}

	if err := server.ListenAndServe(); err != nil {
		slog.Error("Server failed", "error", err)
		os.Exit(1)
	}
}

func handleDiscover() {
	parquetPath := flag.String("parquet", "", "path to parquet file")
	outputPath := flag.String("output", "", "where to write schema.json (default: stdout)")
	flag.CommandLine.Parse(os.Args[2:])

	if *parquetPath == "" {
		fmt.Println("Error: -parquet is required")
		os.Exit(1)
	}

	cfg, err := config.Discover(*parquetPath)
	if err != nil {
		fmt.Printf("Error: %v\n", err)
		os.Exit(1)
	}

	if *outputPath != "" {
		if err := cfg.Save(*outputPath); err != nil {
			fmt.Printf("Error saving: %v\n", err)
			os.Exit(1)
		}
	} else {
		data, _ := json.MarshalIndent(cfg, "", "  ")
		fmt.Println(string(data))
	}
}
