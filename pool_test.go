package archer

import (
	"context"
	"database/sql"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/dyaksa/archer/job"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockTx is a mock type for the Tx interface
type MockTx struct {
	mock.Mock
}

func (m *MockTx) Schedule(ctx context.Context, id string, queueName string, args_any any, options ...FnOptions) error {
	optionsAny := make([]any, len(options))
	for i, opt := range options {
		optionsAny[i] = opt
	}
	allArgs := append([]any{ctx, id, queueName, args_any}, optionsAny...)
	calledArgs := m.Called(allArgs...)
	return calledArgs.Error(0)
}

func (m *MockTx) Poll(ctx context.Context, queueName string) (*job.Job, error) {
	args := m.Called(ctx, queueName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*job.Job), args.Error(1)
}

func (m *MockTx) Update(ctx context.Context, j job.Job) error {
	args := m.Called(ctx, j)
	return args.Error(0)
}

func (m *MockTx) RequeueTimeout(ctx context.Context, queueName string, startedBefore time.Time) error {
	args := m.Called(ctx, queueName, startedBefore)
	return args.Error(0)
}

func (m *MockTx) Cancel(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTx) ScheduleNow(ctx context.Context, id string) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *MockTx) Get(ctx context.Context, id string) (*job.Job, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*job.Job), args.Error(1)
}

// MockHandler is a mock type for the Handler interface
type MockHandler struct {
	mock.Mock
	handleCalled chan bool // To signal that Handle was called
	handlingDone chan bool // To signal that Handle finished or was canceled
}

func NewMockHandler() *MockHandler {
	return &MockHandler{
		handleCalled: make(chan bool, 1),
		handlingDone: make(chan bool, 1),
	}
}

func (m *MockHandler) Handle(ctx context.Context, j job.Job) error {
	m.handleCalled <- true
	args := m.Called(ctx, j)

	// Simulate work that can be canceled
	select {
	case <-ctx.Done():
		m.handlingDone <- true
		return ctx.Err() // Return context error upon cancellation
	case <-time.After(5 * time.Millisecond): // Shortened duration
		m.handlingDone <- true
		return args.Error(0)
	}
}

func TestPool_Run_ContextCancellation(t *testing.T) {
	db, sqlMock, err := sqlmock.New(sqlmock.QueryMatcherOption(sqlmock.QueryMatcherRegexp))
	if err != nil {
		t.Fatalf("an error '%s' was not expected when opening a stub database connection", err)
	}
	defer db.Close()

	queueName := "test_context_cancel_queue"
	realQueue := NewQueue(db, queueName)
	mockTx := new(MockTx)
	realQueue.tx = func(dbtx *sql.Tx) Tx {
		return mockTx
	}

	mockHandler := NewMockHandler()

	p := newPool(realQueue, nil, nil, 10*time.Millisecond)
	p.handler = mockHandler

	ctx, cancel := context.WithCancel(context.Background())
	errChan := make(chan error, 1)

	testJob := &job.Job{ID: "test_job_id", QueueName: queueName}

	sqlMock.ExpectBegin()
	mockTx.On("Poll", mock.Anything, queueName).Return(testJob, nil).Once()
	sqlMock.ExpectCommit()

	mockTx.On("Poll", mock.Anything, queueName).Return(nil, job.ErrorJobNotFound).Maybe()

	mockHandler.On("Handle", mock.AnythingOfType("*context.cancelCtx"), *testJob).Return(context.Canceled).Once()

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		p.Run(ctx, errChan)
	}()

	select {
	case <-mockHandler.handleCalled:
		cancel()
	case <-time.After(2 * time.Second):
		t.Fatal("Handler.Handle was not called within timeout")
	}

	select {
	case <-mockHandler.handlingDone:
	case <-time.After(2 * time.Second):
		t.Fatal("Handler.Handle did not complete after context cancellation")
	}

	wg.Wait()

	mockHandler.AssertCalled(t, "Handle", mock.AnythingOfType("*context.cancelCtx"), *testJob)
	mockTx.AssertExpectations(t)
	assert.NoError(t, sqlMock.ExpectationsWereMet(), "sqlmock expectations were not met")

	select {
	case err := <-errChan:
		if !errors.Is(err, context.Canceled) {
			t.Logf("Received error from errChan: %v. Expected context.Canceled or nil.", err)
		}
	default:
	}
}
