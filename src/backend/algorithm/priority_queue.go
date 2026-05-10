package algorithm

import "tucil3/src/backend/engine"

type SearchNode struct {
	State  engine.State
	GCost  int
	FCost  int
	Moves  []string
	Slides []engine.SlideResult
}

type MinHeap struct {
	nodes []*SearchNode
}

func (h *MinHeap) Len() int      { return len(h.nodes) }
func (h *MinHeap) IsEmpty() bool { return len(h.nodes) == 0 }

func (h *MinHeap) Push(node *SearchNode) {
	h.nodes = append(h.nodes, node)
	h.siftUp(len(h.nodes) - 1)
}

func (h *MinHeap) Pop() *SearchNode {
	min := h.nodes[0]
	last := len(h.nodes) - 1
	h.nodes[0] = h.nodes[last]
	h.nodes[last] = nil
	h.nodes = h.nodes[:last]
	if len(h.nodes) > 0 {
		h.siftDown(0)
	}
	return min
}

func (h *MinHeap) siftUp(i int) {
	for i > 0 {
		parent := (i - 1) / 2
		if h.nodes[parent].FCost <= h.nodes[i].FCost {
			break
		}
		h.nodes[parent], h.nodes[i] = h.nodes[i], h.nodes[parent]
		i = parent
	}
}

func (h *MinHeap) siftDown(i int) {
	n := len(h.nodes)
	for {
		smallest := i
		left, right := 2*i+1, 2*i+2
		if left < n && h.nodes[left].FCost < h.nodes[smallest].FCost {
			smallest = left
		}
		if right < n && h.nodes[right].FCost < h.nodes[smallest].FCost {
			smallest = right
		}
		if smallest == i {
			break
		}
		h.nodes[i], h.nodes[smallest] = h.nodes[smallest], h.nodes[i]
		i = smallest
	}
}
