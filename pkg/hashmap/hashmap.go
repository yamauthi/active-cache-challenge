package hashmap

import "hash/maphash"

const DefaultTableSize = 10

// HashMap is a basic hashmap implementation
//
// values will be the type of `V` (any)
type HashMap[V any] struct {
	data [DefaultTableSize][]*entry[V]
	hash maphash.Hash
}

// entry represents a hashmap key value entry
type entry[V any] struct {
	HashKey uint64
	Key     []byte
	Value   V
}

// Delete removes the entry with key `key` if exists
func (h *HashMap[V]) Delete(key []byte) {
	h.resetAndWriteHash(key)
	for i, v := range h.data[(h.hash.Sum64() % DefaultTableSize)] {
		if h.hash.Sum64() == v.HashKey {
			// Remove element
			h.data[(h.hash.Sum64() % DefaultTableSize)] = append(
				h.data[(h.hash.Sum64() % DefaultTableSize)][:i],
				h.data[(h.hash.Sum64() % DefaultTableSize)][i+1:]...,
			)
			return
		}
	}
}

// Get returns the value stored using `key`.
//
// returns value of type `V` and `true` if key exists
//
// otherwise return empty `V` and `false`
func (h *HashMap[V]) Get(key []byte) (V, bool) {
	h.resetAndWriteHash(key)
	for _, v := range h.data[(h.hash.Sum64() % DefaultTableSize)] {
		if h.hash.Sum64() == v.HashKey {
			return v.Value, true
		}
	}
	return *new(V), false
}

// GetAll returns all stored keys as an array of `V`.
//
// returns nil if no values are found
func (h *HashMap[V]) GetAll() []entry[V] {
	var values []entry[V]
	for _, entries := range h.data {
		for _, e := range entries {
			values = append(values, *e)
		}
	}

	if len(values) > 0 {
		return values
	}

	return nil
}

// Put stores `value` into hashmap with specified `key`
func (h *HashMap[V]) Put(key []byte, value V) {
	h.resetAndWriteHash(key)
	for _, v := range h.data[(h.hash.Sum64() % DefaultTableSize)] {
		if h.hash.Sum64() == v.HashKey {
			v.Value = value
			return
		}
	}

	h.data[(h.hash.Sum64() % DefaultTableSize)] = append(
		h.data[(h.hash.Sum64()%DefaultTableSize)],
		&entry[V]{
			HashKey: h.hash.Sum64(),
			Key:     key,
			Value:   value,
		},
	)
}

// resetAndWriteHash reset the hash bytes and write new ones
func (h *HashMap[V]) resetAndWriteHash(k []byte) {
	h.hash.Reset()
	h.hash.Write(k)
}
