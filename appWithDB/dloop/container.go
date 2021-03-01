package dloop

import (
	"sync"
	"time"
)

// TimeHandle be hold to be call on time
type TimeHandle func(n time.Time, quit <-chan int) error

type workPair struct {
	d    time.Duration
	core TimeHandle
}

type container struct {
	l    []workPair
	lock sync.Mutex
}

func newContainer() *container {
	return &container{
		l: make([]workPair, 0, 1),
	}
}

func (conta *container) push(d time.Duration, cb TimeHandle) {
	conta.lock.Lock()
	defer conta.lock.Unlock()

	conta.l = append(conta.l, workPair{d, cb})
}

func (conta *container) get() []workPair {
	return conta.l
}

var globalContainer *container

func initContainer() {
	globalContainer = newContainer()
}

func getContainer() *container {
	return globalContainer
}
