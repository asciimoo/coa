package project

import (
	"io/ioutil"
	"os"
	"path"

	"github.com/asciimoo/coa/checker"

	"github.com/go-yaml/yaml"
)

type Project struct {
	Name string
	ConfigPath string
	Checkers []*checker.Checker
	stop chan bool
}

func Load(projectConfigFile string) (*Project, error) {
	configFile, err := os.OpenFile(projectConfigFile, os.O_RDONLY, 0600)
	if err != nil {
		return nil, err
	}
	defer configFile.Close()

	bytes, err := ioutil.ReadAll(configFile)
	if err != nil {
		return nil, err
	}

	p := &Project{ConfigPath: projectConfigFile, stop: make(chan bool)}

	err = yaml.Unmarshal(bytes, &p)
	if err != nil {
		return nil, err
	}
	for _, c := range p.Checkers {
		c.ProjectName = p.Name
	}
	return p, nil
}

func (p *Project) Start() {
	cwd := path.Dir(p.ConfigPath)
	for _, c := range p.Checkers {
		go c.Start(cwd)
	}

	select {
	case <- p.stop:
		for _, c := range p.Checkers {
			c.Stop()
		}
	}
}

func (p *Project) Stop() {
	p.stop <- true
}
