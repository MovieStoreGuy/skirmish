package types

type Instance struct {
	Id      uint64
	Name    string
	Zone    string
	Project string
	Labels  map[string]string
}
