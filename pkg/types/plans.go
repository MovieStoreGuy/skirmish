package types

import (
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

// Plan defines the structure of the game day
type Plan struct {
	Mode     string   `json:"mode" yaml:"mode" description:"defines how aggressive each step is preformed"`
	Projects []string `json:"projects" yaml:"projects" description:"define each Google Cloud Project to operate in"`
	Steps    []Step   `json:"steps" yaml:"steps"`
}

type Step struct {
	Name        string        `json:"name" yaml:"name"`
	Description string        `json:"description" yaml:"description"`
	Operations  []string      `json:"operations" yaml:"operations" description:"It is the name of the loaded minions in the orchestrator"`
	Projects    []string      `json:"projects" yaml:"projects"`
	Exclude     Exclude       `json:"exclude" yaml:"exclude" description:"define all the things to exclude on"`
	Settings    Settings      `json:"settings" yaml:"settings"`
	Wait        time.Duration `json:"wait" yaml:"wait"`
	Sample      float32       `json:"sample" yaml:"sample" description:"Sample is rate [0.0,100.0] that will determine the likely hood of an instance being affected"`
}

type Exclude struct {
	Labels    map[string]string `json:"labels" yaml:"labels" description:"define the labels to ignore resource "`
	Zones     []string          `json:"zones" yaml:"zones" description:"define the zones to ignore"`
	Regions   []string          `json:"regions" yaml:"regions" description:"define the regions to ignore"`
	Wildcards []string          `json:"wildcards" yaml:"wildcards" description:"If the affected resources doesn't match, see if its name matches the wildcard'"`
}

type Settings struct {
	Network []struct {
		Project string `json:"project" yaml:"project"`
		Network string `json:"network" yaml:"network"`
		Deny    []Deny `json:"deny" yaml:"deny"`
	}
}

type Deny struct {
	Protocol string   `json:"protocol" yaml:"protocol"`
	Ports    []string `json:"ports" yaml:"ports"`
}

// Validate will ensure that the expected format of the plan
func (p *Plan) validate() error {
	switch p.Mode {
	case DryRun, Repairable, Destruction:
		// Valid options
	default:
		return fmt.Errorf("unknown mode %s", p.Mode)
	}
	// Validate that each step has a valid component
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
		if s.Sample < 0.0 || s.Sample > 100.0 {
			return fmt.Errorf("step %d has invalid sample, sample is require to be within [0.0, 100.0]", index)
		}
		for _, project := range s.Projects {
			found := false
			for _, p := range p.Projects {
				if p == project {
					found = true
				}
			}
			if !found {
				return fmt.Errorf("step %d has defined an additional project %s which is missing from global list", index, project)
			}
		}
	}
	return nil
}

// LoadPlan will read the filepath and try load it into a valid plan
func LoadPlan(filepath string) (*Plan, error) {
	buff, err := ioutil.ReadFile(filepath)
	if err != nil {
		return nil, err
	}
	var p Plan
	if err := yaml.Unmarshal(buff, &p); err != nil {
		return nil, err
	}
	for _, step := range p.Steps {
		if step.Sample == 0.0 {
			step.Sample = 100.0
		}
		step.Sample = step.Sample / 100.0
	}
	return &p, (&p).validate()
}
