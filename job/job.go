package job

import (
	"database/sql/driver"
	"time"

	"github.com/goccy/go-json"

	"github.com/dyaksa/archer/types"
)

type Job struct {
	ID            string          `json:"id"`
	QueueName     string          `json:"queue_name"`
	Status        string          `json:"status"`
	LastError     string          `json:"last_error"`
	RetryCount    int             `json:"retry_count"`
	MaxRetry      int             `json:"max_retry"`
	Arguments     json.RawMessage `json:"arguments"`
	Result        json.RawMessage `json:"result"`
	RetryInterval time.Duration   `json:"retry_interval"`
	ScheduleAt    time.Time       `json:"scheduled_at"`
	StartedAt     types.NullTime  `json:"started_at"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdadatedAt   time.Time       `json:"updated_at"`
}

func (j *Job) ParseArguments(v interface{}) error {
	return json.Unmarshal([]byte(j.Arguments), v)
}

func (j *Job) ShouldRetry() bool {
	return j.RetryCount < j.MaxRetry
}

func (j *Job) ScheduleRetry(t time.Time) Job {
	j.RetryCount++
	j.ScheduleAt = t.Add(j.RetryInterval)
	j.Status = StatusScheduled
	return *j
}

func (j *Job) SetStatus(status string) Job {
	j.Status = status
	return *j
}

func (j *Job) SetLastError(err error) Job {
	j.LastError = err.Error()
	return *j
}

func (j *Job) SetResult(v interface{}) (Job, error) {
	if v == nil {
		return *j, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return *j, err
	}

	j.Result = b
	return *j, nil
}

func (j *Job) SetArgs(v interface{}) (Job, error) {
	if v == nil {
		return *j, nil
	}

	b, err := json.Marshal(v)
	if err != nil {
		return *j, err
	}

	j.Arguments = b
	return *j, nil
}

func (j *Job) Value() (driver.Value, error) {
	return json.Marshal(j)
}

func (j *Job) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), j)
}
