package Flow

import "sync"

// SharedState represents the shared state passed between nodes (like PocketFlow)
type SharedState struct {
	data map[string]interface{}
	mu   sync.RWMutex
}

// NewSharedState creates a new shared state instance
func NewSharedState() *SharedState {
	return &SharedState{
		data: make(map[string]interface{}),
	}
}

// Set stores a value in the shared state
func (s *SharedState) Set(key string, value interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = value
}

// Get retrieves a value from the shared state
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