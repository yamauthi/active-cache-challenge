package cache

import (
	"bytes"
	"testing"
	"time"
)

func TestCacheEntry_emptyValueTTL(t *testing.T) {
	// Test
	val, ttl := emptyValueTTL()
	if val != nil || ttl != 0 {
		t.Errorf("wrong value for emptyValueTTL. Expected (nil, 0) but got (%s, %s)", val, ttl)
	}
}

func TestCacheEntry_GetValueTTL(t *testing.T) {
	// Setup
	expectedVal := []byte("test")
	expectedTtl := time.Second
	entry := &cacheEntry{
		Value:     expectedVal,
		Ttl:       expectedTtl,
		ExpiresAt: time.Now().Add(time.Second).UnixNano(),
	}

	nonexpiringEntry := &cacheEntry{
		Value:     expectedVal,
		Ttl:       NoExpiration,
		ExpiresAt: NoExpiration,
	}

	// Test
	val, ttl := entry.GetValueTTL()
	if !bytes.Equal(expectedVal, val) || ttl != expectedTtl {
		t.Errorf(
			"wrong value for GetValueTTL(). Expected (%s, %v) but got (%s, %v)",
			expectedVal,
			expectedTtl,
			val,
			ttl,
		)
	}

	time.Sleep(time.Second)
	val, ttl = entry.GetValueTTL() // expired value should return empty
	if val != nil || ttl != 0 {
		t.Errorf("wrong value for GetValueTTL(). Expected (nil, 0) but got (%s, %v)", val, ttl)
	}

	val, ttl = nonexpiringEntry.GetValueTTL()
	if !bytes.Equal(expectedVal, val) || ttl != NoExpiration {
		t.Errorf(
			"wrong value for GetValueTTL(). Expected (%s, %v) but got (%s, %v)",
			expectedVal,
			NoExpiration,
			val,
			ttl,
		)
	}
}

func TestCacheEntry_IsExpired(t *testing.T) {
	// Setup
	entry := &cacheEntry{
		Value:     []byte("test"),
		Ttl:       time.Second,
		ExpiresAt: time.Now().Add(time.Second).UnixNano(),
	}

	nonexpiringEntry := &cacheEntry{
		Value:     []byte("test"),
		Ttl:       NoExpiration,
		ExpiresAt: NoExpiration,
	}

	// Test
	if entry.IsExpired() {
		t.Error("wrong value for IsExpired(). Expected (false) but got (true)")
	}

	time.Sleep(time.Second)
	if !entry.IsExpired() {
		t.Error("wrong value for IsExpired(). Expected (true) but got (false)")
	}

	if nonexpiringEntry.IsExpired() {
		t.Error("wrong value for IsExpired(). Expected (false) but got (true)")
	}

}
