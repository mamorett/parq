package watcher

import (
	"log/slog"

	"github.com/fsnotify/fsnotify"
)

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
