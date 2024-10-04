package hashmap

import (
	"bytes"
	"hash/maphash"
	"reflect"
	"sort"
	"testing"
)

const BenchmarkEntries = 10

var hashmap HashMap[[]byte]

func TestHashMap_Delete(t *testing.T) {
	hashmap = HashMap[[]byte]{}
	key := []byte("lorem")
	val := []byte("ipsum")
	hashTest := maphash.Hash{}
	hashTest.SetSeed(hashmap.hash.Seed())
	hashTest.Write(key)

	hashmap.data[(hashTest.Sum64() % DefaultTableSize)] = append(
		hashmap.data[(hashTest.Sum64()%DefaultTableSize)],
		&entry[[]byte]{
			HashKey: hashTest.Sum64(),
			Key:     key,
			Value:   val,
		},
	)

	hashmap.Delete(key)

	if len(hashmap.data[(hashTest.Sum64()%DefaultTableSize)]) != 0 {
		t.Error("key was now deleted")
	}
}

func TestHashMap_Get(t *testing.T) {
	hashmap = HashMap[[]byte]{}
	hashTest := maphash.Hash{}
	hashTest.SetSeed(hashmap.hash.Seed())

	type testEntry struct {
		key   []byte
		value []byte
	}

	// Test existing keys
	existingCases := []testEntry{
		{key: []byte("key"), value: []byte("value")},
		{key: []byte("lorem"), value: []byte("ipsum")},
		{key: []byte("john"), value: []byte("doe")},
	}

	for _, ec := range existingCases {
		hashTest.Reset()
		hashTest.Write(ec.key)
		hashmap.data[(hashTest.Sum64() % DefaultTableSize)] = append(
			hashmap.data[(hashTest.Sum64()%DefaultTableSize)],
			&entry[[]byte]{
				HashKey: hashTest.Sum64(),
				Key:     ec.key,
				Value:   ec.value,
			},
		)

		out, ok := hashmap.Get(ec.key)
		if !ok || !bytes.Equal(ec.value, out) {
			t.Errorf(
				"Wrong value for key %v. Expected %s, but received %s",
				ec.key,
				ec.value,
				out,
			)
		}
	}

	// Test unexisting keys
	unexistingCases := [][]byte{
		[]byte("unexisting key"),
		[]byte("valpsum"),
		[]byte("ipdoe"),
	}

	for _, u := range unexistingCases {
		_, ok := hashmap.Get(u)
		if ok {
			t.Errorf("Key %s not expected to be found", u)
		}
	}
}

func TestHashMap_GetAll(t *testing.T) {
	// Setup
	hashmap = HashMap[[]byte]{}
	hashTest := maphash.Hash{}
	hashTest.SetSeed(hashmap.hash.Seed())

	type testEntry struct {
		key   []byte
		value []byte
	}

	entries := []testEntry{
		{key: []byte("john"), value: []byte("doe")},
		{key: []byte("key"), value: []byte("value")},
		{key: []byte("lorem"), value: []byte("ipsum")},
	}

	expected := []entry[[]byte]{}

	for _, ec := range entries {
		hashTest.Reset()
		hashTest.Write(ec.key)
		hashmap.data[(hashTest.Sum64() % DefaultTableSize)] = append(
			hashmap.data[(hashTest.Sum64()%DefaultTableSize)],
			&entry[[]byte]{
				HashKey: hashTest.Sum64(),
				Key:     ec.key,
				Value:   ec.value,
			},
		)

		expected = append(expected, entry[[]byte]{
			HashKey: hashTest.Sum64(),
			Key:     ec.key,
			Value:   ec.value,
		})
	}

	// Test hashmap with values
	out := hashmap.GetAll()
	sort.Slice(out, func(i, j int) bool {
		return string(out[i].Key) < string(out[j].Key)
	})

	if !reflect.DeepEqual(expected, out) {
		t.Errorf(
			"Wrong value on HashMap.GetAll. Expected %v, but received %v",
			expected,
			out,
		)
	}

	// Test empty hashmap
	hashmap = HashMap[[]byte]{}
	if hashmap.GetAll() != nil {
		t.Errorf(
			"Wrong value on HashMap.GetAll. Expected %v, but received %v",
			expected,
			out,
		)
	}
}

func TestHashMap_Put(t *testing.T) {
	hashmap = HashMap[[]byte]{}
	hashTest := maphash.Hash{}
	hashTest.SetSeed(hashmap.hash.Seed())

	type testEntry struct {
		key   []byte
		value []byte
	}

	testsCase := []testEntry{
		// Simple put test
		{key: []byte("key"), value: []byte("value")},
		{key: []byte("lorem"), value: []byte("ipsum")},
		{key: []byte("john"), value: []byte("doe")},

		// Overwrite value on key test
		{key: []byte("key"), value: []byte("value2")},
		{key: []byte("lorem"), value: []byte("ipsum2")},
	}

	getEntryByKey := func(key uint64, entries []*entry[[]byte]) *entry[[]byte] {
		for _, e := range entries {
			if key == e.HashKey {
				return e
			}
		}
		return nil
	}

	for _, tc := range testsCase {
		hashmap.Put(tc.key, tc.value)

		// prepare to calculate key hash
		hashTest.Reset()
		hashTest.Write(tc.key)

		entries := hashmap.data[(hashTest.Sum64() % DefaultTableSize)]
		entry := getEntryByKey(hashTest.Sum64(), entries)

		if entry != nil {
			if entry.Value != nil && !bytes.Equal(tc.value, entry.Value) {
				t.Errorf(
					"Wrong value for key %v. Expected %s, but received %s",
					tc.key,
					tc.value,
					entry.Value,
				)
			}
			continue
		}

		t.Errorf("Expected existing pointer to entry with key %v and value %s", tc.key, tc.value)
	}
}
