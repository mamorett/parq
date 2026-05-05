package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/trithemius/parq/internal/api"
	"github.com/trithemius/parq/internal/config"
	"github.com/trithemius/parq/internal/store"
	"github.com/trithemius/parq/internal/watcher"
)

func main() {
	// Check for subcommand before defining flags
	if len(os.Args) > 1 && os.Args[1] == "discover" {
		handleDiscover()
		return
	}

	addr := flag.String("addr", ":8080", "listen address")
	schemaPath := flag.String("schema", "./schema.json", "path to schema.json (legacy)")
	configPath := flag.String("config", "./parqs.json", "path to parqs.json (multi-parquet config)")

	// Repeatable --parquet flags for server mode
	var parquetPaths []string
	mf := &multiStringFlag{&parquetPaths}
	flag.Var(mf, "parquet", "path to a parquet file (repeatable)")

	basePath := flag.String("base-path", "/", "URL prefix for reverse-proxy support")
	staticDir := flag.String("static-dir", "./web/dist", "path to React build")
	corsOrigins := flag.String("cors-origins", "*", "allowed CORS origins")
	autoDiscover := flag.Bool("auto-discover", false, "generate parqs.json if it doesn't exist")
	parquetDir := flag.String("parquet-dir", "", "directory to scan for all .parquet files (autodiscovery)")

	flag.Parse()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("Starting Parq server", "addr", *addr)

	// Helper function to scan directory for parquet files
	scanParquetDir := func(dir string) ([]string, error) {
		entries, err := os.ReadDir(dir)
		if err != nil {
			return nil, err
		}
		var files []string
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".parquet") {
				files = append(files, filepath.Join(dir, entry.Name()))
			}
		}
		return files, nil
	}

	// If --parquet-dir is provided, scan directory for all parquet files
	if *parquetDir != "" {
		files, err := scanParquetDir(*parquetDir)
		if err != nil {
			slog.Error("Failed to scan directory", "dir", *parquetDir, "error", err)
			os.Exit(1)
		}
		if len(files) == 0 {
			slog.Error("No parquet files found in directory", "dir", *parquetDir)
			os.Exit(1)
		}
		parquetPaths = append(parquetPaths, files...)
	}

	// If --parquet flags were provided, use them to create a MultiConfig
	if len(parquetPaths) > 0 {
		var entries []config.ParquetEntry
		for _, path := range parquetPaths {
			discovered, err := config.Discover(path)
			if err != nil {
				slog.Error("Failed to discover parquet", "path", path, "error", err)
				os.Exit(1)
			}
			base := filepath.Base(path)
			name := strings.TrimSuffix(base, filepath.Ext(base))
			entries = append(entries, config.ParquetEntry{
				Path:      path,
				Name:      name,
				Columns:   discovered.Columns,
				Thumbnail: discovered.Thumbnail,
			})
		}
		mc := &config.MultiConfig{Parquets: entries}

		dataStore, err := store.NewMultiStore(mc)
		if err != nil {
			slog.Error("Failed to initialize store", "error", err)
			os.Exit(1)
		}

		// Watch for changes in all parquet files
		var storeParquetPaths []string
		for _, entry := range mc.Parquets {
			storeParquetPaths = append(storeParquetPaths, entry.Path)
		}
		if err := watcher.WatchMany(storeParquetPaths, func(name string) {
			if err := dataStore.Reload(name); err != nil {
				slog.Error("Failed to reload store", "name", name, "error", err)
			}
		}); err != nil {
			slog.Error("Failed to start watcher", "error", err)
		}

		router := api.NewRouter(dataStore, mc, *staticDir, *corsOrigins, *basePath)

		server := &http.Server{
			Addr:    *addr,
			Handler: router,
		}

		if err := server.ListenAndServe(); err != nil {
			slog.Error("Server failed", "error", err)
			os.Exit(1)
		}
		return
	}

	// Try to load multi-config first, fall back to legacy schema.json
	mc, err := config.LoadMulti(*configPath)
	if err != nil {
		if os.IsNotExist(err) && *schemaPath != "./schema.json" {
			// User specified a custom schema path, try legacy load
			mc, err = config.LoadLegacy(*schemaPath, "", *autoDiscover)
		} else if os.IsNotExist(err) && *autoDiscover {
			// Auto-discover mode: requires --parquet or --parquet-dir flags
			slog.Error("Auto-discover requires -parquet or -parquet-dir flag")
			os.Exit(1)
		} else if os.IsNotExist(err) {
			// Try legacy schema.json as fallback
			mc, err = config.LoadLegacy(*schemaPath, "", false)
			if err != nil && os.IsNotExist(err) {
				// No config file found, default to autodiscovery from current directory
				slog.Info("No config file found, scanning for parquet files in current directory")
				files, scanErr := scanParquetDir(".")
				if scanErr != nil {
					slog.Error("Failed to scan directory", "dir", ".", "error", scanErr)
					os.Exit(1)
				}
				if len(files) == 0 {
					slog.Error("No parquet files found in current directory and no config file")
					os.Exit(1)
				}
				// Scan complete, restart with discovered files
				parquetPaths = files
				var entries []config.ParquetEntry
				for _, path := range parquetPaths {
					discovered, err := config.Discover(path)
					if err != nil {
						slog.Error("Failed to discover parquet", "path", path, "error", err)
						os.Exit(1)
					}
					base := filepath.Base(path)
					name := strings.TrimSuffix(base, filepath.Ext(base))
					entries = append(entries, config.ParquetEntry{
						Path:      path,
						Name:      name,
						Columns:   discovered.Columns,
						Thumbnail: discovered.Thumbnail,
					})
				}
				mc = &config.MultiConfig{Parquets: entries}
			}
		}
		if err != nil {
			slog.Error("Failed to load config", "error", err)
			os.Exit(1)
		}
	}

	dataStore, err := store.NewMultiStore(mc)
	if err != nil {
		slog.Error("Failed to initialize store", "error", err)
		os.Exit(1)
	}

	// Watch for changes in all parquet files
	var storeParquetPaths []string
	for _, entry := range mc.Parquets {
		storeParquetPaths = append(storeParquetPaths, entry.Path)
	}
	if err := watcher.WatchMany(storeParquetPaths, func(name string) {
		if err := dataStore.Reload(name); err != nil {
			slog.Error("Failed to reload store", "name", name, "error", err)
		}
	}); err != nil {
		slog.Error("Failed to start watcher", "error", err)
	}

	router := api.NewRouter(dataStore, mc, *staticDir, *corsOrigins, *basePath)

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
	var parquetPaths []string
	dirPath := flag.String("dir", "", "directory to scan for *.parquet files")
	outputPath := flag.String("output", "", "where to write parqs.json (default: stdout)")

	// Parse repeatable --parquet flags
	mf := &multiStringFlag{&parquetPaths}
	flag.Var(mf, "parquet", "path to a parquet file (repeatable)")
	
	// Parse flags starting from os.Args[2] (skip program name and "discover" subcommand)
	flag.CommandLine.Parse(os.Args[2:])

	// Collect parquet files
	var files []string

	if *dirPath != "" {
		// Scan directory for .parquet files
		entries, err := os.ReadDir(*dirPath)
		if err != nil {
			fmt.Printf("Error reading directory: %v\n", err)
			os.Exit(1)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(strings.ToLower(entry.Name()), ".parquet") {
				files = append(files, filepath.Join(*dirPath, entry.Name()))
			}
		}
	}

	// Add explicit --parquet paths
	files = append(files, parquetPaths...)

	if len(files) == 0 {
		fmt.Println("Error: --parquet or --dir is required")
		os.Exit(1)
	}

	// Discover each file
	var entries []config.ParquetEntry
	for _, path := range files {
		cfg, err := config.Discover(path)
		if err != nil {
			fmt.Printf("Error discovering %s: %v\n", path, err)
			os.Exit(1)
		}
		base := filepath.Base(path)
		name := strings.TrimSuffix(base, filepath.Ext(base))
		entries = append(entries, config.ParquetEntry{
			Path:   path,
			Name:   name,
			Columns: cfg.Columns,
		})
	}

	mc := &config.MultiConfig{Parquets: entries}

	if *outputPath != "" {
		if err := saveMultiConfig(mc, *outputPath); err != nil {
			fmt.Printf("Error saving: %v\n", err)
			os.Exit(1)
		}
	} else {
		data, _ := json.MarshalIndent(mc, "", "  ")
		fmt.Println(string(data))
	}
}

// multiStringFlag allows repeatable flag usage: --parquet a.parquet --parquet b.parquet
type multiStringFlag struct {
	values *[]string
}

func (m *multiStringFlag) String() string {
	if m == nil || m.values == nil || *m.values == nil {
		return ""
	}
	return strings.Join(*m.values, ",")
}

func (m *multiStringFlag) Set(value string) error {
	*m.values = append(*m.values, value)
	return nil
}

// saveMultiConfig saves a MultiConfig to a file
func saveMultiConfig(mc *config.MultiConfig, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	return enc.Encode(mc)
}
