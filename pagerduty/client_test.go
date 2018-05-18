package pagerduty

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/echohead/stickyshift"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type failDoer struct{}

func (*failDoer) Do(*http.Request) (*http.Response, error) {
	return nil, errors.New("failDoer")
}

type mockDoer struct {
	status int
	body   io.ReadCloser
}

func newMockDoer(status int, body string) *mockDoer {
	return &mockDoer{
		status,
		noopCloser{bytes.NewBufferString(body)},
	}
}

func (d *mockDoer) Do(*http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: d.status,
		Body:       d.body,
	}, nil
}

type resp struct {
	status int
	body   string
}

type multiDoer struct {
	resps []resp
}

func newMultiDoer(resps []resp) *multiDoer {
	return &multiDoer{resps}
}

func (d *multiDoer) Do(req *http.Request) (*http.Response, error) {
	if len(d.resps) < 1 {
		return nil, fmt.Errorf("multiDoer has no responses to return for %+v", req)
	}
	r := d.resps[0]
	d.resps = d.resps[1:]
	return &http.Response{
		StatusCode: r.status,
		Body:       noopCloser{bytes.NewBufferString(r.body)},
	}, nil
}

type failReader struct{}

func (failReader) Read([]byte) (int, error) {
	return 0, errors.New("failReader")
}

type noopCloser struct {
	io.Reader
}

func (noopCloser) Close() error {
	return nil
}

var (
	_bodyEmptyString  = noopCloser{bytes.NewBufferString(`""`)}
	_bodyXString      = noopCloser{bytes.NewBufferString(`"x"`)}
	_bodyFailRead     = noopCloser{failReader{}}
	_clientOk         = newMockDoer(http.StatusOK, ``)
	_clientFail       = &failDoer{}
	_clientFailRead   = &mockDoer{http.StatusOK, _bodyFailRead}
	_clientBadRequest = newMockDoer(http.StatusBadRequest, ``)
	_headers          = map[string]string{}
	_methodBad        = "💥"
	_urlOk            = "http://_"
	_urlBad           = "💥"
)

func TestNew(t *testing.T) {
	require.NoError(t, os.Unsetenv(_tokenEnvVar))
	c, err := New()
	assert.Nil(t, c)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), _tokenEnvVar)

	require.NoError(t, os.Setenv(_tokenEnvVar, "_"))
	defer os.Unsetenv(_tokenEnvVar)
	c, err = New()
	assert.NotNil(t, c)
	assert.NoError(t, err)
}

func TestSetHeaders(t *testing.T) {
	c := &clientImpl{
		headers: map[string]string{},
	}
	c.headers["foo"] = "bar"
	r, err := http.NewRequest(http.MethodGet, "", nil)
	require.NoError(t, err)

	c.setHeaders(r)
	assert.Equal(t, "bar", r.Header.Get("foo"))
}

