package state

import "sync"

// Closer is a function that will be executed on termination. It's main purpose is to
// wrap up on going tasks.
type Closer func()

var (
	closers   []Closer
	closerMux = &sync.Mutex{}
)

// AppendCloser appends a closer function into the set of closers.
func AppendCloser(f Closer) {
	closerMux.Lock()
	closers = append(closers, f)
	closerMux.Unlock()
}

// CloserFn returns a closer that internally will exectute all the closers appended.
func CloserFn() Closer {
	return func() {
		for _, f := range closers {
			f()
		}
	}
}

func reset() {
	closerMux.Lock()
	closers = nil
	closerMux.Unlock()
}
