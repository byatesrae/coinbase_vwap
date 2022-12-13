package slidingslice

type SlidingSlice[T any] struct {
	values     []T
	len        int
	startIndex int
	lastIndex  int
}

func New[T any](capacity int) *SlidingSlice[T] {
	return &SlidingSlice[T]{
		values: make([]T, capacity),
	}
}

func (s *SlidingSlice[T]) Len() int {
	return s.len
}

func (s *SlidingSlice[T]) Cap() int {
	return cap(s.values)
}

func (s *SlidingSlice[T]) At(a int) T {
	if a >= s.len {
		panic("index outside of range")
	}

	a = (a + s.startIndex) % (s.Cap() - 1)

	return s.values[a]
}

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
