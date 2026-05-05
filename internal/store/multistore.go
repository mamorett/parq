package store

import (
	"fmt"
	"log/slog"
	"sync"

	"github.com/trithemius/parq/internal/config"
)

// MultiStore holds multiple MemoryStore instances, one per parquet file
type MultiStore struct {
	mu     sync.RWMutex
	stores map[string]*MemoryStore // key = parquet name
	cfg    *config.MultiConfig
}

// NewMultiStore creates a MultiStore from a MultiConfig
func NewMultiStore(mc *config.MultiConfig) (*MultiStore, error) {
	ms := &MultiStore{
		stores: make(map[string]*MemoryStore),
		cfg:    mc,
	}

	// Create and initialize each store
	for i := range mc.Parquets {
		entry := &mc.Parquets[i]
		name := entry.GetName()

		// Resolve the entry to get a Config
		cfg, err := entry.Resolve()
		if err != nil {
			return nil, fmt.Errorf("failed to resolve %s: %w", name, err)
		}

		store, err := NewMemoryStore(cfg)
		if err != nil {
			return nil, fmt.Errorf("failed to create store for %s: %w", name, err)
		}

		ms.stores[name] = store
		slog.Info("Loaded parquet store", "name", name, "path", entry.Path)
	}

	return ms, nil
}

// StoreFor returns the MemoryStore for a given parquet name
func (ms *MultiStore) StoreFor(name string) (*MemoryStore, error) {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	store, ok := ms.stores[name]
	if !ok {
		return nil, fmt.Errorf("unknown parquet: %s (available: %v)", name, ms.StoreNames())
	}
	return store, nil
}

// StoreForDefault returns the first store if no name is provided
func (ms *MultiStore) StoreForDefault(name string) (*MemoryStore, error) {
	if name == "" {
		ms.mu.RLock()
		defer ms.mu.RUnlock()
		// Return first store
		for _, store := range ms.stores {
			return store, nil
		}
		return nil, fmt.Errorf("no stores available")
	}
	return ms.StoreFor(name)
}

// StoreNames returns all available parquet names
func (ms *MultiStore) StoreNames() []string {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	names := make([]string, 0, len(ms.stores))
	for name := range ms.stores {
		names = append(names, name)
	}
	return names
}

// Reload reloads a specific parquet store by name
func (ms *MultiStore) Reload(name string) error {
	ms.mu.RLock()
	store, ok := ms.stores[name]
	ms.mu.RUnlock()

	if !ok {
		return fmt.Errorf("unknown parquet: %s", name)
	}

	slog.Info("Reloading parquet store", "name", name)
	return store.Reload()
}

// GetConfig returns the config for a specific parquet
func (ms *MultiStore) GetConfig(name string) (*config.Config, error) {
	ms.mu.RLock()
	store, ok := ms.stores[name]
	ms.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown parquet: %s", name)
	}
	return store.GetConfig(), nil
}

// GetMultiConfig returns the full MultiConfig
func (ms *MultiStore) GetMultiConfig() *config.MultiConfig {
	return ms.cfg
}

// Close closes all stores
func (ms *MultiStore) Close() error {
	ms.mu.RLock()
	defer ms.mu.RUnlock()

	for name, store := range ms.stores {
		if err := store.Close(); err != nil {
			slog.Error("Error closing store", "name", name, "error", err)
		}
	}
	return nil
}
