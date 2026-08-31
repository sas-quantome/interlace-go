/*
 * Copyright 2026 Quantome SAS
 * SPDX-License-Identifier: Apache-2.0
 */

package runner

import (
	"cmp"
	"fmt"
	"slices"
)

// Stack is a generic type representing a LIFO (last-in, first-out) data structure.
type Stack[T any] struct {
	Items []T
}

// List is a generic type that represents a collection of elements of any type T.
type List[T any] []T

// ListOf creates and returns a List containing the provided elements of type T.
func ListOf[T any](items ...T) List[T] {
	return items
}

// Reset clears all items in the stack, bringing it to an empty state.
func (s *Stack[T]) Reset() {
	s.Clear()
}

// Clear removes all elements from the stack; keeps the underlying array capacity to avoid re-allocation.
func (s *Stack[T]) Clear() {
	s.Items = s.Items[:0]
}

// Push adds the specified item to the top of the stack.
// It intercepts empty strings to prevent memory leaks from retained backing arrays.
func (s *Stack[T]) Push(item T) {
	// 1. Check if the generic item is a string and if it is empty
	if str, ok := any(item).(string); ok && str == "" {

		// 2. Create an absolute zero-value of type T (for strings, this is a clean "")
		var zero T

		// 3. Append the clean zero-value instead of the memory-heavy substring
		s.Items = append(s.Items, zero)
		return
	}

	// For all other types or non-empty strings, append normally
	s.Items = append(s.Items, item)
}

// PushOnce inserts an item into a sorted slice if it doesn't already exist, maintaining the slice's sorted order.
func PushOnce[T cmp.Ordered](items []T, item T) []T {
	i, found := slices.BinarySearch(items, item)
	if found == false {
		items = append(items, *new(T))
		copy(items[i+1:], items[i:])
		items[i] = item
	}
	return items
}

// Pop removes and returns the top item from the stack. If the stack is empty, it returns the zero value of T and false.
func (s *Stack[T]) Pop() (T, bool) {
	if len(s.Items) == 0 {
		var zero T
		return zero, false
	}
	item := s.Items[len(s.Items)-1]
	s.Items = s.Items[:len(s.Items)-1]
	return item, true
}

// Len returns the number of elements currently present in the stack.
func (s *Stack[T]) Len() int {
	return len(s.Items)
}

// Empty returns true if the stack has no elements, otherwise it returns false.
func (s *Stack[T]) Empty() bool {
	return len(s.Items) == 0
}

// Peek retrieves the top element of the stack without removing it.
func (s *Stack[T]) Peek() T {
	return s.Items[len(s.Items)-1]
}

// String returns a string representation of the stack's items in slice format.
func (s *Stack[T]) String() string {
	return fmt.Sprintf("%v", s.Items)
}

// Copy creates and returns a new Stack instance containing a duplicate of the current stack's items.
func (s *Stack[T]) Copy() Stack[T] {
	return Stack[T]{Items: append([]T{}, s.Items...)}
}

// Reverse reverses the order of the elements in the stack.
func (s *Stack[T]) Reverse() {
	for i, j := 0, len(s.Items)-1; i < j; i, j = i+1, j-1 {
		s.Items[i], s.Items[j] = s.Items[j], s.Items[i]
	}
}