func TestRequest(t *testing.T) {
	for _, test := range []struct {
		msg     string
		d       doer
		url     string
		method  string
		wantErr string
	}{
		{
			msg:     "bad url",
			d:       &http.Client{},
			url:     _urlBad,
			method:  http.MethodGet,
			wantErr: "unsupported protocol",
		},
		{
			msg:     "bad method",
			d:       &http.Client{},
			url:     _urlOk,
			method:  _methodBad,
			wantErr: "invalid method",
		},
		{
			msg:     "request fails",
			d:       _clientFail,
			url:     _urlOk,
			method:  http.MethodGet,
			wantErr: "failDoer",
		},
		{
			msg:     "body read fail",
			d:       _clientFailRead,
			url:     _urlOk,
			method:  http.MethodGet,
			wantErr: "failReader",
		},
		{
			msg:     "unexpected status code",
			d:       _clientBadRequest,
			url:     _urlOk,
			method:  http.MethodGet,
			wantErr: "expected 200 response",
		},
		{
			msg:     "ok",
			d:       _clientOk,
			url:     _urlOk,
			method:  http.MethodGet,
			wantErr: "",
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{test.d, test.url, _headers, map[string]string{}}
			_, err := c.request(test.method, "_", http.StatusOK, nil)
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestGet(t *testing.T) {
	res := ""

	c := &clientImpl{_clientFail, "http://_", _headers, map[string]string{}}
	err := c.get("_", &res)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failDoer")
	assert.Empty(t, res)

	c = &clientImpl{
		&mockDoer{
			status: http.StatusOK,
			body:   _bodyXString,
		},
		"http://_",
		_headers,
		map[string]string{},
	}
	err = c.get("_", &res)
	assert.NoError(t, err)
	assert.Equal(t, "x", res)
}

func TestPost(t *testing.T) {
	c := &clientImpl{_clientFail, "http://_", _headers, map[string]string{}}

	err := c.post("_", math.Inf(1))
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unsupported value")

	err = c.post("_", "_")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "failDoer")
}

func TestSync(t *testing.T) {
	t0, err := time.Parse(time.RFC3339, "1970-01-01T00:00:00-07:00")
	require.NoError(t, err)
	t1, err := time.Parse(time.RFC3339, "2100-01-01T00:00:00-07:00")
	require.NoError(t, err)

	users := map[string]string{"foo@bar.com": "someID"}
	shifts := []stickyshift.Shift{stickyshift.Shift{Email: "foo@bar.com", Start: t0, End: t1}}
	unknownUserShifts := []stickyshift.Shift{stickyshift.Shift{Email: "💥", Start: t0, End: t1}}

	for _, test := range []struct {
		msg     string
		d       doer
		in      []stickyshift.Shift
		wantErr string
	}{
		{
			msg: "no shifts to sync",
			in:  []stickyshift.Shift{},
		},
		{
			msg:     "get schedule fails",
			d:       _clientBadRequest,
			in:      shifts,
			wantErr: "got 400",
		},
		{
			msg: "get overrides fails",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusBadRequest, "_"},
			}),
			in:      shifts,
			wantErr: "got 400",
		},
		{
			msg: "shift in the past",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusOK, `{"overrides": []}`},
			}),
			in: []stickyshift.Shift{stickyshift.Shift{End: time.Unix(0, 0)}},
		},
		{
			msg: "error getting user",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusOK, `{"overrides": []}`},
				{http.StatusBadGateway, "_"},
			}),
			in:      unknownUserShifts,
			wantErr: "got 502",
		},
		{
			msg: "override already exists",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusOK, `{"overrides": [{"user": {"id": "someID"}, "start": "1970-01-01T00:00:00-07:00", "end": "2100-01-01T00:00:00-07:00"}]}`},
			}),
			in: shifts,
		},
		{
			msg: "create override fails",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusOK, `{"overrides": []}`},
				{http.StatusBadRequest, "_"},
			}),
			in:      shifts,
			wantErr: "got 400",
		},
		{
			msg: "ok",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"id": "_", "name": "_"}`},
				{http.StatusOK, `{"overrides": []}`},
				{http.StatusCreated, "_"},
			}),
			in: shifts,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{test.d, "_", _headers, users}
			err := c.Sync("_", test.in)
			if test.wantErr == "" {
				require.NoError(t, err)
				return
			}
			require.Error(t, err)
			assert.Contains(t, err.Error(), test.wantErr)
		})
	}
}

func TestGetSchedule(t *testing.T) {
	for _, test := range []struct {
		msg     string
		status  int
		body    string
		want    Schedule
		wantErr bool
	}{
		{
			"ok",
			http.StatusOK,
			`{"schedule": {"name": "foo", "id": "bar", "...": "..."}}`,
			Schedule{Name: "foo", Id: "bar"},
			false,
		},
		{
			"fail",
			http.StatusInternalServerError,
			"_",
			Schedule{},
			true,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{newMockDoer(test.status, test.body), "_", _headers, map[string]string{}}
			res, err := c.GetSchedule("_")
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want, res)
			}
		})
	}
}

func TestGetOverrides(t *testing.T) {
	t1, err := time.Parse(time.RFC3339, "2018-05-18T12:13:14-07:00")
	require.NoError(t, err)
	t2, err := time.Parse(time.RFC3339, "2018-05-19T01:02:03-07:00")
	require.NoError(t, err)

	for _, test := range []struct {
		msg     string
		status  int
		body    string
		want    []Override
		wantErr bool
	}{
		{
			"ok",
			http.StatusOK,
			`{"overrides": [{"start": "2018-05-18T12:13:14-07:00", "end": "2018-05-19T01:02:03-07:00", "user": {"id": "xyz", "type": "user_reference"}}]}`,
			[]Override{
				{
					User: userRef{
						Id:   "xyz",
						Type: "user_reference",
					},
					Start: t1,
					End:   t2,
				},
			},
			false,
		},
		{
			"fail",
			http.StatusInternalServerError,
			"_",
			[]Override{},
			true,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{newMockDoer(test.status, test.body), "_", _headers, map[string]string{}}
			res, err := c.GetOverrides("_", time.Now(), time.Now())
			if test.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, test.want, res)
			}
		})
	}
}

func TestGetUser(t *testing.T) {
	for _, test := range []struct {
		msg     string
		status  int
		body    string
		wantErr string
	}{
		{
			msg:     "request fails",
			status:  http.StatusInternalServerError,
			wantErr: "got 500",
		},
		{
			msg:     "multiple users returned",
			status:  http.StatusOK,
			body:    `{"users": [{}, {}]}`,
			wantErr: "expected one user",
		},
		{
			msg:     "mismatched email",
			status:  http.StatusOK,
			body:    `{"users": [{"email": "💥"}]}`,
			wantErr: "got user with email",
		},
		{
			msg:    "ok",
			status: http.StatusOK,
			body:   `{"users": [{"email": "foo@bar.com"}]}`,
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{newMockDoer(test.status, test.body), "_", _headers, map[string]string{}}
			res, err := c.getUser("foo@bar.com")
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, res.Email, "foo@bar.com")
		})
	}
}

func TestCreateOverride(t *testing.T) {
	for _, test := range []struct {
		msg     string
		d       doer
		wantErr string
	}{
		{
			msg:     "fail to get user",
			d:       _clientBadRequest,
			wantErr: "got 400",
		},
		{
			msg: "ok",
			d: newMultiDoer([]resp{
				{http.StatusOK, `{"users": [{"email": "foo@bar.com"}]}`},
				{http.StatusCreated, "_"},
			}),
		},
	} {
		t.Run(test.msg, func(t *testing.T) {
			c := &clientImpl{test.d, "_", _headers, map[string]string{}}
			err := c.CreateOverride("_", stickyshift.Shift{Email: "foo@bar.com"})
			if test.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), test.wantErr)
				return
			}
			require.NoError(t, err)
		})
	}
}
