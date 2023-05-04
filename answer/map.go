package answer

import (
	"github.com/emirpasic/gods/maps/linkedhashmap"
)

//////
// Const, var, and types.
//////

// Map is a map of questions.
//
// NOTE: Golang map is an unordered collection, it does not preserve the order
// of keys.
type Map struct {
	*linkedhashmap.Map
}

//////
// Methods.
//////

// Store a question in the map.
func (m *Map) Store(key string, q IAnswer) *Map {
	m.Map.Put(key, q)

	return m
}

// Load a question from the map.
func (m *Map) Load(key string) IAnswer {
	v, ok := m.Map.Get(key)
	if !ok {
		return nil
	}

	q, ok := v.(IAnswer)
	if !ok {
		return nil
	}

	return q
}

// LoadByIndex loads a question from the map by index.
func (m *Map) LoadByIndex(i int) IAnswer {
	j := 0

	// Iterate over.
	for _, v := range m.Map.Values() {
		if j == i {
			q, ok := v.(IAnswer)
			if !ok {
				return nil
			}

			return q
		}

		j++
	}

	return nil
}

// Delete a question from the map.
func (m *Map) Delete(key string) *Map {
	m.Map.Remove(key)

	return m
}

// Size returns the size of the map.
func (m *Map) Size() int {
	return m.Map.Size()
}

// GetIndex returns the index of the question in the map.
func (m *Map) GetIndex(key string) int {
	i := 0

	for _, k := range m.Map.Keys() {
		if k == key {
			return i
		}

		i++
	}

	return -1
}

//////
// Factory.
//////

// NewMap returns a new map of questions.
func NewMap() *Map {
	return &Map{
		linkedhashmap.New(),
	}
}
