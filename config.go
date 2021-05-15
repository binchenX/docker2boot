package main

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Kernel        string   `yaml:"kernel,omitempty"`
	UbuntuVersion string   `yaml:"ubuntuVersion,omitempty"`
	Login         string   `yaml:"login,omitempty"`
	Packages      []string `yaml:"packages,omitempty"`
	Systemd       Systemd  `yaml:"systemd,omitempty"`
	Files         []File   `yaml:"files,omitempty"`
}

type Systemd struct {
	Units []SystemdUnit `yaml:"units,omitempty"`
}

type SystemdUnit struct {
	Name    string `yaml:"name,omitempty"`
	Enabled bool   `yaml:"enabled,omitempty"`
}

type File struct {
	Path    string `yaml:"path,omitempty"`
	Mode    string `yaml:"mode,omitempty"`
	Content string `yaml:"content,omitempty"`
}

func getConfigFromFile(file string) (*Config, error) {
	data, err := ioutil.ReadFile(file)
	if err != nil {
		return nil, err
	}

	var config Config
	err = yaml.Unmarshal([]byte(data), &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
