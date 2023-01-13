package esi

import "container/heap"

// assure interface compliance
var _ heap.Interface = (*PriorityQueue)(nil)

type PriorityQueue []*fetchRequest

// boilerblate implementation of the heap interface
// using the example from the documentation
// https://pkg.go.dev/container/heap#example-package-PriorityQueue
func (pq PriorityQueue) Len() int {
	return len(pq)
}

func (pq PriorityQueue) Less(i, j int) bool {
	return pq[i].Expiry.Before(pq[j].Expiry)
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

func (pq *PriorityQueue) Push(x any) {
	length := len(*pq)
	item := x.(*fetchRequest)
	item.index = length
	*pq = append(*pq, item)
}

func (pq *PriorityQueue) Pop() any {
	old := *pq
	length := len(old)
	item := old[length-1]
	old[length-1] = nil
	item.index = -1
	*pq = old[0 : length-1]
	return item
}
