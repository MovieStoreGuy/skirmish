package types

// Config is the struct definition that is consumed by the cloud.Provider
type Config struct {
	Google struct {
		Projects []string `json:"projects" yaml:"projects"`
	} `json:"google" yaml:"google"`
}
