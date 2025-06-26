# Usage

Archer exposes a `Client` that manages workers and job scheduling.

## Worker Example

Below is a simplified worker that processes `call_api` jobs. Each job receives a context and a `job.Job` from which you can parse arguments:

```go
func CallClient(ctx context.Context, job job.Job) (any, error) {
    args := CallApiArgs{}
    if err := job.ParseArguments(&args); err != nil {
        return nil, err
    }

    // perform work here
    return CallApiResults{StatusCode: 200}, nil
}

func main() {
    c := archer.NewClient(&archer.Options{
        Addr:     "localhost:5432",
        Password: "password",
        User:     "admin",
        DBName:   "core",
    })

    c.Register("call_api", CallClient,
        archer.WithInstances(1),
        archer.WithTimeout(1*time.Second),
    )

    if err := c.Start(); err != nil {
        panic(err)
    }
}
```

## Client Example

Jobs can be enqueued from anywhere using the same client:

```go
func main() {
    c := archer.NewClient(&archer.Options{
        Addr:     "localhost:5432",
        Password: "password",
        User:     "admin",
        DBName:   "core",
    })

    _, err := c.Schedule(
        context.Background(),
        uuid.NewString(),
        "call_api",
        CallApiArgs{URL: "http://localhost:3001/v4/upsert", Method: "POST"},
        archer.WithMaxRetries(3),
    )
    if err != nil {
        panic(err)
    }
}
```

