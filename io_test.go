package stickyshift

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	yaml "gopkg.in/yaml.v2"
)

func TestRead(t *testing.T) {
	// bad file
	_, err := Read("💥")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")

	for _, test := range []struct {
		msg     string
		in      string
		want    Schedule
		wantErr string
	}{
		{
			msg: "ok",
			in: `
id: _
shifts: []
`,
			want: Schedule{
				Id: "_",
			},
			wantErr: "",
		},
		{
			msg: "bad yaml",
			in: `
_: _
`,
			wantErr: "field _ not found in type",
		},
		{
			msg:     "fail checks",
			in:      `{}`,
			wantErr: "schedule is missing `id`",
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			f := tmp(t, test.in)
			defer os.Remove(f)
			s, err := Read(f)

			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, s)
			}
		})
	}
}

func TestUnmarshalShifts(t *testing.T) {
	sl := &ShiftList{}
	assert.Error(t, sl.UnmarshalYAML(func(interface{}) error {
		return errors.New("_")
	}))

	t0 := time.Time{}
	t1 := t0.Add(time.Second)
	t2 := t1.Add(time.Second)

	for _, test := range []struct {
		msg     string
		in      string
		want    ShiftList
		wantErr string
	}{
		{
			msg: "ok",
			in: fmt.Sprintf(`
%v: a
%v: b
%v: %s
`, t0.Format(_timeFmt), t1.Format(_timeFmt), t2.Format(_timeFmt), _shiftListEnder),
			want: ShiftList{
				{
					Start: t0,
					End:   t1,
					Email: "a",
				},
				{
					Start: t1,
					End:   t2,
					Email: "b",
				},
			},
			wantErr: "",
		},
		{
			msg: "no shift list ending",
			in: fmt.Sprintf(`
%v: a
%v: b`, t0.Format(_timeFmt), t1.Format(_timeFmt)),
			wantErr: fmt.Sprintf("last shift must have user %q", _shiftListEnder),
		},
		{
			msg:     "bad yaml",
			in:      `{{{{{`,
			wantErr: "did not find expected node",
		},
		{
			msg:     "bad key",
			in:      `- _: _`,
			wantErr: "shift time is not a yaml string",
		},
		{
			msg:     "bad value",
			in:      `_: 1`,
			wantErr: "shift email is not a yaml string",
		},
		{
			msg:     "bad timestamp",
			in:      `_: _`,
			wantErr: `cannot parse "_" as "2006"`,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			sl := ShiftList{}
			err := yaml.Unmarshal([]byte(test.in), &sl)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				require.NoError(t, err)
				assert.Equal(t, test.want, sl)
			}
		})
	}
}

func TestWrite(t *testing.T) {
	err := Write("/💩", Schedule{Shifts: ShiftList{{}}})
	require.Error(t, err)
	require.Contains(t, err.Error(), "permission denied")

	for _, test := range []struct {
		msg     string
		sched   Schedule
		want    string
		wantErr string
	}{
		{
			msg:     "empty",
			sched:   Schedule{},
			wantErr: "cannot marshal an empty shift list",
		},
		{
			msg: "ok",
			sched: Schedule{
				Id: "xxx",
				Shifts: ShiftList{
					{
						Start: mustTime(t, "1970-01-01T00:00:00-07:00"),
						End:   mustTime(t, "1970-01-02T00:00:00-07:00"),
						Email: "foo",
					},
					{
						Start: mustTime(t, "1970-01-02T00:00:00-07:00"),
						End:   mustTime(t, "1970-01-03T00:00:00-07:00"),
						Email: "bar",
					},
				},
			},
			want: `id: xxx
shifts:
  1970-01-01T00:00:00-07:00: foo
  1970-01-02T00:00:00-07:00: bar
  1970-01-03T00:00:00-07:00: TBD
`,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			f, err := ioutil.TempFile("", "")
			require.NoError(t, err)
			path := f.Name()
			defer os.Remove(path)

			err = Write(path, test.sched)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			bs, err := ioutil.ReadFile(path)
			require.NoError(t, err)
			assert.Equal(t, test.want, string(bs))
		})
	}
}

func mustTime(t *testing.T, ts string) time.Time {
	res, err := time.Parse(time.RFC3339, ts)
	require.NoError(t, err)
	return res
}

func tmp(t *testing.T, contents string) string {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = f.Write([]byte(contents))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}
