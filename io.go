package stickyshift

import (
	"errors"
	"fmt"
	"io/ioutil"
	"time"

	"gopkg.in/yaml.v2"
)

type (
	// Schedule represents an oncall schedule
	Schedule struct {
		Id     string      `yaml:"id"`
		Extend *ExtendOpts `yaml:"extend,omitempty"`
		Shifts ShiftList   `yaml:"shifts"`
	}

	// Shift represents an oncall shift
	Shift struct {
		Email string
		Start time.Time
		End   time.Time
	}

	ShiftList []Shift

	// ExtendOpts contains options for extending the schedule
	ExtendOpts struct {
		MinDays int      `yaml:"minDays"`
		MaxDays int      `yaml:"maxDays"`
		Users   []string `yaml:"users"`
	}
)

const (
	_shiftListEnder = "TBD"
	_timeFmt        = time.RFC3339
)

// UnmarshalYAML deserializes a yaml input map into a ShiftList
// a custom unmarshaller is used because we care about the order of the keys in the input.
func (sl *ShiftList) UnmarshalYAML(unmarshal func(interface{}) error) error {
	m := yaml.MapSlice{}
	if err := unmarshal(&m); err != nil {
		return err
	}
	for i, mi := range m {
		k, ok := mi.Key.(string)
		if !ok {
			return errors.New("shift time is not a yaml string")
		}
		v, ok := mi.Value.(string)
		if !ok {
			return errors.New("shift email is not a yaml string")
		}

		s, err := kvToShift(k, v)
		if err != nil {
			return err
		}
		if i > 0 {
			(*sl)[i-1].End = s.Start
		}

		if i < len(m)-1 {
			*sl = append(*sl, s)
		}

		if i == len(m)-1 {
			if v != _shiftListEnder {
				return fmt.Errorf("last shift must have user %q, but found %q", _shiftListEnder, v)
			}
		}
	}
	return nil
}

func kvToShift(k, v string) (s Shift, err error) {
	t, err := time.Parse(time.RFC3339, k)
	if err != nil {
		return
	}
	s.Email = v
	s.Start = t
	return
}

// Read loads a schedule from the given yaml file
func Read(f string) (s Schedule, err error) {
	bs, err := ioutil.ReadFile(f)
	if err != nil {
		return
	}
	if err = yaml.UnmarshalStrict(bs, &s); err != nil {
		return Schedule{}, err
	}
	if err = check(s); err != nil {
		return Schedule{}, err
	}
	return s, nil
}

// MarshalYAML serializes a ShiftList to a yaml map
// a custom marshaller is used so we can translate a list into a map with ordered keys.
func (sl ShiftList) MarshalYAML() (interface{}, error) {
	if len(sl) < 1 {
		return nil, errors.New("cannot marshal an empty shift list")
	}
	m := yaml.MapSlice{}
	for _, s := range sl {
		m = append(m, yaml.MapItem{Key: s.Start, Value: s.Email})
	}
	m = append(m, yaml.MapItem{Key: sl[len(sl)-1].End, Value: _shiftListEnder})
	return m, nil
}

// Write serializes a schedule into the given path
func Write(path string, s Schedule) error {
	serialized, err := yaml.Marshal(s)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(path, serialized, 0644)
}
