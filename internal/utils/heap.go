package utils

import "container/heap"

// Scored is an interface for items that can be ranked by score
type Scored interface {
	GetScore() float64
}

// minHeap implements heap.Interface for any Scored type
type minHeap[T Scored] []T

func (h minHeap[T]) Len() int           { return len(h) }
func (h minHeap[T]) Less(i, j int) bool { return h[i].GetScore() < h[j].GetScore() }
func (h minHeap[T]) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *minHeap[T]) Push(x any) {
	*h = append(*h, x.(T))
}

func (h *minHeap[T]) Pop() any {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}

// TopK returns the top K items with highest scores from the input slice.
// Results are returned in descending order (highest score first).
// Time complexity: O(n log k), Space complexity: O(k)
func TopK[T Scored](items []T, k int) []T {
	if k <= 0 || len(items) == 0 {
		return []T{}
	}

	h := &minHeap[T]{}
	heap.Init(h)

	for _, item := range items {
		if h.Len() < k {
			heap.Push(h, item)
		} else if item.GetScore() > (*h)[0].GetScore() {
			heap.Pop(h)
			heap.Push(h, item)
		}
	}

	// Extract in descending order
	result := make([]T, h.Len())
	for i := len(result) - 1; i >= 0; i-- {
		result[i] = heap.Pop(h).(T)
	}

	return result
}
