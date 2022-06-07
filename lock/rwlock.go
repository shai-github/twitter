// Package lock provides an implementation of a read-write lock
// that uses condition variables and mutexes.
package lock

import (
	"sync"
)

type RWLock struct {
	condition  *sync.Cond
	mu         sync.Mutex
	writeCount int32
	readCount  int32
	maxReaders int32
}

func NewRWLock(maxReaders int32) *RWLock {
	cond := RWLock{}
	cond.condition = sync.NewCond(&cond.mu)
	cond.maxReaders = maxReaders
	return &cond
}

// Lock locks rw for writing
func (rw *RWLock) Lock() {
	rw.mu.Lock()
	if rw.writeCount == 0 {
		rw.condition.Signal()
	} else {
		for rw.writeCount > 0 {
			rw.condition.Wait()
		}
	}
	rw.writeCount++
	rw.mu.Unlock()
}

// Unlock unlocks rw for writing
func (rw *RWLock) Unlock() {
	rw.mu.Lock()
	rw.writeCount--
	rw.condition.Signal()
	rw.mu.Unlock()
}

// RLock locks rw for reading
func (rw *RWLock) RLock() {
	rw.mu.Lock()
	if rw.writeCount == 0 && rw.readCount < rw.maxReaders {
		rw.condition.Signal()
	} else {
		if rw.writeCount > 0 || rw.readCount == rw.maxReaders {
			rw.condition.Wait()
		}
	}
	rw.readCount++
	rw.mu.Unlock()
}

// RUnlock undoes a single RLock call; does not affect other simultaneous readers
func (rw *RWLock) RUnlock() {
	rw.mu.Lock()
	rw.readCount--
	rw.condition.Signal()
	rw.mu.Unlock()
}
