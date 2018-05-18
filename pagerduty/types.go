package pagerduty

import (
	"time"

	"github.com/echohead/stickyshift"
)

func New() (Client, error) {
	return newClientImpl()
}

type (
	// Client writes and reads to/from pagerduty API
	Client interface {
		Sync(string, stickyshift.ShiftList) error
		GetSchedule(string) (Schedule, error)
		GetOverrides(string, time.Time, time.Time) ([]Override, error)
		CreateOverride(string, stickyshift.Shift) error
	}

	// Schedule holds only the needed fields of a pagerduty schedule
	Schedule struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}

	// Override represents a pagerduty override
	Override struct {
		User  userRef   `json:"user"`
		Start time.Time `json:"start"`
		End   time.Time `json:"end"`
	}

	// User holds only the needed fields of a pagerduty user
	User struct {
		Id    string `json:"id"`
		Name  string `json:"name"`
		Email string `json:"email"`
	}

	getScheduleResponse struct {
		Schedule Schedule `json:"schedule"`
	}

	getOverridesResponse struct {
		Overrides []Override `json:"overrides"`
	}

	getUsersResponse struct {
		Users []User `json:"users"`
	}

	userRef struct {
		Id   string `json:"id"`
		Type string `json:"type"`
		Name string `json:"summary"`
	}
)
