package types

// Firewall defines the internal structure of what
// the orchestrator needs to know to restore operations
type Firewall struct {
	Project   string
	Name      string
	Id        uint64
	Instances []*Instance
	Labels    map[string]string
}
