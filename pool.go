package archer

import (
	"context"
	"time"

	"github.com/dyaksa/archer/job"
)

type pool struct {
	queue         Queue
	handler       Handler
	sleepInterval time.Duration
}

func newPool(q *Queue, m mutate, w Worker, sleepInterval time.Duration, callbackSuccess func(ctx context.Context, job job.Job, res any) (any, error), callbackFailed func(ctx context.Context, job job.Job, err error) (any, error)) *pool {
	return &pool{
		queue:         *q,
		handler:       newHandler(w, m, callbackSuccess, callbackFailed),
		sleepInterval: sleepInterval,
	}
}

func (p *pool) Run(ctx context.Context, errChan chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			j, err := p.queue.Poll(ctx)
			if err == job.ErrorJobNotFound {
				time.Sleep(p.sleepInterval)
				continue
			}

			if err != nil {
				errChan <- err
				continue
			}

			if err := p.handler.Handle(ctx, *j); err != nil {
				errChan <- err
				continue
			}
		}
	}
}
