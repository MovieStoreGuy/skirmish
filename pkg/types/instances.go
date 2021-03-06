package types

// Instance defines all the required values for the internal structure
// so that the orchestrator can restore it back the original state.
type Instance struct {
	Id      uint64
	Name    string
	Zone    string
	Region  string
	Project string
	Labels  map[string]string
}

func (i *Instance) CompleteZone() string {
	return i.Region + "-" + i.Zone
}
