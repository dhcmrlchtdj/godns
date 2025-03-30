package util

import "sync"

type Deferred[T any] struct {
	val       T
	err       error
	fulfilled chan struct{}
	once      sync.Once
}

func MakeDeferred[T any]() *Deferred[T] {
	deferred := Deferred[T]{
		fulfilled: make(chan struct{}),
	}
	return &deferred
}

func (d *Deferred[T]) Resolve(val T) {
	d.once.Do(func() {
		d.val = val
		close(d.fulfilled)
	})
}

func (d *Deferred[T]) Reject(err error) {
	d.once.Do(func() {
		d.err = err
		close(d.fulfilled)
	})
}

func (d *Deferred[T]) Wait() (T, error) {
	<-d.fulfilled
	return d.val, d.err
}
