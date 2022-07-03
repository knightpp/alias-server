package fp

import (
	"fmt"
	"sync"
)

type Locker[T any] struct {
	inner T
	mutex *sync.Mutex
}

func NewLocker[T any](inner T) Locker[T] {
	return Locker[T]{
		inner: inner,
		mutex: &sync.Mutex{},
	}
}

func (l Locker[T]) String() string {
	return fmt.Sprintf("%v", l.inner)
}

func (l Locker[T]) LockE(f func(inner T) error) error {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return f(l.inner)
}

func Lock1[T any, U any](l Locker[T], f func(inner T) U) U {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return f(l.inner)
}

func Lock2[T any, U any, J any](l Locker[T], f func(inner T) (U, J)) (U, J) {
	l.mutex.Lock()
	defer l.mutex.Unlock()

	return f(l.inner)
}
