# DAG

Archer includes a small package for composing tasks as a directed acyclic graph (DAG). Each node represents a unit of work and can depend on the output of previous nodes.

## Creating a DAG

A DAG is constructed from nodes. Each node has an ID, a `Run` function, optional edges to other nodes and may reference a sub-DAG.

```go
n1 := &dag.Node{ID: "start", Run: func(ctx context.Context, in any) (any, error) {
    // do something
    return in, nil
}}

n2 := &dag.Node{ID: "finish", Run: func(ctx context.Context, in any) (any, error) {
    return in, nil
}}

n1.Edges = []dag.Edge{{To: "finish"}}

flow := dag.New(n1)
flow.AddNode(n2)
```

Execute the graph with:

```go
out, err := flow.Execute(ctx, input)
```

Edges may have a condition or be executed for each item in a slice. See the `dag` package for more details.

