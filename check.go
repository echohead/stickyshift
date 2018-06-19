package stickyshift

import (
	"errors"
	"fmt"
	"reflect"
	"sort"

	"go.uber.org/multierr"
)

func check(s Schedule) error {
	var errs error

	for _, check := range []func(Schedule) error{
		checkId,
		checkShiftListDupes,
		checkShiftListDupeEmail,
		checkShiftListSorted,
		checkExtendMinDays,
		checkExtendMaxDays,
		checkExtendMinLessThanMax,
	} {
		errs = multierr.Append(errs, check(s))

	}
	return errs
}

func checkId(s Schedule) error {
	if s.Id == "" {
		return errors.New("schedule is missing `id` field")
	}
	return nil
}

func checkShiftListDupes(s Schedule) error {
	for i := 0; i < len(s.Shifts)-1; i += 1 {
		if s.Shifts[i].Start == s.Shifts[i+1].Start {
			return errors.New("start timestamps in `shifts` must be unique")
		}
	}
	return nil
}

func checkShiftListDupeEmail(s Schedule) error {
	for i := 0; i < len(s.Shifts)-1; i += 1 {
		if s.Shifts[i].Email == s.Shifts[i+1].Email {
			return fmt.Errorf("%v appears for two shifts in a row.  this should instead be expressed as a single, longer shift", s.Shifts[i].Email)
		}
	}
	return nil
}

func checkShiftListSorted(s Schedule) error {
	if s.Shifts == nil {
		s.Shifts = ShiftList{}
	}
	tmp := make(ShiftList, len(s.Shifts))
	copy(tmp, s.Shifts)
	sort.Slice(tmp, func(i, j int) bool {
		return tmp[i].Start.Before(tmp[j].Start)
	})
	if !reflect.DeepEqual(tmp, s.Shifts) {
		return errors.New("`shifts` must be ordered by time")
	}
	return nil
}

const (
	_minMaxDays = 21
	_maxMaxDays = 56
)

func checkExtendMaxDays(s Schedule) error {
	if s.Extend == nil {
		return nil
	}
	if s.Extend.MaxDays < _minMaxDays || s.Extend.MaxDays > _maxMaxDays {
		return fmt.Errorf("extend.maxDays must be between %v and %v, but found %v", _minMaxDays, _maxMaxDays, s.Extend.MaxDays)
	}
	return nil
}

const (
	_minMinDays = 14
	_maxMinDays = 30
)

func checkExtendMinDays(s Schedule) error {
	if s.Extend == nil {
		return nil
	}
	if s.Extend.MinDays < _minMinDays || s.Extend.MinDays > _maxMinDays {
		return fmt.Errorf("extend.minDays must be between %v and %v, but found %v", _minMinDays, _maxMinDays, s.Extend.MinDays)
	}
	return nil
}

func checkExtendMinLessThanMax(s Schedule) error {
	if s.Extend == nil || s.Extend.MinDays < s.Extend.MaxDays {
		return nil
	}
	return fmt.Errorf("extend.minDays must be less than extend.maxDays, but %v >= %v", s.Extend.MinDays, s.Extend.MaxDays)
}
