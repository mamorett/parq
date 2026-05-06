package watcher

import (
	"log/slog"
	"path/filepath"
	"sync"
	"time"

	"github.com/fsnotify/fsnotify"
)

type debounceEntry struct {
	timer *time.Timer
	mu    sync.Mutex
}

func (d *debounceEntry) fire(delay time.Duration, fn func()) {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.timer != nil {
		d.timer.Stop()
	}
	d.timer = time.AfterFunc(delay, fn)
}

// Watch watches a single parquet file for changes by watching its parent directory.
func Watch(path string, callback func()) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	dir := filepath.Dir(path)
	filename := filepath.Base(path)
	debounce := &debounceEntry{}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if filepath.Base(event.Name) != filename {
					continue
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) != 0 {
					slog.Info("Parquet file modified, reloading", "path", path, "op", event.Op)
					debounce.fire(500*time.Millisecond, func() {
						callback()
					})
				}
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Watcher error", "error", err)
			}
		}
	}()

	return watcher.Add(dir)
}

// WatchMany watches multiple parquet files for changes by watching their parent directories.
// Callback is called with the name (basename without extension) of the changed file.
func WatchMany(paths []string, callback func(name string)) error {
	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return err
	}

	fileToName := make(map[string]string)
	dirs := make(map[string]bool)
	debounces := make(map[string]*debounceEntry)

	for _, path := range paths {
		base := filepath.Base(path)
		name := base[:len(base)-len(filepath.Ext(base))]
		fileToName[path] = name
		dir := filepath.Dir(path)
		dirs[dir] = true
		debounces[path] = &debounceEntry{}
	}

	go func() {
		defer watcher.Close()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				if event.Op&(fsnotify.Write|fsnotify.Create|fsnotify.Rename) == 0 {
					continue
				}
				eventBase := filepath.Base(event.Name)
				eventDir := filepath.Dir(event.Name)
				eventPath := filepath.Join(eventDir, eventBase)

				name, known := fileToName[eventPath]
				if !known {
					continue
				}
				slog.Info("Parquet file modified, reloading", "path", eventPath, "name", name, "op", event.Op)
				deb := debounces[eventPath]
				capturedName := name
				deb.fire(500*time.Millisecond, func() {
					callback(capturedName)
				})
			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				slog.Error("Watcher error", "error", err)
			}
		}
	}()

	for dir := range dirs {
		if err := watcher.Add(dir); err != nil {
			return err
		}
	}

	return nil
}
