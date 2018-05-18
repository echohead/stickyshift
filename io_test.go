package stickyshift

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRead(t *testing.T) {
	// bad file
	_, err := Read("💥")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no such file")

	// ok
	in := `
name: _
users:
- x
- y
shifts:
  _: _
`
	want := Schedule{
		Name: "_",
		Users: []string{
			"x",
			"y",
		},
		Shifts: map[string]string{
			"_": "_",
		},
	}

	f := tmp(t, in)
	defer os.Remove(f)
	s, err := Read(f)
	assert.NoError(t, err)
	assert.Equal(t, want, s)
}

func tmp(t *testing.T, contents string) string {
	f, err := ioutil.TempFile("", "")
	require.NoError(t, err)
	_, err = f.Write([]byte(contents))
	require.NoError(t, err)
	require.NoError(t, f.Close())
	return f.Name()
}
