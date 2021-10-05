package state

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAppendCloser(t *testing.T) {
	f1Called := false
	f1 := func() { f1Called = true }
	AppendCloser(f1)

	f2Called := false
	f2 := func() { f2Called = true }
	AppendCloser(f2)

	CloserFn()()
	assert.True(t, f1Called)
	assert.True(t, f2Called)

	reset()
}

func TestAppendCloserConcurrently(t *testing.T) {
	wg := sync.WaitGroup{}

	wg.Add(1)
	go func() {
		AppendCloser(func() {})
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		AppendCloser(func() {})
		wg.Done()
	}()

	wg.Wait()

	assert.Len(t, closers, 2)

	reset()
}
