package archer

import (
	"bytes"
	"context"
	"database/sql"
	"log/slog"
	"sync"
	"time"

	"github.com/dyaksa/archer/job"
	"github.com/dyaksa/archer/store"
	"golang.org/x/sync/errgroup"

	_ "github.com/lib/pq"
)

type Options struct {
	Addr     string
	User     string
	Password string
	SSL      string
	DBName   string

	MaxIdleConns int
	MaxOpenConns int
}

type Client struct {
	wrapper   wrapperTx
	tx        func(*sql.Tx) Tx
	tableName string

	register
	spawn  Spawner
	mutate *Mutate

	errChan    chan error
	errHandler func(error)
	shutdown   func()

	sleepInterval  time.Duration
	reaperInterval time.Duration

	queue      func(name string) *Queue
	coRoutines []func() error
}

type wrapperTx interface {
	WrapTx(ctx context.Context, fn func(ctx context.Context, tx *sql.Tx) (any, error)) (any, error)
}

func NewClient(opt *Options, options ...ClientOptionFunc) *Client {
	dsn := bytes.Buffer{}
	dsn.WriteString("postgres://")
	dsn.WriteString(opt.User)
	dsn.WriteString(":")
	dsn.WriteString(opt.Password)
	dsn.WriteString("@")
	dsn.WriteString(opt.Addr)
	dsn.WriteString("/")
	dsn.WriteString(opt.DBName)
	dsn.WriteString("?sslmode=disable")

	db, err := sql.Open("postgres", dsn.String())
	if err != nil {
		panic(err)
	}

	db.SetMaxIdleConns(opt.MaxIdleConns)
	db.SetMaxOpenConns(opt.MaxOpenConns)

	errChan := make(chan error)
	ctx, cancel := context.WithCancel(context.Background())

	c := &Client{}
	c.sleepInterval = time.Second * 2   // default sleepinterval
	c.reaperInterval = time.Second * 10 // default reaper interval
	c.errHandler = defaultErrorHandler  // default errhandler
	c.tableName = "jobs"                // sleep tableName

	for _, opt := range options {
		c = opt(c)
	}

	c.register = newRegister()
	c.wrapper = store.NewWrapperTx(db)
	c.spawn = newSpawner(ctx, errChan)
	c.errChan = errChan
	c.shutdown = cancel
	c.tx = func(tx *sql.Tx) Tx {
		return newTx(tx, c.tableName)
	}
	c.queue = func(name string) *Queue {
		return NewQueue(db, name, c.tableName)
	}

	c.mutate = newMutate(db, c.tableName)

	return c
}

func (c *Client) WithTx(tx *sql.Tx) Tx {
	return c.tx(tx)
}

func (c *Client) Schedule(ctx context.Context, id string, queueName string, arguments interface{}, options ...FnOptions) (any, error) {
	return c.wrapper.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, c.tx(tx).Schedule(ctx, id, queueName, arguments, options...)
	})
}

func (c *Client) Cancel(ctx context.Context, id string) (any, error) {
	return c.wrapper.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, c.tx(tx).Cancel(ctx, id)
	})
}

func (c *Client) ScheduleNow(ctx context.Context, id string) (any, error) {
	return c.wrapper.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return nil, c.tx(tx).ScheduleNow(ctx, id)
	})
}

func (c *Client) Get(ctx context.Context, id string) (any, error) {
	res, err := c.wrapper.WrapTx(ctx, func(ctx context.Context, tx *sql.Tx) (any, error) {
		return c.tx(tx).Get(ctx, id)
	})
	if err != nil {
		return nil, err
	}

	return res.(*job.Job), nil
}

func (c *Client) Stop() {
	c.spawn.Shutdown()
}

func (c *Client) Start() error {
	g := new(errgroup.Group)

	for _, co := range c.coRoutines {
		g.Go(co)
	}

	g.Go(func() error {
		c.start()
		return nil
	})

	return g.Wait()
}

func (c *Client) start() {
	errwg := &sync.WaitGroup{}
	errwg.Add(1)

	go errorRoutine(c.errChan, c.errHandler, errwg)

	for name, config := range c.register.getWorkers() {
		q := c.queue(name)

		for i := 0; i < config.instances; i++ {
			s := newPool(q, c.mutate, config.w, c.sleepInterval, config.callbackSuccess, config.callbackFailed)
			c.spawn.Spawn(s)
		}

		r := newReaper(q, c.reaperInterval, config.timeout)
		c.spawn.Spawn(r)
	}

	c.spawn.Wait()

	close(c.errChan)

	errwg.Wait()
}

func errorRoutine(errChan <-chan error, errorHandler func(error), wg *sync.WaitGroup) {
	defer wg.Done()
	for err := range errChan {
		errorHandler(err)
	}
}

func defaultErrorHandler(err error) {
	slog.Info("error in worker pool", "err", err)
}
