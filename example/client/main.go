package main

import (
	"context"
	"log/slog"

	"github.com/dyaksa/archer"
	"github.com/google/uuid"
	"github.com/sourcegraph/conc/pool"
)

type CallApiArgs struct {
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
		Addr:     "localhost:5432",
		Password: "password",
		User:     "admin",
		DBName:   "webhooks",
	})

	dto := DataDto{
		Name:     "John",
		Email:    "sample@gmail.com",
		Password: "sample123",
	}

	p := pool.New().WithMaxGoroutines(10).WithErrors()

	p.Go(func() error {
		for i := 0; i < 1000; i++ {
			_, err := c.Schedule(
				context.Background(),
				uuid.NewString(),
				"call_api",
				CallApiArgs{URL: "http://localhost:3001/v4/upsert", Method: "POST", Body: dto},
				archer.WithMaxRetries(3),
			)

			if err != nil {
				return err
			}
		}

		return nil
	})

	if err := p.Wait(); err != nil {
		panic(err)
	}

	slog.Info("done")
}
