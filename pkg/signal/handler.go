package signal

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"sync"
)

type Handler struct {
	operations []func()
	shutdown   chan bool
	lock       sync.Mutex
}

func (h *Handler) Register(op func()) {
	h.lock.Lock()
	defer h.lock.Unlock()
	h.operations = append(h.operations, op)
}

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
}

// Finalise will either recover from a panic or await the shutdown signal to run the final operations
func (h *Handler) Finalise() {
	r := recover()
	switch r {
	case nil:
		<-h.shutdown
	default:
		fmt.Fprintf(os.Stderr, "Recovered from %v, terminating gracefully", r)
	}
	for _, op := range h.operations {
		op()
	}
}

// Done to be called outside of the
func (h *Handler) Done() {
	h.shutdown <- true
}

func NewHandler() *Handler {
	return &Handler{
		shutdown: make(chan bool, 1),
	}
}
