package stickyshift

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	_shiftEmpty = Shift{}
	_shiftEnd   = Shift{Email: _shiftListEnder}

	t0 = time.Time{}
	t1 = t0.Add(time.Second)
)

func TestCheck(t *testing.T) {
	err := check(Schedule{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing `id`")
}

func TestCheckId(t *testing.T) {
	expectValid(t, checkId,
		Schedule{Id: "_"},
	)
	expectInvalid(t, checkId,
		Schedule{},
	)
}

func TestShiftListSorted(t *testing.T) {
	first := Shift{Start: t0}
	second := Shift{Start: t1}

	expectValid(t, checkShiftListSorted,
		Schedule{Shifts: ShiftList{}},
		Schedule{Shifts: ShiftList{first, second}},
	)
	expectInvalid(t, checkShiftListSorted,
		Schedule{Shifts: ShiftList{second, first}},
	)
}

func TestShiftListDupes(t *testing.T) {
	a := Shift{Start: t0}
	b := Shift{Start: t1}

	expectValid(t, checkShiftListDupes,
		Schedule{Shifts: ShiftList{}},
		Schedule{Shifts: ShiftList{a}},
		Schedule{Shifts: ShiftList{a, b}},
	)
	expectInvalid(t, checkShiftListDupes,
		Schedule{Shifts: ShiftList{a, a}},
	)
}

func TestShiftListDupeEmails(t *testing.T) {
	expectValid(t, checkShiftListDupeEmail,
		Schedule{Shifts: ShiftList{Shift{Start: t0, Email: "a"}, Shift{Start: t1, Email: "b"}}},
	)
	expectInvalid(t, checkShiftListDupeEmail,
		Schedule{Shifts: ShiftList{Shift{Start: t0, Email: "a"}, Shift{Start: t1, Email: "a"}}},
	)
}

func TestCheckExtendMaxDays(t *testing.T) {
	expectValid(t, checkExtendMaxDays,
		Schedule{},
		Schedule{Extend: &Extend{MaxDays: _minMaxDays}},
	)
	expectInvalid(t, checkExtendMaxDays,
		Schedule{Extend: &Extend{MaxDays: _maxMaxDays + 1}},
	)
}

func TestCheckExtendMinDays(t *testing.T) {
	expectValid(t, checkExtendMinDays,
		Schedule{},
		Schedule{Extend: &Extend{MinDays: _minMinDays}},
	)
	expectInvalid(t, checkExtendMinDays,
		Schedule{Extend: &Extend{MinDays: _maxMinDays + 1}},
	)
}

func TestExtendMinLessThanMax(t *testing.T) {
	expectValid(t, checkExtendMinLessThanMax,
		Schedule{},
		Schedule{Extend: &Extend{MinDays: 1, MaxDays: 2}},
	)
	expectInvalid(t, checkExtendMinLessThanMax,
		Schedule{Extend: &Extend{MinDays: 3, MaxDays: 2}},
	)
}

func timeFromStr(t *testing.T, s string) time.Time {
	res, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return res
}

func expectValid(t *testing.T, f func(Schedule) error, ss ...Schedule) {
	for _, s := range ss {
		assert.NoError(t, f(s), "expected %+v to be valid", s)
	}
}

func expectInvalid(t *testing.T, f func(Schedule) error, ss ...Schedule) {
	for _, s := range ss {
		assert.Error(t, f(s), "expected %+v to be invalid", s)
	}
}
