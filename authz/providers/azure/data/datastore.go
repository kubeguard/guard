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
	"encoding/json"
	"errors"
	"time"

	"github.com/appscode/guard/authz"

	"github.com/allegro/bigcache"
)

const (
	maxCacheSizeInMB = 5
	totalShards      = 128
	ttlInMins        = 5
	cleanupInMins    = 2
	maxEntrySize     = 100
	maxEntriesInWin  = 10 * 10 * 60
)

type DataStore struct {
	cache *bigcache.BigCache
}

// Set stores the given value for the given key.
// The key must not be "" and the value must not be nil.
func (s *DataStore) Set(key string, value interface{}) error {
	if key == "" || value == nil {
		return errors.New("invalid key value pair")
	}

	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return s.cache.Set(key, data)
}

// Get retrieves the Stored value for the given key.
// If no value is found it returns (false, nil).
// The key must not be "" and the pointer must not be nil.
func (s *DataStore) Get(key string, value interface{}) (found bool, err error) {
	if key == "" || value == nil {
		return false, errors.New("invalid key value pair")
	}

	data, err := s.cache.Get(key)
	if err != nil {
		return false, err
	}

	return true, json.Unmarshal(data, value)
}

// Delete deletes the stored value for the given key.
// The key must not be "".
func (s *DataStore) Delete(key string) error {
	if key == "" {
		return errors.New("invalid key")
	}

	err := s.cache.Delete(key)
	if err != nil {
		return err
	}

	return nil
}

// Close closes the DataStore.
// When called, the cache is left for removal by the garbage collector.
func (s *DataStore) Close() error {
	return s.cache.Close()
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

	return &DataStore{cache: cache}, nil
}
