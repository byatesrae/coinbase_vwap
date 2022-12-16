// package slidingslice provides an implementation of the Sliding Window technique.
package slidingslice

// SlidingSlice acts as a Sliding Window of values.
//
// The zero-value of this type has no capacity, and therefore no utility.
type SlidingSlice[T any] struct {
	values     []T
	len        int
	startIndex int
	lastIndex  int
}

// New creates a new SlidingSlice.
func New[T any](capacity int) *SlidingSlice[T] {
	return &SlidingSlice[T]{
		values: make([]T, capacity), // allocate for full capacity up front.
	}
}

// Len returns the length of the SlidingSlice.
func (s *SlidingSlice[T]) Len() int {
	return s.len
}

// Cap returns the capacity of the SlidingSlice.
func (s *SlidingSlice[T]) Cap() int {
	return cap(s.values)
}

// At returns the element at index a.
//
// Panics if a is out of range with respect to the length of the SlidingSlice.
func (s *SlidingSlice[T]) At(a int) T {
	if a >= s.len || a < 0 {
		panic("index out of range")
	}

	a = (a + s.startIndex) % s.Cap()

	return s.values[a]
}

// Push appends element v to the SlidingSlice. If the SlidingSlice is at capacity,
// the first element is removed.
func (s *SlidingSlice[T]) Push(v T) {
	if cap(s.values) == 0 {
		return
	}

	if s.len == 0 {
		s.values[0] = v
		s.len++

		return
	}

	s.lastIndex++

	if s.lastIndex >= cap(s.values) {
		s.lastIndex = 0
	}

	s.values[s.lastIndex] = v

	if s.len < s.Cap() {
		s.len++
	}

	if s.startIndex == s.lastIndex {
		s.startIndex++
	}

	if s.startIndex >= cap(s.values) {
		s.startIndex = 0
	}
}
