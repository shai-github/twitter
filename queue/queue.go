package queue

import (
	"sync/atomic"
	"unsafe"
)

// Represents an unbounded lock-free queue
type Queue interface {
	Enq(input *Request)
	Deq() *Request
	Emp() bool
}

// Struct to represent request
type Request struct {
	Command   string
	Id        int32
	Body      string
	Timestamp float64
}

// Represents a node in the lock-free queue
type queueNode struct {
	Next  unsafe.Pointer
	Input *Request
}

// Feed of unbounded lock-free queue nodes
type queueFeed struct {
	head unsafe.Pointer
	tail unsafe.Pointer
}

// Creates a new lock-free queue node
func MakeNode(input *Request, next unsafe.Pointer) *queueNode {
	thisNode := queueNode{}
	thisNode.Next = next
	thisNode.Input = input
	return &thisNode
}

// Creates a new lock-free queue
func MakeQueue() Queue {
	thisNode := MakeNode(nil, nil)
	head := unsafe.Pointer(thisNode)
	tail := unsafe.Pointer(thisNode)
	thisQueue := queueFeed{head, tail}
	return &thisQueue
}

// Method for enqueueing items in the lock-free queue
func (q *queueFeed) Enq(input *Request) {
	thisNode := MakeNode(input, nil)
	for {
		last := q.tail
		next := (*queueNode)(last).Next
		if last == q.tail {
			if next == nil {
				if atomic.CompareAndSwapPointer(&(*queueNode)(last).Next, nil, unsafe.Pointer(thisNode)) {
					atomic.CompareAndSwapPointer(&q.tail, last, unsafe.Pointer(thisNode))
					return
				} else {
					atomic.CompareAndSwapPointer(&q.tail, last, unsafe.Pointer(thisNode))
				}
			}
		}
	}
}

// Method for dequeueing items in the lock-free queue
func (q *queueFeed) Deq() *Request {
	for {
		first := q.head
		next := (*queueNode)(first).Next
		if next != nil {
			if atomic.CompareAndSwapPointer(&q.head, first, unsafe.Pointer(next)) {
				return (*queueNode)(next).Input
			}
		} else {
			return nil
		}
	}
}

// Checks for the lock-free queue being empty
func (q *queueFeed) Emp() bool {
	if q.head == q.tail {
		return true
	} else {
		return false
	}
}
