package signal

import "sync"

var (
	global *Handler
	once   sync.Once
)

func GlobalHandler() *Handler {
	once.Do(func() {
		global = NewHandler()
	})
	return global
}
