package pagerduty

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"time"

	"github.com/echohead/stickyshift"
)

type (
	clientImpl struct {
		doer
		url     string
		headers map[string]string
		userIds map[string]string
	}

	doer interface {
		Do(*http.Request) (*http.Response, error)
	}

	// override represents a pagerduty override
	override struct {
		User  userRef   `json:"user"`
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// user holds only the needed fields of a pagerduty user
	user struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	getScheduleResponse struct {
		Schedule Schedule `json:"schedule"`
	}

	getOverridesResponse struct {
		Overrides []override `json:"overrides"`
	}

	getUsersResponse struct {
		Users []user `json:"users"`
	}

	userRef struct {
		Id   string `json:"id"`
		Type string `json:"type"`
		Name string `json:"summary"`
	}
)

const (
	_getOverridesTimeFmt = "2006-01-02"
	_pdUrl               = "https://api.pagerduty.com"
	_tokenEnvVar         = "PD_TOKEN"
)

func newClientImpl() (Client, error) {
	token := os.Getenv(_tokenEnvVar)
	if token == "" {
		return nil, fmt.Errorf("environment variable $%s must be set", _tokenEnvVar)
	}

	return &clientImpl{
		&http.Client{},
		_pdUrl,
		map[string]string{
			"Authorization": token,
			"Accept":        "application/vnd.pagerduty+json;version=2",
			"Content-Type":  "application/json",
		},
		map[string]string{},
	}, nil
}

func (c *clientImpl) Sync(sid string, shifts stickyshift.ShiftList) error {
	if len(shifts) < 1 {
		return nil
	}
	if _, err := c.GetSchedule(sid); err != nil {
		return err
	}

	os, err := c.getOverrides(sid, shifts[0].Start, shifts[len(shifts)-1].End)
	if err != nil {
		return err
	}

	for _, shift := range shifts {
		if shift.End.Before(time.Now()) {
			continue
		}
		exists, err := c.overrideExists(os, shift)
		if err != nil {
			return err
		}
		if exists {
			continue
		}

		err = c.createOverride(sid, shift)
		if err != nil {
			return err
		}
	}

	return nil
}

func (c *clientImpl) GetSchedule(id string) (Schedule, error) {
	resp := &getScheduleResponse{}
	if err := c.get(fmt.Sprintf("/schedules/%v", id), &resp); err != nil {
		return Schedule{}, err
	}
	return resp.Schedule, nil
}

func (c *clientImpl) getUser(email string) (user, error) {
	resp := &getUsersResponse{}
	if err := c.get(fmt.Sprintf("/users?query=%s", url.QueryEscape(email)), resp); err != nil {
		return user{}, err
	}
	if len(resp.Users) != 1 {
		return user{}, fmt.Errorf("expected one user for %q, found %v", email, len(resp.Users))
	}
	if resp.Users[0].Email != email {
		return user{}, fmt.Errorf("got user with email %q, expected %q", resp.Users[0].Email, email)
	}
	return resp.Users[0], nil
}

func (c *clientImpl) getOverrides(sid string, start, end time.Time) ([]override, error) {
	resp := &getOverridesResponse{}
	t0 := start.Format(_getOverridesTimeFmt)
	t1 := end.Format(_getOverridesTimeFmt)
	url := fmt.Sprintf("/schedules/%v/overrides?since=%v&until=%v", sid, t0, t1)
	if err := c.get(url, &resp); err != nil {
		return nil, err
	}
	return resp.Overrides, nil
}

func (c *clientImpl) createOverride(sid string, shift stickyshift.Shift) error {
	uid, err := c.userId(shift.Email)
	if err != nil {
		return err
	}
	o := override{
		User: userRef{
			Id:   uid,
			Type: "user_reference",
		},
		Start: shift.Start,
		End:   shift.End,
	}
	return c.post(fmt.Sprintf("/schedules/%s/overrides", sid), o)
}

func (c *clientImpl) overrideExists(os []override, shift stickyshift.Shift) (bool, error) {
	uid, err := c.userId(shift.Email)
	if err != nil {
		return false, err
	}
	for _, o := range os {
		if o.Start.Equal(shift.Start) &&
			o.End.Equal(shift.End) &&
			o.User.Id == uid {
			return true, nil
		}
	}
	return false, nil
}

func (c *clientImpl) userId(email string) (string, error) {
	if id, ok := c.userIds[email]; ok {
		return id, nil
	}
	user, err := c.getUser(email)
	if err != nil {
		return "", err
	}
	c.userIds[email] = user.Id
	return user.Id, nil
}

func (c *clientImpl) setHeaders(r *http.Request) {
	for k, v := range c.headers {
		r.Header.Set(k, v)
	}
}

func (c *clientImpl) request(method, path string, wantStatus int, body io.Reader) ([]byte, error) {
	req, err := http.NewRequest(method, c.url+path, body)
	if err != nil {
		return nil, err
	}
	c.setHeaders(req)
	resp, err := c.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	bs, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != wantStatus {
		return nil, fmt.Errorf("expected %v response for %v, got %v: %v", wantStatus, path, resp.StatusCode, string(bs))
	}
	return bs, nil
}

func (c *clientImpl) get(path string, into interface{}) error {
	bs, err := c.request(http.MethodGet, path, http.StatusOK, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(bs, &into)
}

func (c *clientImpl) post(path string, body interface{}) error {
	bs, err := json.Marshal(body)
	if err != nil {
		return err
	}
	_, err = c.request(http.MethodPost, path, http.StatusCreated, bytes.NewReader(bs))
	return err
}
