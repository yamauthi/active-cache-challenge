package cache

import "time"

// A cacheEntry represents an entry with Value and TTL
type cacheEntry struct {
	// Entry value
	Value []byte

	// Entry duration time
	Ttl time.Duration

	// Expiration time in nanoseconds
	ExpiresAt int64
}

// emptyValueTTL returns a nil value and time duration 0
func emptyValueTTL() ([]byte, time.Duration) {
	return nil, 0
}

// GetValueTTL returns the value and TTL
func (c *cacheEntry) GetValueTTL() ([]byte, time.Duration) {
	if c.IsExpired() {
		return emptyValueTTL()
	}

	return c.Value, c.Ttl
}

// IsExpired reports whether the cache entry is expired or not
func (c *cacheEntry) IsExpired() bool {
	return NoExpiration != c.ExpiresAt && time.Now().UnixNano() >= c.ExpiresAt
}
