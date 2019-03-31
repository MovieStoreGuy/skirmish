package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"

	"go.uber.org/atomic"
)

// Handler stores all the required information to gracefully shutdown
type Handler struct {
	operations []func()
	shutdown   chan bool
	done       atomic.Int32
	lock       sync.Mutex
}

// Register allows to define what operations need to happen when the system
// is shutting down. Register doesn't ensure order outside of a single threaded context.
func (h *Handler) Register(op func()) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.operations = append(h.operations, op)
}

// Await registers a signal handler for all the passed OS signals, awaits either the system signal to be called
// or for the context to be done the it will issue a shutdown
func (h *Handler) Await(ctx context.Context, cancel context.CancelFunc, sigs ...os.Signal) {
	ch := make(chan os.Signal, 1)
	defer close(ch)
	signal.Notify(ch, sigs...)
	select {
	case s := <-ch:
		fmt.Fprintf(os.Stderr, `{"recovered" : "%v"}`, s)
		cancel()
	case <-ctx.Done():
	}
	if h.done.Load() == 0 {
		h.shutdown <- true
	}
	h.done.Store(1)
	// In the event that someone tries to fire the same event
	// twice then it will ignore it till
	signal.Ignore(sigs...)
}

// Finalise will either recover from a panic or await the shutdown signal to run the final operations
// Should only be used within the main function as a deferred statement
func (h *Handler) Finalise() {
	defer close(h.shutdown)
	r := recover()
	switch r {
	case nil:
		if h.done.Load() == 0 {
			h.shutdown <- true
		}
		h.done.Store(1)
	default:
		fmt.Fprintf(os.Stderr, "Recovered from %v, terminating gracefully\n", r)
	}
	for _, op := range h.operations {
		op()
	}
	if h.done.Load() == 1 {
		close(h.shutdown)
	}
}

// Done to be called outside of the
func (h *Handler) Done() {
	if h.done.Load() == 0 {
		h.shutdown <- true
	}
	h.done.Store(1)
}

// NewHandler returns a new configured handler
func NewHandler() *Handler {
	return &Handler{
		shutdown: make(chan bool, 1),
	}
}
