package types

type Firewall struct {
	Project   string
	Name      string
	Id        uint64
	Instances []Instance
	Labels    map[string]string
}
