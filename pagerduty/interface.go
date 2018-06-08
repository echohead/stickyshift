package pagerduty

import (
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
	}

	// Schedule holds only the needed fields of a pagerduty schedule
	Schedule struct {
		Id   string `json:"id"`
		Name string `json:"name"`
	}
)
