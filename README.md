# Archer - Golang, Simple Job Queue Postgresql

Archer is a simple, lightweight queue/job worker that uses PostgreSQL as its backing store. It aims to provide an easy-to-use API for defining, enqueuing, and executing background jobs in your Go applications.

## Features

- `Simple API` – Define and register your jobs easily.

- `Backed by PostgreSQL` – Reliable storage and concurrency control.

- `Flexible Worker Pool` – Customize worker concurrency (instances) and job timeouts as needed.

- `Context Awareness` – Each job runs with a context, enabling you to cancel or time out jobs gracefully.

## Installation

```
go get github.com/dyaksa/archer
```

### Requirements

- Go 1.20+ (or newer)
- PostgreSQL 12+ (or newer)

### Database Setup

Make sure you have a PostgreSQL database up and running. You’ll need to provide the connection details (address, user, password, etc.) in the archer.Options.

Archer relies on a table structure to manage jobs. Ensure you have run the migration script (if provided), or set up your table accordingly.

## How It Works

archer under the hood uses a table like the following.

```sql
CREATE TABLE jobs (
  id varchar primary key,
  queue_name varchar not null,
  status varchar not null,
  arguments jsonb not null default '{}'::jsonb,
  result jsonb not null default '{}'::jsonb,
  last_error varchar,
  retry_count integer not null default 0,
  max_retry integer not null default 0,
  retry_interval integer not null default 0,
  scheduled_at timestamptz default now(),
  started_at timestamptz,
  created_at timestamptz not null default now(),
  updated_at timestamptz not null default now()
);

CREATE INDEX ON jobs (queue_name);
CREATE INDEX ON jobs (scheduled_at);
CREATE INDEX ON jobs (status);
CREATE INDEX ON jobs (started_at);
```

## Usage

### Importing the package

This package can be used by adding the following import statement to your `.go` files.

```go
import "github.com/dyaksa/archer"
```

### Worker Example

Below is a complete example of a worker that processes jobs named `call_api`. The job is expected to have `CallApiArgs` as arguments and returns a `status code`.

```go
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
```

### Client Example (Enqueuing Jobs)

To enqueue a job for processing, create or import the same archer.Client in a different part of your code or even a different service. Then call something like:

```go
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
		Email:    "sample@mail.com",
		Password: "sample123",
	}


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

	slog.Info("done")
}
```

## Options

You can configure the Archer client with various options:

- `WithInstances(n int)`
  Sets the number of concurrent workers that will process a particular job type.

- `WithTimeout(d time.Duration)`
  Sets a timeout for each job. If the job does not complete within this duration, it is considered failed/cancelled.

- `archer.NewClient(&archer.Options{ ... })`
  - `Addr`: PostgreSQL host and port (e.g., localhost:5432)
  - `User`: DB user
  - `Password`: DB user’s password
  - `DBName`: Database name

## Contributing

- Fork the repository.
- Create a new branch for your feature or bugfix.
- Commit your changes and push them to your fork.
- Create a Pull Request describing your changes.

We appreciate all contributions, whether they are documentation improvements, bug fixes, or new features!

## License

```
Copyright 2024

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

       http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```
