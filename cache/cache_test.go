package cache

import (
	"bytes"
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/yamauthi/active-cache-challenge/pkg/hashmap"
)

func TestActiveCache_defaultClean(t *testing.T) {
	// Setup
	const expiringEntries = 150
	const nonExpiringEntries = 25
	const customKeysAmountByCycle = 100

	var entries hashmap.HashMap[*cacheEntry]
	var expectedEntries, entriesLen int

	durations := [3]time.Duration{
		time.Second * 1,
		time.Second * 4,
		time.Second * 7,
	}
	defaultConf := DefaultConfig()
	conf := &Config{
		CleanerInterval:   DefaultCleanerInterval,
		KeysAmountByCycle: customKeysAmountByCycle,
	}

	for i := 0; i < nonExpiringEntries; i++ {
		entries.Put([]byte(fmt.Sprintf("key nonexp %v", i)), &cacheEntry{
			Value:     []byte(fmt.Sprintf("value %v", i)),
			Ttl:       NoExpiration,
			ExpiresAt: NoExpiration,
		})
	}

	for i := 0; i < expiringEntries; i++ {
		ttl := durations[i%len(durations)]
		entries.Put([]byte(fmt.Sprintf("key exp %v", i)), &cacheEntry{
			Value:     []byte(fmt.Sprintf("value %v", i)),
			Ttl:       ttl,
			ExpiresAt: time.Now().Add(ttl).UnixNano(),
		})
	}

	// Test
	defaultClean(&hashmap.HashMap[*cacheEntry]{}, defaultConf) // Empty entries

	defaultClean(&entries, defaultConf) // entries, no expired
	entriesLen = len(entries.GetAll())
	expectedEntries = expiringEntries + nonExpiringEntries
	if entriesLen != expectedEntries {
		t.Errorf("wrong entries amount. Expected %v but got %v", expectedEntries, entriesLen)
	}

	time.Sleep(durations[1])
	defaultClean(&entries, conf) // entries, more than half expired. Should call recursive
	entriesLen = len(entries.GetAll())
	if entriesLen >= expectedEntries {
		t.Errorf("wrong entries amount. Expected less than or equal %v but got %v", expectedEntries, entriesLen)
	}

	time.Sleep(durations[2] - durations[1])
	defaultClean(&entries, defaultConf)
	time.Sleep(DefaultCleanerInterval * time.Millisecond)
	defaultClean(&entries, conf) //only non-expiring entries
	entriesLen = len(entries.GetAll())
	expectedEntries = nonExpiringEntries
	if entriesLen != expectedEntries {
		t.Errorf("wrong entries amount. Expected %v but got %v", expectedEntries, entriesLen)
	}
}

func TestActiveCache_Get(t *testing.T) {
	// Setup
	const expiringEntries = 10
	var entries hashmap.HashMap[*cacheEntry]
	durations := [2]time.Duration{
		NoExpiration,
		time.Second,
	}

	for i := 0; i < expiringEntries; i++ {
		var expiresAt int64
		ttl := durations[i%len(durations)]
		if ttl > NoExpiration {
			expiresAt = time.Now().Add(ttl).UnixNano()
		}
		entries.Put([]byte(fmt.Sprintf("%v", i)), &cacheEntry{
			Value:     []byte(fmt.Sprintf("value %v", i)),
			Ttl:       ttl,
			ExpiresAt: expiresAt,
		})
	}

	cache := NewActiveCache()
	cache.StopCleaner()
	cache.entries = entries
	cacheEntries := entries.GetAll()
	sort.Slice(cacheEntries, func(i, j int) bool {
		return string(cacheEntries[i].Key) < string(cacheEntries[j].Key)
	})

	// Test
	var outVal []byte
	var outTTL time.Duration
	outVal, outTTL = cache.Get(nil) //get with key nil
	if outVal != nil || outTTL != 0 {
		t.Errorf("wrong value for get(nil). Expected (nil, 0) but got (%s, %v)", outVal, outTTL)
	}

	outVal, outTTL = cache.Get([]byte("nonexistent key")) //nonexistent key
	if outVal != nil || outTTL != 0 {
		t.Errorf("wrong value for get(nonexistent key). Expected (nil, 0) but got (%s, %v)", outVal, outTTL)
	}

	for _, e := range cacheEntries {
		outVal, outTTL = cache.Get(e.Key)
		if !bytes.Equal(e.Value.Value, outVal) || outTTL != e.Value.Ttl {
			t.Errorf(
				"wrong value for get(%s). Expected (%s, %s) but got (%s, %s)",
				e.Key,
				e.Value.Value,
				e.Value.Ttl,
				outVal,
				outTTL,
			)
		}
	}

	time.Sleep(time.Second)

	for i, e := range cacheEntries {
		outVal, outTTL = cache.Get(e.Key)
		if i%2 == 0 {
			// No expiring keys
			if !bytes.Equal(e.Value.Value, outVal) || outTTL != e.Value.Ttl {
				t.Errorf(
					"wrong value for get(%s). Expected (%s, %s) but got (%s, %s)",
					e.Key,
					e.Value.Value,
					e.Value.Ttl,
					outVal,
					outTTL,
				)
			}

		} else {
			// expired keys
			if outVal != nil || outTTL != 0 {
				t.Errorf(
					"wrong value for get(%s). Expected (nil, 0) but got (%s, %s)",
					e.Key,
					outVal,
					outTTL,
				)
			}
		}
	}
}

