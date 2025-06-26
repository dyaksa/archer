# Options

Archer can be customized with a set of options that control worker behaviour, retry logic and other aspects.

## Worker Registration Options

- `WithInstances(n int)` – number of concurrent workers for a job type.
- `WithTimeout(d time.Duration)` – job timeout duration.
- `WithRetryInterval(d time.Duration)` – wait time before retrying a failed job.
- `WithMaxRetries(n int)` – maximum number of retry attempts.

## Client Options

When creating a client, you can supply database connection information and other parameters:

```go
c := archer.NewClient(&archer.Options{
    Addr:     "localhost:5432",
    User:     "admin",
    Password: "password",
    DBName:   "core",
}, archer.WithSetTableName("outbox"))
```

- `Addr` – PostgreSQL host and port.
- `User` – database user.
- `Password` – user's password.
- `DBName` – database name.
- `WithSetTableName(name string)` – store jobs in a custom table.
- `WithSleepInterval(d time.Duration)` – delay between polling cycles for new jobs.
- `WithReaperInterval(d time.Duration)` – interval for cleaning up finished or dead jobs.
- `WithErrHandler(func(error))` – custom error handler for worker errors.

