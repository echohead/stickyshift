package stickyshift

import (
	"errors"
	"reflect"
	"sort"

	"go.uber.org/multierr"
)

func check(s Schedule) error {
	var errs error

	for _, check := range []func(Schedule) error{
		checkId,
		checkShiftListDupes,
		checkShiftListSorted,
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
