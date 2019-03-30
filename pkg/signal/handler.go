package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

// Handler stores all the required information to gracefully shutdown
type Handler struct {
	operations []func()
	shutdown   chan bool
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
	h.shutdown <- true
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
		<-h.shutdown
	default:
		fmt.Fprintf(os.Stderr, "Recovered from %v, terminating gracefully\n", r)
	}
	for _, op := range h.operations {
		op()
	}
}

// Done to be called outside of the
func (h *Handler) Done() {
	h.shutdown <- true
}

// NewHandler returns a new configured handler
func NewHandler() *Handler {
	return &Handler{
		shutdown: make(chan bool, 1),
	}
}
