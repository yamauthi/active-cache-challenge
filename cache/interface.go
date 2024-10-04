package cache

import (
	"time"
)

type Cache interface {
	// Set will store the key value pair with a given TTL.
	Set(key, value []byte, ttl time.Duration)

	// Get returns the value stored using `key`.
	//
	// If the key is not present value will be set to nil.
	Get(key []byte) (value []byte, ttl time.Duration)
}
