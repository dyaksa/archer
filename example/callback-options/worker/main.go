package main

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/dyaksa/archer"
	"github.com/dyaksa/archer/job"
)

type CallTestCallbackArgs struct {
	URL    string
	Method string
	Body   any
}

type CallTestCallbackResults struct {
	Message string `json:"message"`
}

func CallTestCallback(ctx context.Context, job job.Job) (any, error) {
	args := CallTestCallbackArgs{}
	if err := job.ParseArguments(&args); err != nil {
		return nil, err
	}

	time.Sleep(10 * time.Second)

	slog.Info("started job request id: " + job.ID)
	defer func() {
		slog.Info("finished job request id: " + job.ID)
	}()

	if args.Method == "POST" {
		return nil, fmt.Errorf("invalid method")
	}

	res := CallTestCallbackResults{Message: "Success!"}

	return res, nil
}

func CallTestCallbackSuccess(ctx context.Context, job job.Job, res any) (any, error) {
	slog.Info("Job completed successfully", "job_id", job.ID)

	slog.Info("RESULT SUCCESS", "result", res)

	return map[string]any{
		"callback_executed": true,
		"job_id":            job.ID,
		"completed_at":      time.Now(),
	}, nil
}

func CallTestCallbackFailed(ctx context.Context, job job.Job, err error) (any, error) {
	slog.Info("Job failed", "job_id", job.ID)
	slog.Error("Job failed", "error", err.Error())

	return map[string]any{
		"callback_executed": true,
		"job_id":            job.ID,
		"retry_count":       job.RetryCount,
		"failed_at":         time.Now(),
	}, nil
}

func main() {
	c := archer.NewClient(&archer.Options{
		Addr:     "localhost:5433",
		Password: "azozink",
		User:     "postgres",
		DBName:   "local_fabd_revenue",
	}, archer.WithSetTableName("jobs"))

	slog.Info("waiting for jobs")

	c.Register("call_test_callback", CallTestCallback,
		archer.WithInstances(1),
		archer.WithTimeout(1*time.Second),
		archer.WithCallbackSuccess(CallTestCallbackSuccess),
		archer.WithCallbackFailed(CallTestCallbackFailed),
	)

	if err := c.Start(); err != nil {
		panic(err)
	}
}
