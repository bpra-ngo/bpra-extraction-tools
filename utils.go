package main

import (
	"errors"
	"slices"
)

// ValueSortedMaps
var ErrKeyNotFound = errors.New("key not found")
var ErrValueNotFound = errors.New("value not found")
var ErrOutOfBounds = errors.New("index out of bounds")

type comparator[V any] func(x, y V) int

// Sorted Map according to a comparator[V any] function
// Sorts only in increasing order
// dict holds a key::index map, where the index is the i
// of the element in the ordered slice xs
type ValueSortedMap[K comparable, V comparable] struct {
	// sample fields, subject to change
	dict map[K]int
	xs   []V
	// [x < y] < 0
	// [x > y] > 0
	// [x == y] = 0
	comp comparator[V]
}

func NewValueSortedMap[K comparable, V comparable](comp comparator[V]) *ValueSortedMap[K, V] {
	return &ValueSortedMap[K, V]{
		dict: make(map[K]int),
		xs:   make([]V, 0),
		comp: comp,
	}
}

func (m *ValueSortedMap[K, V]) Copy() *ValueSortedMap[K, V] {
	var copy ValueSortedMap[K, V]
	for k, v := range m.dict {
		copy.dict[k] = v
	}
	for _, v := range m.xs {
		copy.xs = append(m.xs, v)
	}
	return &copy
}

func (m *ValueSortedMap[K, V]) Insert(key K, val V) *ValueSortedMap[K, V] {
	if m.Len() == 0 {
		m.dict[key] = 0
		m.xs = append(m.xs, val)
	} else {
		below := m.ValuesBelow(val, true)
		above := m.ValuesAbove(val, false)
		m.xs = make([]V, 0)
		m.xs = append(m.xs, below...)
		m.xs = append(m.xs, val)
		m.xs = append(m.xs, above...)
		insertIndex := len(below)
		//push all other elements above by one
		for k, v := range m.dict {
			if v >= insertIndex {
				m.dict[k] = v + 1
			}
		}
		m.dict[key] = insertIndex
	}
	return m
}

func (m *ValueSortedMap[K, V]) Len() int {
	return len(m.dict)
}

func (m *ValueSortedMap[K, V]) HasKey(key K) bool {
	_, ok := m.dict[key]
	return ok
}

func (m *ValueSortedMap[K, V]) Get(key K) (*V, error) {
	if !m.HasKey(key) {
		return nil, ErrKeyNotFound
	}
	return &m.xs[m.dict[key]], nil
}

func (m *ValueSortedMap[K, V]) Remove(key K) (*V, error) {
	var out *V
	if !m.HasKey(key) {
		return out, ErrKeyNotFound
	}
	out, _ = m.Get(key)
	delete(m.dict, key)
	s := slices.Index(m.xs, *out)
	if s < 0 {
		return out, ErrValueNotFound
	}
	m.xs = append(m.xs[:s], m.xs[s+1:]...)
	return out, nil
}

// returns the (key, value) at position `idx` in the sorted (value)
// list (according to m.comp)
func (m *ValueSortedMap[K, V]) PeekItemByPos(idx int) (*V, *K, error) {
	if idx < 0 || idx > m.Len() {
		return nil, nil, ErrOutOfBounds
	}
	var k *K
	for key, i := range m.dict {
		if idx == i {
			k = &key
			break
		}
	}
	return &m.xs[idx], k, nil
}

// if `k_1` precedes `k_2` in m.Keys(), then m.comp(m.Get(k_1),m.Get(k_2))) <= 0
func (m *ValueSortedMap[K, V]) Keys() []K {
	keys := make([]K, len(m.dict))
	for k, i := range m.dict {
		keys[i] = k
	}
	return keys
}

// Values() should be sorted, ideally this call should not force any
// computation
func (m *ValueSortedMap[K, V]) Values() []V {
	return m.xs
}

// checks if new val V is between two old values
func (m *ValueSortedMap[K, V]) isBetween(val V, x V, y V) bool {
	if m.isGreaterThan(val, x) && m.isLessThan(val, y) {
		return true
	} else {
		return false
	}
}

// is x greater than y (strict)
func (m *ValueSortedMap[K, V]) isGreaterThan(x V, y V) bool {
	return m.comp(x, y) > 0
}

// is x less than y (strict)
func (m *ValueSortedMap[K, V]) isLessThan(x V, y V) bool {
	return m.comp(x, y) < 0
}

// returns the sorted values greater than or equal to `val`.  if `val`
// is present in `m.Values()`, whether or not it is included in
// `m.ValuesAbove()` is determined by `including` ideally, shouldn't
// force too much computation
// panics if val V not in m.xs
func (m *ValueSortedMap[K, V]) ValuesAbove(val V, including bool) []V {
	start := slices.Index(m.xs, val)
	if start < 0 {
		for i := 0; i < len(m.xs); i++ {
			curr := m.xs[i]
			if m.isLessThan(val, curr) {
				start = i
				break
			}
			if i == len(m.xs)-1 {
				//last element clause
				if m.isGreaterThan(val, curr) {
					start = i + 1
					break
				}
				continue
			}
		}
	} else if !including {
		for i := start; i < len(m.xs); i++ {
			if m.xs[i] == val {
				start++
			} else {
				break
			}
		}
	}
	return m.xs[start:]
}

// returns the sorted values less than or equal to `val`.  if `val` is
// present in `m.Values()`, whether or not it is included in
// `m.ValuesBelow()` is determined by `including` ideally, shouldn't
// force too much computation
// panics if val V not in m.xs
func (m *ValueSortedMap[K, V]) ValuesBelow(val V, including bool) []V {
	end := slices.Index(m.xs, val)
	if end < 0 {
		for i := 0; i < len(m.xs); i++ {
			curr := m.xs[i]
			if m.isLessThan(val, curr) {
				end = i
				break
			}
			if i == len(m.xs)-1 {
				//last element clause
				if m.isGreaterThan(val, curr) {
					end = i + 1
					break
				}
				continue
			}
		}
	} else if including {
		for i := end; i < len(m.xs); i++ {
			if m.xs[i] == val {
				end++
			} else {
				break
			}
		}
	}
	return m.xs[0:end]
}
