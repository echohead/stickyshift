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
)

func TestCheck(t *testing.T) {
	err := check(Schedule{})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "missing `id`")
}

func TestCheckId(t *testing.T) {
	assert.Error(t, checkId(Schedule{}))
	assert.NoError(t, checkId(Schedule{Id: "_"}))
}

func TestShiftListSorted(t *testing.T) {
	t0 := time.Time{}
	first := Shift{Start: t0}
	second := Shift{Start: t0.Add(time.Second)}

	assert.NoError(t, checkShiftListSorted(Schedule{Shifts: ShiftList{}}))
	assert.NoError(t, checkShiftListSorted(Schedule{Shifts: ShiftList{first, second}}))
	assert.Error(t, checkShiftListSorted(Schedule{Shifts: ShiftList{second, first}}))
}

func TestShiftListDupes(t *testing.T) {
	t0 := time.Time{}
	a := Shift{Start: t0}
	b := Shift{Start: t0.Add(time.Second)}

	assert.NoError(t, checkShiftListDupes(Schedule{Shifts: ShiftList{}}))
	assert.NoError(t, checkShiftListDupes(Schedule{Shifts: ShiftList{a}}))
	assert.NoError(t, checkShiftListDupes(Schedule{Shifts: ShiftList{a, b}}))
	assert.Error(t, checkShiftListDupes(Schedule{Shifts: ShiftList{a, a}}))

}

func timeFromStr(t *testing.T, s string) time.Time {
	res, err := time.Parse(time.RFC3339, s)
	require.NoError(t, err)
	return res
}
