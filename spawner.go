package archer

import (
	"context"
	"sync"
)

func newSpawner(ctx context.Context, errChan chan<- error) *spawn {
	ctx, cancel := context.WithCancel(ctx)
	return &spawn{
		wg:       &sync.WaitGroup{},
		ctx:      ctx,
		shutdown: cancel,
		errChan:  errChan,
	}
}

type Spawner interface {
	Spawn(runner)
	Wait()
	Shutdown()
	Done() <-chan struct{}
}

type spawn struct {
	wg *sync.WaitGroup

	ctx      context.Context
	shutdown func()
	errChan  chan<- error
}

type runner interface {
	Run(ctx context.Context, errChan chan<- error)
}

func (s *spawn) Spawn(runner runner) {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		runner.Run(s.ctx, s.errChan)
	}()
}

func (s *spawn) Wait() {
	s.wg.Wait()
}

func (s *spawn) Shutdown() {
	s.shutdown()
}

func (s *spawn) Done() <-chan struct{} {
	c := make(chan struct{})
	go func() {
		<-s.ctx.Done()
		c <- struct{}{}
	}()

	return c
}
