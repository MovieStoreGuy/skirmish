package types

type Plan struct {
	Mode           string `json:"mode" yaml:"mode" description:"defines how aggressive each step is preformed"`
	DefaultKeyPath string `json:"defaultKeyPath" yaml:"defaultKeyPath"`
	Projects       []struct {
		Project string `json:"project" yaml:"project"`
		KeyPath string `json:"keyPath" yaml:"keyPath"`
	} `json:"projects" yaml:"projects" description:"define each Google Cloud Project to operate in"`
	Steps []struct {
		Operations []string `json:"operations" yaml:"operations"`
		Exclude    struct {
			Labels  []string `json:"labels" yaml:"labels" description:"define the labels to ignore resource "`
			Zones   []string `json:"zones" yaml:"zones" description:"define the zones to ignore"`
			Regions []string `json:"regions" yaml:"regions" description:"define the regions to ignore"`
		} `json:"exclude" yaml:"exclude" description:"define all the things to exclude on"`
	} `json:"steps" yaml:"steps"`
}
