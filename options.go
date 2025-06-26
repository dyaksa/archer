package archer

import (
	"time"

	"github.com/dyaksa/archer/job"
)

type FnOptions func(j job.Job) job.Job

func WithMaxRetries(max int) FnOptions {
	return func(j job.Job) job.Job {
		j.MaxRetry = max
		return j
	}
}

func WithRetryInterval(interval time.Duration) FnOptions {
	return func(j job.Job) job.Job {
		j.RetryInterval = interval
		return j
	}
}

func WithScheduleTime(t time.Time) FnOptions {
	return func(j job.Job) job.Job {
		j.ScheduleAt = t
		return j
	}
}

type WorkerOptionFunc func(registerConfig) registerConfig

func WithTimeout(t time.Duration) WorkerOptionFunc {
	return func(r registerConfig) registerConfig {
		r.timeout = t
		return r
	}
}

func WithInstances(i int) WorkerOptionFunc {
	return func(r registerConfig) registerConfig {
		r.instances = i
		return r
	}
}

type ClientOptionFunc func(*Client) *Client

func WithSetTableName(table string) ClientOptionFunc {
	return func(c *Client) *Client {
		c.tableName = table
		return c
	}
}

func WithSleepInterval(t time.Duration) ClientOptionFunc {
	return func(c *Client) *Client {
		c.sleepInterval = t
		return c
	}
}

func WithReaperInterval(t time.Duration) ClientOptionFunc {
	return func(c *Client) *Client {
		c.reaperInterval = t
		return c
	}
}

func WithErrHandler(fn func(error)) ClientOptionFunc {
	return func(c *Client) *Client {
		c.errHandler = fn
		return c
	}
}
