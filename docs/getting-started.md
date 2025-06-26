# Getting Started

This guide walks you through installing Archer and running the example worker and client.

## Installation

Archer requires **Go 1.20+** and **PostgreSQL 12+**.

Install the package:

```bash
go get github.com/dyaksa/archer
```

Ensure your PostgreSQL database is running and that you have created the `jobs` table. A minimal schema is shown below:

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

## Example

The `example` directory contains a sample worker and client. Start the worker:

```bash
go run ./example/worker
```

In another terminal enqueue jobs with:

```bash
go run ./example/client
```

Both programs demonstrate registering jobs and scheduling them for execution.

