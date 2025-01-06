package archer

import (
	"context"
	"time"
)

func newReaper(queue *Queue, every time.Duration, timeout time.Duration) *reaper {
	return &reaper{
		queue:   *queue,
		ticker:  time.NewTicker(every),
		timeout: timeout,
	}
}

type reaper struct {
	queue   Queue
	ticker  *time.Ticker
	timeout time.Duration
}

func (r *reaper) Run(ctx context.Context, errChan chan<- error) {
	for {
		select {
		case <-ctx.Done():
			return
		case <-r.ticker.C:
			if err := r.queue.RequeueTimeout(ctx, r.timeout); err != nil {
				errChan <- err
			}
		}
	}
}
