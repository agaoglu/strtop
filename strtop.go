package main

import (
	"bufio"
	"container/heap"
	"flag"
	"fmt"
	"os"
	"sync"
	"time"
)

var max = flag.Int("max", 10, "output this much items")
var interval = flag.Duration(
	"interval", time.Duration(1)*time.Second,
	"output with this interval")

// An Item is something we manage in a priority queue.
type Item struct {
	value    string // The value of the item; arbitrary.
	priority int    // The priority of the item in the queue.
	// The index is needed by update and is maintained by the heap.Interface methods.
	index int // The index of the item in the heap.
}

// A PriorityQueue implements heap.Interface and holds Items.
type PriorityQueue []*Item

// Len returns the length of the PriorityQueue
func (pq PriorityQueue) Len() int { return len(pq) }

func (pq PriorityQueue) Less(i, j int) bool {
	// We want Pop to give us the highest, not lowest, priority so we use greater than here.
	return pq[i].priority > pq[j].priority
}

func (pq PriorityQueue) Swap(i, j int) {
	pq[i], pq[j] = pq[j], pq[i]
	pq[i].index = i
	pq[j].index = j
}

// Push adds the item to PriorityQueue
func (pq *PriorityQueue) Push(x interface{}) {
	n := len(*pq)
	item := x.(*Item)
	item.index = n
	*pq = append(*pq, item)
}

// Pop removes and returns the first item in the PriorityQueue
func (pq *PriorityQueue) Pop() interface{} {
	old := *pq
	n := len(old)
	item := old[n-1]
	item.index = -1 // for safety
	*pq = old[0 : n-1]
	return item
}

// update modifies the priority and value of an Item in the queue.
func (pq *PriorityQueue) update(item *Item, value string, priority int) {
	item.value = value
	item.priority = priority
	heap.Fix(pq, item.index)
}

func output(pq *PriorityQueue, lock *sync.Mutex, quit chan int) {
	for {
		select {
		case <-quit:
			return
		default:
			time.Sleep(*interval)
		}

		lock.Lock()
		items := make([]*Item, 0, *max)
		size := pq.Len()
		for i := 0; i < *max && i < size; i++ {
			it := heap.Pop(pq).(*Item)
			items = append(items, it)
			fmt.Printf("%7d\t%v\n", it.priority, it.value)
		}
		for _, it := range items {
			heap.Push(pq, it)
		}
		fmt.Printf("Total %d items\n", size)
		lock.Unlock()
	}
}

func main() {
	flag.Parse()

	pq := new(PriorityQueue)
	lock := new(sync.Mutex)
	quit := make(chan int)

	heap.Init(pq)
	go output(pq, lock, quit)

	counts := make(map[string]*Item, 0)
	scanner := bufio.NewScanner(os.Stdin)
	for scanner.Scan() {
		value := scanner.Text()
		item := counts[value]
		lock.Lock()
		if item == nil {
			item = &Item{value: value, priority: 1}
			heap.Push(pq, item)
			counts[value] = item
		} else {
			pq.update(item, item.value, item.priority+1)
		}
		lock.Unlock()
	}
	quit <- 1
}
