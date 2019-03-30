package signal

import "sync"

var (
	global *Handler
	once   sync.Once
)

// GlobalHandler will return a singleton to be used throughout the
// lifetime of the application
func GlobalHandler() *Handler {
	once.Do(func() {
		global = NewHandler()
	})
	return global
}
