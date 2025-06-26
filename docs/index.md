# Archer Documentation

Welcome to the Archer job queue library documentation. Archer is a lightweight background job system for Go applications that uses PostgreSQL as its backing store.

## Features

- **Simple API** – Register jobs and enqueue tasks with minimal boilerplate.
- **Backed by PostgreSQL** – Reliable storage with transactional guarantees.
- **Flexible Worker Pool** – Control worker concurrency and job timeouts.
- **Context Support** – Cancel or time out jobs gracefully using contexts.
- **Custom Table Name** – Store jobs in a table of your choosing.
- **DAG Package** – Build complex workflows using directed acyclic graphs.

## Getting Started

See [Getting Started](getting-started.md) for installation instructions and a quick example.

Additional topics:

- [Usage](usage.md) – Worker and client examples.
- [Options](options.md) – Configuration options for workers and jobs.
- [DAG](dag.md) – Using Archer's DAG utilities for complex workflows.

