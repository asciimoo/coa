package config

import (
	"bufio"
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"path/filepath"

	"github.com/asciimoo/coa/notification"
	"github.com/asciimoo/coa/project"

	"github.com/go-yaml/yaml"
)

const SettingsFileName = "settings.yml"
const ProjectListFileName = "project.list"

type Config struct {
	initialized   bool
	ConfigFolder  string
	Projects      []*project.Project
	ServerAddress string
	Notifiers     []*notification.NotifierBackend
}

func countChar(s []byte, c byte) int {
	charCount := 0
	for _, r := range s {
		if r == c {
			charCount += 1
		}
	}
	return charCount
}

func (c *Config) Init(configFolder string) error {
	if c.initialized {
		return nil
	}

	_, err := os.Stat(configFolder)
	if err != nil {
		if err := os.MkdirAll(configFolder, os.ModePerm); err != nil {
			return errors.New(fmt.Sprintf("Config directory (%v) not found and cannot be created (%v).", configFolder, err.Error()))
		}
	}

	settingsFilePath := path.Join(configFolder, SettingsFileName)
	settingsFile, err := os.OpenFile(settingsFilePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return errors.New(fmt.Sprintf("Settings file (%v) not found.", settingsFilePath))
	}
	defer settingsFile.Close()

	settingsBytes, err := ioutil.ReadAll(settingsFile)
	if err != nil {
		return err
	}

	err = yaml.Unmarshal(settingsBytes, &c)
	if err != nil {
		return err
	}

	c.Projects = make([]*project.Project, 0, countChar(settingsBytes, byte('\n')))
	c.initialized = true
	c.ConfigFolder = configFolder

	projectListFilePath := path.Join(configFolder, ProjectListFileName)
	projectListFile, err := os.OpenFile(projectListFilePath, os.O_RDONLY|os.O_CREATE, 0600)
	if err != nil {
		return err
	}
	defer projectListFile.Close()

	scanner := bufio.NewScanner(projectListFile)
	for scanner.Scan() {
		projectPath := scanner.Text()
		p, err := project.Load(projectPath)
		if err != nil {
			return err
		}
		c.Projects = append(c.Projects, p)
	}

	return nil
}

func (c *Config) String() string {
	d, _ := yaml.Marshal(c)
	return string(d)
}

func (c *Config) AddProject(projectSettingsPath string) (*project.Project, error) {
	for _, p := range c.Projects {
		if p.ConfigPath == projectSettingsPath {
			return nil, errors.New("Project already added")
		}
	}

	projectSettingsPath, err := filepath.Abs(projectSettingsPath)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat(projectSettingsPath)
	if err != nil {
		return nil, err
	}

	p, err := project.Load(projectSettingsPath)
	if err != nil {
		return nil, err
	}

	projectListFilePath := path.Join(c.ConfigFolder, ProjectListFileName)
	projectFile, err := os.OpenFile(projectListFilePath, os.O_RDWR|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	defer projectFile.Close()

	fmt.Fprintf(projectFile, "%v\n", projectSettingsPath)

	c.Projects = append(c.Projects, p)

	return p, nil
}

func Load(configFolder string) (*Config, error) {
	c := &Config{
		ServerAddress: "127.0.0.1:4224",
	}
	err := c.Init(configFolder)
	return c, err
}
