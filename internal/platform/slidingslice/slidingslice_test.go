package slidingslice

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSlidingSliceLen(t *testing.T) {
	t.Parallel()

	newWithPushCount := func(capacity int, count int) *SlidingSlice[int] {
		s := New[int](capacity)

		for a := 0; a < count; a++ {
			s.Push(a)
		}

		return s
	}

	for _, tc := range []struct {
		name     string
		with     *SlidingSlice[int]
		expected int
	}{
		{
			name:     "empty",
			with:     New[int](5),
			expected: 0,
		},
		{
			name:     "some_values",
			with:     newWithPushCount(5, 3),
			expected: 3,
		},
		{
			name:     "no_capacity",
			with:     New[int](0),
			expected: 0,
		},
		{
			name:     "zero_value",
			with:     &SlidingSlice[int]{},
			expected: 0,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.with.Len()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestSlidingSliceCap(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name     string
		with     *SlidingSlice[int]
		expected int
	}{
		{
			name:     "capacity",
			with:     New[int](5),
			expected: 5,
		},
		{
			name:     "no_capacity",
			with:     New[int](0),
			expected: 0,
		},
		{
			name:     "zero_value",
			with:     &SlidingSlice[int]{},
			expected: 0,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			actual := tc.with.Cap()
			assert.Equal(t, tc.expected, actual)
		})
	}
}

func TestSlidingSliceAt(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name          string
		with          *SlidingSlice[int]
		give          int
		expected      int
		expectedPanic string
	}{
		{
			name:     "at_first",
			with:     newWithPushes(3, []int{1, 2, 3}),
			give:     0,
			expected: 1,
		},
		{
			name:     "at_middle",
			with:     newWithPushes(3, []int{1, 2, 3}),
			give:     1,
			expected: 2,
		},
		{
			name:     "at_final",
			with:     newWithPushes(3, []int{1, 2, 3}),
			give:     2,
			expected: 3,
		},
		{
			name:          "at_before_first",
			with:          newWithPushes(3, []int{1, 2, 3}),
			give:          -1,
			expectedPanic: "index out of range",
		},
		{
			name:          "at_after_last",
			with:          newWithPushes(3, []int{1, 2}),
			give:          2,
			expectedPanic: "index out of range",
		},
		{
			name:          "zero_value",
			with:          &SlidingSlice[int]{},
			give:          0,
			expectedPanic: "index out of range",
		},
		{
			name:     "lots_of_pushes",
			with:     newWithPushes(2, []int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}),
			give:     0,
			expected: 11,
		},
		{
			name:     "at_first_when_start_is_last",
			with:     newWithPushes(3, []int{1, 2, 3, 4, 5}), // inner start index would be 2
			give:     0,
			expected: 3,
		},
		{
			name:     "at_middle_when_start_is_last",
			with:     newWithPushes(3, []int{1, 2, 3, 4, 5}), // inner start index would be 2
			give:     1,
			expected: 4,
		},
		{
			name:     "at_last_when_start_is_last",
			with:     newWithPushes(3, []int{1, 2, 3, 4, 5}), // inner start index would be 2
			give:     2,
			expected: 5,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			if tc.expectedPanic != "" {
				assert.PanicsWithValue(t, tc.expectedPanic, func() { tc.with.At(tc.give) })
			} else {
				actual := tc.with.At(tc.give)

				assert.Equal(t, tc.expected, actual)
			}
		})
	}
}

func TestSlidingSlicePush(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name string
		with *SlidingSlice[int]
		give int
	}{
		{
			name: "push_lots_1capacity",
			with: newWithPushes(1, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			give: 10,
		},
		{
			name: "push_lots_3capacity",
			with: newWithPushes(3, []int{1, 2, 3, 4, 5, 6, 7, 8, 9}),
			give: 10,
		},
		{
			name: "push_empty",
			with: New[int](5),
			give: 0,
		},
		{
			name: "push_full",
			with: newWithPushes(3, []int{1, 2, 3}),
			give: 4,
		},
		{
			name: "zero_value",
			with: &SlidingSlice[int]{},
			give: 1,
		},
	} {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert.NotPanics(t, func() { tc.with.Push(tc.give) })
		})
	}
}

func newWithPushes(capacity int, values []int) *SlidingSlice[int] {
	s := New[int](capacity)

	for _, value := range values {
		s.Push(value)
	}

	return s
}
