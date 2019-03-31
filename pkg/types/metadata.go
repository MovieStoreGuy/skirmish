package types

import "sync"

type Metadata struct {
	Once    sync.Once
	Regions []string
	Zones   []string
}
