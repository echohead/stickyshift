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

func tmp(t *testing.T, contents string) string {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = f.Write([]byte(contents))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}
