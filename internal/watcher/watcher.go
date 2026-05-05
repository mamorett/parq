package watcher

import (
	"log/slog"
	"path/filepath"

	"github.com/fsnotify/fsnotify"
)

// Watch watches a single parquet file for changes
func Watch(path string, callback func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					slog.Info("Parquet file modified, reloading", "path", path)
					callback()
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Watcher error", "error", err)
			}
		}
	}()

	return watcher.Add(path)
}

// WatchMany watches multiple parquet files for changes
// Callback is called with the name (basename without extension) of the changed file
func WatchMany(paths []string, callback func(name string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	// Build a map from path to name
	pathToName := make(map[string]string)
	for _, path := range paths {
		base := filepath.Base(path)
		name := base[:len(base)-len(filepath.Ext(base))]
		pathToName[path] = name
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&fsnotify.Write == fsnotify.Write {
					name := pathToName[event.Name]
					if name == "" {
						name = filepath.Base(event.Name)
					}
					slog.Info("Parquet file modified, reloading", "path", event.Name, "name", name)
					callback(name)
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Watcher error", "error", err)
			}
		}
	}()

	// Add all paths to the watcher
	for _, path := range paths {
		if err := watcher.Add(path); err != nil {
			return err
		}
	}

	return nil
}
