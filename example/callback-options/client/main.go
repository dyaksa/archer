package main

import (
	"context"
	"log/slog"
	"time"

	"github.com/dyaksa/archer"
	"github.com/google/uuid"
)

type CallTestCallbackArgs struct {
	URL    string
	Method string
	Body   any
}

type DataDto struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

func main() {
	c := archer.NewClient(&archer.Options{
		Addr:     "localhost:5433",
		Password: "azozink",
		User:     "postgres",
		DBName:   "local_fabd_revenue",
	}, archer.WithSetTableName("jobs"))

	dto := DataDto{
		Name:     "John",
		Email:    "sample@gmail.com",
		Password: "sample123",
	}

	// Callback Failed
	_, err := c.Schedule(
		context.Background(),
		uuid.NewString(),
		"call_test_callback",
		CallTestCallbackArgs{URL: "http://localhost:3001/v4/upsert", Method: "POST", Body: dto},
		archer.WithMaxRetries(3),
		archer.WithRetryInterval(2*time.Second),
	)

	if err != nil {
		slog.Error("error", "err", err)
		return
	}

	// Callback Success
	_, err2 := c.Schedule(
		context.Background(),
		uuid.NewString(),
		"call_test_callback",
		CallTestCallbackArgs{URL: "http://localhost:3001/v4/upsert", Method: "PUT", Body: dto},
		archer.WithMaxRetries(3),
		archer.WithRetryInterval(2*time.Second),
	)

	if err2 != nil {
		slog.Error("error", "err", err2)
		return
	}

	slog.Info("done")
}
