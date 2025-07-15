package Flow

import "sync"

// SharedState provides thread-safe data sharing between nodes in a workflow.
// It acts as a central data store that nodes can read from and write to during execution.
// All operations are protected by a read-write mutex for safe concurrent access.
//
// SharedState is typically created once per workflow execution and passed to all nodes.
// It supports storing any type of data and provides typed getter methods for convenience.
//
// Example:
//
//	state := NewSharedState()
//	state.Set("user_id", 12345)
//	state.Set("results", []string{"item1", "item2"})
//
//	userID := state.GetInt("user_id")
//	results := state.GetSlice("results")
type SharedState struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSharedState creates a new SharedState instance with an empty data map.
// The returned SharedState is ready for use and thread-safe.
//
// Example:
//
//	state := NewSharedState()
//	state.Set("key", "value")
func NewSharedState() *SharedState {
	return &SharedState{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the shared state under the specified key.
// This operation is thread-safe and will overwrite any existing value for the key.
//
// Parameters:
//   - key: The string key to store the value under
//   - value: The value to store (can be any type)
//
// Example:
//
//	state.Set("counter", 42)
//	state.Set("results", []string{"a", "b", "c"})
func (s *SharedState) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the shared state by key.
// Returns nil if the key doesn't exist.
// This operation is thread-safe for concurrent reads.
//
// Parameters:
//   - key: The string key to retrieve
//
// Returns:
//   - interface{}: The stored value, or nil if key doesn't exist
//
// Example:
//
//	value := state.Get("counter")
//	if value != nil {
//		counter := value.(int)
//	}
func (s *SharedState) Get(key string) interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.data[key]
}

// GetInt retrieves an int value, returning 0 if not found or not an int
func (s *SharedState) GetInt(key string) int {
	val := s.Get(key)
	if i, ok := val.(int); ok {
		return i
	}
	return 0
}

// GetSlice retrieves a slice value, returning empty slice if not found
func (s *SharedState) GetSlice(key string) []interface{} {
	val := s.Get(key)
	if slice, ok := val.([]interface{}); ok {
		return slice
	}
	return []interface{}{}
}

// Append adds an item to a slice in shared state
func (s *SharedState) Append(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if existing, ok := s.data[key].([]interface{}); ok {
		s.data[key] = append(existing, value)
	} else {
		s.data[key] = []interface{}{value}
	}
}
