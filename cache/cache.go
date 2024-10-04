package cache

import (
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/yamauthi/active-cache-challenge/pkg/hashmap"
)

const (
	// Cleaner
	DefaultCleanerInterval   = 200
	DefaultKeysAmountByCycle = 20

	ExpiredKeysPercentageTolerance = 25

	MinCleanerInterval   = 50
	MinKeysAmountByCycle = 5

	// Expiration
	NoExpiration = 0
)

type ActiveCache struct {
	// Function to perform clean on expired keys
	cleanFunc func(entries *hashmap.HashMap[*cacheEntry], conf *Config)

	// Holds all caching configuration
	config *Config

	// Cache entries
	entries hashmap.HashMap[*cacheEntry]

	// Reports whether the cleaner is running
	isCleanerRunning atomic.Bool

	// Mutex for read and write lock
	mtx *sync.RWMutex

	// Channel for stopping cleaner
	stopChan chan interface{}
}

// NewActiveCache returns an ActiveCache pointer instance with default config values
//
// Cleaner is started in a go routine just before return
func NewActiveCache() *ActiveCache {
	return NewActiveCacheWithConfig(nil)
}

// NewActiveCacheWithConfig returns an ActiveCache pointer instance
//
// with config from parameter or DefaultConfig if nil.
//
// Cleaner is started in a go routine just before return
func NewActiveCacheWithConfig(conf *Config) *ActiveCache {
	if conf == nil {
		conf = DefaultConfig()
	} else {
		validateAndAdjustConfig(conf)
	}

	cache := &ActiveCache{
		config:    conf,
		mtx:       &sync.RWMutex{},
		cleanFunc: defaultClean,
	}

	cache.StartCleaner()
	return cache
}

// defaultClean is the default function to perform clean algorithm that iterates through
//
// entries with TTL randomly `X` times and clean expired keys.
//
// If the percentage tolerance exceeds `ExpiredKeysPercentageTolerance`%,
//
// the function will call itself again
//
// `X` can be defined on `Config.KeysAmountByCycle`
func defaultClean(entriesMap *hashmap.HashMap[*cacheEntry], conf *Config) {
	var deleted int
	entries := entriesMap.GetAll()
	sampleSize := min(conf.KeysAmountByCycle, len(entries))

	if sampleSize == 0 {
		return
	}

	indexesToCheck := rand.Perm(len(entries))[:sampleSize]
	for _, i := range indexesToCheck {
		if entries[i].Value.IsExpired() {
			entriesMap.Delete(entries[i].Key)
			deleted++
		}
	}

	if (deleted * 100 / len(indexesToCheck)) > ExpiredKeysPercentageTolerance {
		defaultClean(entriesMap, conf)
	}
}

// Get returns Value and TTL from specified key if it exists.
//
// If key is nil OR does not exist returns (nil, 0)
func (c *ActiveCache) Get(key []byte) ([]byte, time.Duration) {
	if key == nil {
		return emptyValueTTL()
	}

	//Lock cache while reading
	c.mtx.Lock()
	defer c.mtx.Unlock()

	if entry, ok := c.entries.Get(key); ok {
		return entry.GetValueTTL()
	}

	return emptyValueTTL()
}

// IsCleanerRunning reports whether the cleaner is running
func (c *ActiveCache) IsCleanerRunning() bool {
	return c.isCleanerRunning.Load()
}

// performClean locks cache entries and perform clean function
func (c *ActiveCache) performClean() {
	c.mtx.Lock()
	defer c.mtx.Unlock()

	c.cleanFunc(&c.entries, c.config)
}

// Set sets Value for specified Key with TTL.
//
// If TTL is equal to NoExpiration (zero), then it will never expires.
//
// If TTL is negative the key expires instantly
func (c *ActiveCache) Set(key, value []byte, ttl time.Duration) {
	if key != nil {
		// Lock cache while writing
		c.mtx.Lock()
		defer c.mtx.Unlock()

		// delete key if ttl is negative
		if ttl < NoExpiration {
			c.entries.Delete(key)
			return
		}

		var expiresAt int64
		if ttl > NoExpiration {
			expiresAt = time.Now().Add(ttl).UnixNano()
		}

		c.entries.Put(key, &cacheEntry{
			Value:     value,
			Ttl:       ttl,
			ExpiresAt: expiresAt,
		})
	}
}

// StartCleaner starts active cache cleaning
func (c *ActiveCache) StartCleaner() {
	go func() {
		if !c.isCleanerRunning.Load() {
			c.stopChan = make(chan interface{})
			c.isCleanerRunning.Store(true)

			timer := time.NewTimer(time.Millisecond * time.Duration(c.config.CleanerInterval))
			for {
				select {
				case <-c.stopChan:
					c.isCleanerRunning.Store(false)
					return
				case <-timer.C:
					c.performClean()
				}
			}
		}
	}()
}

// StopCleaner stops active cache cleaning
func (c *ActiveCache) StopCleaner() {
	if c.isCleanerRunning.Load() {
		close(c.stopChan)
	}
}

// validateAndAdjustConfig validate if parameters
//
// has valid values and if not change them to default values
func validateAndAdjustConfig(conf *Config) {
	if conf.CleanerInterval < MinCleanerInterval {
		conf.CleanerInterval = DefaultCleanerInterval
	}

	if conf.KeysAmountByCycle < MinKeysAmountByCycle {
		conf.KeysAmountByCycle = DefaultKeysAmountByCycle
	}
}
