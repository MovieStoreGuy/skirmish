package types

import (
	"fmt"
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Plan struct {
	Mode     string `json:"mode" yaml:"mode" description:"defines how aggressive each step is preformed"`
	Projects []struct {
		Project string `json:"project" yaml:"project"`
		KeyPath string `json:"keyPath" yaml:"keyPath"`
	} `json:"projects" yaml:"projects" description:"define each Google Cloud Project to operate in"`
	Steps []Step `json:"steps" yaml:"steps"`
}

type Step struct {
	Name        string   `json:"name" yaml:"name"`
	Description string   `json:"description" yaml:"description"`
	Operations  []string `json:"operations" yaml:"operations"`
	Projects    []string `json:"projects" yaml:"projects"`
	Exclude     struct {
		Labels    []string `json:"labels" yaml:"labels" description:"define the labels to ignore resource "`
		Zones     []string `json:"zones" yaml:"zones" description:"define the zones to ignore"`
		Regions   []string `json:"regions" yaml:"regions" description:"define the regions to ignore"`
		Wildcards []string `json:"wildcards" yaml:"wildcards" description:"If the affected resources doesn't match, see if its name matches the wildcard'"`
	} `json:"exclude" yaml:"exclude" description:"define all the things to exclude on"`
}

func (p *Plan) Validate() error {
	switch p.Mode {
	case DryRun, Repairable, Destruction:
		// Valid options
	default:
		return fmt.Errorf("unknown mode %s", p.Mode)
	}
	for index, s := range p.Steps {
		if s.Name == "" {
			return fmt.Errorf("step %d requires a name", index)
		}
		if len(s.Projects) == 0 {
			return fmt.Errorf("step %d requires a projects to operate in", index)
		}
		if len(s.Operations) == 0 {
			return fmt.Errorf("step %d requires a operations to run", index)
		}
	}
	return nil
}

func LoadPlan(filepath string) (*Plan, error) {
	buff, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var p Plan
	if err := yaml.Unmarshal(buff, &p); err != nil {
		return nil, err
	}
	return &p, (&p).Validate()
}
