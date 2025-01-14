package archer

import (
	"time"
)

type registerConfig struct {
	w         Worker
	instances int
	timeout   time.Duration
}

type register map[string]registerConfig

func newRegister() register {
	return register{}
}

func (r register) registerWorker(name string, w Worker, opts ...WorkerOptionFunc) {
	rc := registerConfig{
		w:         w,
		timeout:   1 * time.Minute,
		instances: 1,
	}

	for _, opt := range opts {
		rc = opt(rc)
	}

	r[name] = rc
}

func (r register) Register(name string, w WorkerFn, opts ...WorkerOptionFunc) {
	r.registerWorker(name, &fnWorker{fn: w}, opts...)
}

func (r register) getWorkers() map[string]registerConfig {
	return r
}
