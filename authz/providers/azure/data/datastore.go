/*
Copyright The Guard Authors.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package data

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"go.kubeguard.dev/guard/authz"

	"github.com/allegro/bigcache"
	"k8s.io/klog/v2"
)

const (
	maxCacheSizeInMB = 5
	totalShards      = 128
	ttlInMins        = 3
	cleanupInMins    = 1
	maxEntrySize     = 100
	maxEntriesInWin  = 10 * 10 * 60
)

type DataStore struct {
	cache *bigcache.BigCache
}

// Set stores the given value for the given key.
// The key must not be "" and the value must not be nil.
func (s *DataStore) Set(ctx context.Context, key string, value interface{}) error {
	log := klog.FromContext(ctx)

	if key == "" || value == nil {
		return errors.New("invalid key value pair")
	}

	data, err := json.Marshal(value)
	if err != nil {
		log.V(8).InfoS("Cache set failed: marshal error", "key", key, "error", err)
		return err
	}

	err = s.cache.Set(key, data)
	if err != nil {
		log.V(8).InfoS("Cache set failed", "key", key, "error", err)
		return err
	}

	log.V(10).InfoS("Cache set successful", "key", key, "entries", s.cache.Len(), "capacity", s.cache.Capacity())
	return nil
}

// Get retrieves the Stored value for the given key.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s *DataStore) Get(ctx context.Context, key string, value interface{}) (found bool, err error) {
	log := klog.FromContext(ctx)

	if key == "" || value == nil {
		return false, errors.New("invalid key value pair")
	}

	data, err := s.cache.Get(key)
	if err != nil {
		if err == bigcache.ErrEntryNotFound {
			log.V(10).InfoS("Cache miss", "key", key, "entries", s.cache.Len())
			return false, nil
		}
		log.V(8).InfoS("Cache get error", "key", key, "error", err)
		return false, err
	}

	err = json.Unmarshal(data, value)
	if err != nil {
		log.V(8).InfoS("Cache get failed: unmarshal error", "key", key, "error", err)
		return false, err
	}

	log.V(10).InfoS("Cache hit", "key", key, "entries", s.cache.Len())
	return true, nil
}

// Delete deletes the stored value for the given key.
// The key must not be "".
func (s *DataStore) Delete(ctx context.Context, key string) error {
	log := klog.FromContext(ctx)

	if key == "" {
		return errors.New("invalid key")
	}

	err := s.cache.Delete(key)
	if err != nil {
		log.V(8).InfoS("Cache delete failed", "key", key, "error", err)
		return err
	}

	log.V(10).InfoS("Cache delete successful", "key", key, "entries", s.cache.Len())
	return nil
}

// Close closes the DataStore.
// When called, the cache is left for removal by the garbage collector.
func (s *DataStore) Close() error {
	s.logCacheStats()
	return s.cache.Close()
}

// logCacheStats logs cache statistics for debugging and insights.
func (s *DataStore) logCacheStats() {
	stats := s.cache.Stats()
	klog.InfoS("Cache statistics on close",
		"entries", s.cache.Len(),
		"capacity", s.cache.Capacity(),
		"hits", stats.Hits,
		"misses", stats.Misses,
		"deleteHits", stats.DelHits,
		"deleteMisses", stats.DelMisses,
		"collisions", stats.Collisions,
	)
}

// Options are the options for the BigCache store.
type Options struct {
	// Number of cache shards, value must be a power of two
	Shards int
	// Time after which entry can be evicted
	LifeWindow time.Duration
	// Interval between removing expired entries (clean up).
	// If set to <= 0 then no action is performed. Setting to < 1 second is counterproductive — bigcache has a one second resolution.
	CleanWindow time.Duration
	// Max number of entries in life window. Used only to calculate initial size for cache shards.
	// When proper value is set then additional memory allocation does not occur.
	MaxEntriesInWindow int
	// Max size of entry in bytes. Used only to calculate initial size for cache shards.
	MaxEntrySize int
	// StatsEnabled if true calculate the number of times a cached resource was requested.
	StatsEnabled bool
	// Verbose mode prints information about new memory allocation
	Verbose bool
	// HardMaxCacheSize is a limit for cache size in MB. Cache will not allocate more memory than this limit.
	// It can protect application from consuming all available memory on machine, therefore from running OOM Killer.
	// Default value is 0 which means unlimited size. When the limit is higher than 0 and reached then
	// the oldest entries are overridden for the new ones.
	HardMaxCacheSize int
}

// DefaultOptions is an Options object with default values.
// Bigcache provides option to give hash function however we are going with default it uses
// FNV 1a: https://en.wikipedia.org/wiki/Fowler–Noll–Vo_hash_function#FNV-1a_hash
// Key : email address/oid - Max length of email is 264 chars but 95% email length is 31
// Value: result bool
// true means access allowed
// false means access denied
// We will tweak MaxEntrySize and MaxEntriesInWindows as per requirement and testing.
var DefaultOptions = Options{
	HardMaxCacheSize:   maxCacheSizeInMB,
	Shards:             totalShards,
	LifeWindow:         ttlInMins * time.Minute,
	CleanWindow:        cleanupInMins * time.Minute,
	MaxEntriesInWindow: maxEntriesInWin,
	MaxEntrySize:       maxEntrySize,
	Verbose:            false,
}

// NewDataStore creates a BigCache store.
func NewDataStore(options Options) (authz.Store, error) {
	config := bigcache.Config{
		Shards:             options.Shards,
		LifeWindow:         options.LifeWindow,
		CleanWindow:        options.CleanWindow,
		MaxEntriesInWindow: options.MaxEntriesInWindow,
		MaxEntrySize:       options.MaxEntriesInWindow,
		Verbose:            options.Verbose,
		HardMaxCacheSize:   options.HardMaxCacheSize,
		OnRemove:           nil,
		OnRemoveWithReason: nil,
	}

	cache, err := bigcache.NewBigCache(config)
	if err != nil || cache == nil {
		return nil, err
	}

	klog.InfoS("Cache initialized",
		"shards", options.Shards,
		"lifeWindow", options.LifeWindow,
		"cleanWindow", options.CleanWindow,
		"maxCacheSizeMB", options.HardMaxCacheSize,
		"maxEntriesInWindow", options.MaxEntriesInWindow,
		"maxEntrySize", options.MaxEntrySize,
	)

	return &DataStore{cache: cache}, nil
}
