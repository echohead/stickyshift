package stickyshift

import (
	"io/ioutil"

	"gopkg.in/yaml.v2"
)

// Schedule represents an oncall schedule
type Schedule struct {
	Name   string            `yaml:"name"`
	Users  []string          `yaml:"users"`
	Shifts map[string]string `yaml:"shifts"`
}

// Read loads a schedule from the given yaml file
func Read(f string) (s Schedule, err error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}
	yaml.UnmarshalStrict(bs, &s)
	return
}
