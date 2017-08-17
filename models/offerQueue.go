package models

import (
	"container/heap"

	"bloodtales/data"
)

type StoreItemHeap []data.DataId

type OfferQueue struct {
	Heap 		StoreItemHeap 		`bson:"hp"`
}

// implementing heap interface on encapsulated heap
func (h StoreItemHeap) Len() int           { return len(h) }
func (h StoreItemHeap) Less(i, j int) bool { return int(data.GetStoreItemData(h[i]).Priority) > int(data.GetStoreItemData(h[j]).Priority) }
func (h StoreItemHeap) Swap(i, j int)      { h[i], h[j] = h[j], h[i] }

func (h *StoreItemHeap) Push(x interface{}) {
	*h = append(*h, x.(data.DataId))
}

func (h *StoreItemHeap) Pop() interface{} {
	old := *h
	n := len(old)
	x := old[n-1]
	*h = old[0 : n-1]
	return x
}
// end heap interface implementation

func (q *OfferQueue) Push(x data.DataId) {
	heap.Push(&q.Heap, x)
}

func (q *OfferQueue) Pop() data.DataId {
	return heap.Pop(&q.Heap).(data.DataId)
}

func (q *OfferQueue) Contains(id data.DataId) bool {
	for _, element := range q.Heap {
		if id == data.DataId(element) {
			return true
		}
	}

	return false
}

func (q *OfferQueue) IsEmpty() bool {
	return len(q.Heap) == 0
}