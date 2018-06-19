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
		Id     string    `yaml:"id"`
		Extend *Extend   `yaml:"extend"`
		Shifts ShiftList `yaml:"shifts"`
	}

	// Shift represents an oncall shift
	Shift struct {
		Email string
		Start time.Time
		End   time.Time
	}

	ShiftList []Shift

	// Extend contains options for extending the schedule
	Extend struct {
		MinDays int      `yaml:"minDays"`
		MaxDays int      `yaml:"maxDays"`
		Users   []string `yaml:"users"`
	}
)

const (
	_shiftListEnder = "TBD"
	_timeFmt        = time.RFC3339
)

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
