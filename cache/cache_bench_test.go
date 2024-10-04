package cache

import (
	"fmt"
	"testing"
	"time"
)

const BenchmarkEntries = 100

func BenchmarkActiveCache_Get(b *testing.B) {
	// Setup
	cache := NewActiveCacheWithConfig(&Config{
		CleanerInterval: 1, // invalid, should adjust to DefaultCleanerInterval
	})
	cache.StopCleaner()
	durations := []time.Duration{
		time.Millisecond,
		time.Second * 1,
		time.Second * 5,
		time.Second * 10,
		NoExpiration,
	}

	for i := 0; i < BenchmarkEntries; i++ {
		cache.Set(
			[]byte(fmt.Sprintf("key%v", i)),
			[]byte(fmt.Sprintf("value%v", i)),
			durations[i%len(durations)],
		)
	}
	cache.StartCleaner()
	b.ResetTimer()

	// Test
	for n := 0; n < b.N; n++ {
		cache.Get([]byte(fmt.Sprintf("key%v", n%10)))
		cache.Get([]byte("nonexistent key"))
		cache.Get(nil)
	}

	b.ReportAllocs()
}

func BenchmarkActiveCache_Set(b *testing.B) {
	// Setup
	cache := NewActiveCache()
	type input struct {
		key   []byte
		value []byte
	}
	durations := []time.Duration{
		time.Millisecond,
		time.Second * -1,
		time.Second * 5,
		time.Second * 10,
		NoExpiration,
	}
	var entries []input
	for i := 0; i < BenchmarkEntries; i++ {
		entries = append(
			entries,
			input{
				key:   []byte(fmt.Sprintf("key%v", i)),
				value: []byte(fmt.Sprintf("value%v", i)),
			},
		)
	}
	entries[BenchmarkEntries-1] = input{
		key:   nil,
		value: []byte("value"),
	}

	b.ResetTimer()
	// Test
	for n := 0; n < b.N; n++ {
		cache.Set(
			entries[(b.N%BenchmarkEntries)].key,
			entries[(b.N%BenchmarkEntries)].value,
			durations[b.N%len(durations)],
		)
	}

	b.ReportAllocs()
}
