package types

import "sync"

// Metadata stores all the relevant cloud data that needs to be computed at runtime.
type Metadata struct {
	Once    sync.Once
	Regions []string
	Zones   []string
}
