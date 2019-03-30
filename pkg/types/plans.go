package types

import (
	"fmt"
	"io/ioutil"
	"os"

	"gopkg.in/yaml.v2"
)

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

func (p *Plan) Validate() error {
	switch p.Mode {
	case DryRun, Repairable, Destruction:
		// Valid options
	default:
		return fmt.Errorf("unknown mode %s", p.Mode)
	}
	if _, err := os.Stat(p.DefaultKeyPath); len(p.DefaultKeyPath) != 0 || os.IsNotExist(err) {
		return fmt.Errorf("default key path %s does not exist", p.DefaultKeyPath)
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
