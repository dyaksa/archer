package store

import (
	"strings"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/types"
)

var (
	columns = []string{
		"id",
		"queue_name",
		"status",
		"last_error",
		"retry_count",
		"max_retry",
		"arguments",
		"result",
		"retry_interval",
		"scheduled_at",
		"started_at",
		"created_at",
		"updated_at",
	}

	entryFields = strings.Join(columns, ", ")
)

type entity struct {
	ID            string           `json:"id"`
	QueueName     string           `json:"queue_name"`
	Status        string           `json:"status"`
	LastError     types.NullString `json:"last_error"`
	RetryCount    int              `json:"retry_count"`
	MaxRetry      int              `json:"max_retry"`
	Arguments     []byte           `json:"arguments"`
	Result        []byte           `json:"result"`
	RetryInterval time.Duration    `json:"retry_interval"`
	ScheduledAt   types.NullTime   `json:"scheduled_at"`
	StartedAt     types.NullTime   `json:"started_at"`
	CreatedAt     time.Time        `json:"created_at"`
	UpdatedAt     time.Time        `json:"updated_at"`
}

func (e *entity) To() *job.Job {
	return &job.Job{
		ID:            e.ID,
		QueueName:     e.QueueName,
		Status:        e.Status,
		LastError:     e.LastError.String,
		RetryCount:    e.RetryCount,
		MaxRetry:      e.MaxRetry,
		Arguments:     e.Arguments,
		Result:        e.Result,
		RetryInterval: e.RetryInterval,
		ScheduleAt:    e.ScheduledAt.Time,
		StartedAt:     e.StartedAt.Time,
		CreatedAt:     e.CreatedAt,
		UpdadatedAt:   e.UpdatedAt,
	}
}

func (e *entity) ScanDestinations() []interface{} {
	return []interface{}{
		&e.ID,
		&e.QueueName,
		&e.Status,
		&e.LastError,
		&e.RetryCount,
		&e.MaxRetry,
		&e.Arguments,
		&e.Result,
		&e.RetryInterval,
		&e.ScheduledAt,
		&e.StartedAt,
		&e.CreatedAt,
		&e.UpdatedAt,
	}
}
