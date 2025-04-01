package utils

// Set is a generic type alias for a set implemented using map[T]struct{}
type Set[T comparable] map[T]struct{}

// NewSet creates a new Set from a list of values
func NewSet[T comparable](values ...T) Set[T] {
	s := make(Set[T])
	for _, v := range values {
		s[v] = struct{}{}
	}
	return s
}

// Add inserts an element into the set
func (s Set[T]) Add(value T) {
	s[value] = struct{}{}
}

// Remove deletes an element from the set
func (s Set[T]) Remove(value T) {
	delete(s, value)
}

// Contains checks if an element exists in the set
func (s Set[T]) Contains(value T) bool {
	_, exists := s[value]
	return exists
}

// Values returns all elements of the set as a slice
func (s Set[T]) Values() []T {
	result := make([]T, 0, len(s))
	for key := range s {
		result = append(result, key)
	}
	return result
}
