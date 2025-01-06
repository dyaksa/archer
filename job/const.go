package job

import "errors"

const (
	StatusCompleted   = "completed"
	StatusScheduled   = "scheduled"
	StatusFailed      = "failed"
	StatusCanceled    = "canceled"
	StatusInitialized = "initialized"
)

var (
	ErrorJobNotFound = errors.New("job not found")
	ErrorJobCanceled = "job canceled"
	ErrorJobFailed   = "job failed"
	ErrorJobTimeout  = "job timeout"
	ErrorJobUnknown  = "job unknown"
)
