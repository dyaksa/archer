package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/dyaksa/archer"
	"github.com/dyaksa/archer/job"
)

type CallApiResults struct {
	StatusCode int `json:"status_code"`
}

type CallApiArgs struct {
	URL    string
	Method string
	Body   any
}

func CallClient(ctx context.Context, job job.Job) (any, error) {
	args := CallApiArgs{}
	if err := job.ParseArguments(&args); err != nil {
		return nil, err
	}

	slog.Info("started job request id: " + job.ID)
	defer func() {
		slog.Info("finished job request id: " + job.ID)
	}()

	client := &http.Client{
		Timeout: 10 * time.Second,
	}

	b, err := json.Marshal(args.Body)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(args.Method, args.URL, bytes.NewBuffer(b))
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	res := CallApiResults{StatusCode: resp.StatusCode}

	return res, nil
}

func main() {
	c := archer.NewClient(&archer.Options{
		Addr:     "localhost:5432",
		Password: "password",
		User:     "admin",
		DBName:   "webhooks",
	})

	c.Register("call_api",
		CallClient,
		archer.WithInstances(1),
		archer.WithTimeout(1*time.Second),
	)

	if err := c.Start(); err != nil {
		panic(err)
	}
}