func TestActiveCache_IsCleanerRunning(t *testing.T) {
	// Setup
	cache := NewActiveCache()
	cache.StopCleaner()
	time.Sleep(time.Millisecond * 10)

	//Test
	if cache.IsCleanerRunning() != cache.isCleanerRunning.Load() {
		t.Error("wrong value on IsCleanerRunning(). Must return the same value as ActiveCache.isCleanerRunning")
	}

	cache.StartCleaner()
	time.Sleep(time.Millisecond * 10)
	if cache.IsCleanerRunning() != cache.isCleanerRunning.Load() {
		t.Error("wrong value on IsCleanerRunning(). Must return the same value as ActiveCache.isCleanerRunning")
	}
}

func TestActiveCache_performClean(t *testing.T) {
	// Setup
	var cleanExecuted bool
	cache := NewActiveCache()
	cache.StopCleaner()
	cache.cleanFunc = func(entries *hashmap.HashMap[*cacheEntry], conf *Config) {
		cleanExecuted = true
	}
	cache.StartCleaner()
	time.Sleep(time.Millisecond * 200)

	// Test
	if !cleanExecuted {
		t.Error("performClean() is not being called or is not calling ActiveCache.cleanFunc")
	}
}

func TestActiveCache_Set(t *testing.T) {
	// Setup
	cache := NewActiveCache()
	cache.StopCleaner()
	time.Sleep(time.Millisecond * 100)

	type testEntry struct {
		key   []byte
		value []byte
		ttl   time.Duration
	}

	testsCase := []testEntry{
		{key: []byte("lorem"), value: []byte("ipsum"), ttl: NoExpiration},
		{key: []byte("lorem"), value: []byte("dolor"), ttl: 10},
		{key: []byte("lorem"), value: []byte("ipsum"), ttl: 5},
		{key: []byte("jane"), value: []byte("foster"), ttl: 1},
	}

	// Test
	for _, tc := range testsCase {
		cache.Set(tc.key, tc.value, tc.ttl)
		e, ok := cache.entries.Get(tc.key)

		if !ok || !bytes.Equal(tc.value, e.Value) || tc.ttl != e.Ttl {
			t.Errorf(
				"wrong value when performing Set() for key %s. Expected (%s, %v) got (%s, %v)",
				tc.key,
				tc.value,
				tc.ttl,
				e.Value,
				e.Ttl,
			)
		}
	}

	cache.Set(nil, []byte("doe"), time.Second) // nil key
	e, ok := cache.entries.Get(nil)
	if ok || e != nil {
		t.Errorf(
			"wrong value when performing Set() for key nil. Expected to not found entry but got (%s, %v)",
			e.Value,
			e.Ttl,
		)
	}

	cache.Set([]byte("jane"), []byte("thor"), -100) // negative TTL
	e, ok = cache.entries.Get([]byte("jane"))
	if ok || e != nil {
		t.Errorf(
			"wrong value when performing Set() for key %s with negative TTL. Expected to not found entry but got (%s, %v)",
			[]byte("jane"),
			e.Value,
			e.Ttl,
		)
	}
}

func TestActiveCache_StartCleaner(t *testing.T) {
	// Setup
	var cleanExecuted bool
	conf := &Config{
		CleanerInterval: MinCleanerInterval,
	}
	cache := NewActiveCacheWithConfig(conf)
	cache.StopCleaner()
	cache.cleanFunc = func(entries *hashmap.HashMap[*cacheEntry], conf *Config) {
		cleanExecuted = true
	}

	// Test
	cache.StartCleaner()
	time.Sleep(time.Second)
	if !cache.IsCleanerRunning() || !cleanExecuted {
		t.Error("StartCleaner() is not being called or is not calling ActiveCache.performClean()")
	}

	cache.StopCleaner()
	time.Sleep(time.Second)
	if cache.isCleanerRunning.Load() {
		t.Error("StartCleaner() should stop running when ActiveCache.stopChan is closed")
	}
}

func TestActiveCache_StopCleaner(t *testing.T) {
	// Setup
	cache := NewActiveCache()

	// Test
	cache.StopCleaner()
	if cache.IsCleanerRunning() {
		t.Error("StopCleaner() is not working properly")
	}
}

func TestActiveCache_validateAndAdjustConfig(t *testing.T) {
	// Setup
	conf := &Config{
		CleanerInterval:   0,
		KeysAmountByCycle: 1,
	}
	cache := NewActiveCacheWithConfig(conf)

	if cache.config.CleanerInterval != DefaultCleanerInterval {
		t.Error("validateAndAdjustConfig shold force DefaultCleanerInterval if CleanerInterval less than MinCleanerInterval")
	}

	if cache.config.KeysAmountByCycle != DefaultKeysAmountByCycle {
		t.Error("validateAndAdjustConfig shold force DefaultKeysAmountByCycle if KeysAmountByCycle less than DefaultKeysAmountByCycle")
	}
}
