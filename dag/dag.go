package dag

import (
	"context"
	"fmt"

	"golang.org/x/sync/errgroup"
)

type NodeFn func(ctx context.Context, input any) (any, error)

type ConditionFn func(input any) bool

type Edge struct {
	To        string
	Condition ConditionFn
	Foreach   bool
}

type DAG struct {
	nodes map[string]*Node
	start string
}

type Node struct {
	ID     string
	Run    NodeFn
	SubDag *DAG
	Edges  []Edge
}

func New(start *Node) *DAG {
	d := &DAG{nodes: map[string]*Node{}, start: start.ID}
	d.AddNode(start)
	return d
}

func (d *DAG) AddNode(n *Node) {
	if d.nodes == nil {
		d.nodes = make(map[string]*Node, 0)
	}

	d.nodes[n.ID] = n
}

func (d *DAG) Execute(ctx context.Context, input any) (any, error) {
	return d.executeFrom(ctx, d.start, input)
}

func (d *DAG) executeFrom(ctx context.Context, id string, input any) (any, error) {
	node, ok := d.nodes[id]
	if !ok {
		return nil, fmt.Errorf("node %s not found", id)
	}

	var data = input
	var err error

	if node.SubDag != nil {
		data, err = node.SubDag.Execute(ctx, data)
		if err != nil {
			return nil, err
		}
	} else if node.Run != nil {
		data, err = node.Run(ctx, data)
		if err != nil {
			return nil, err
		}
	}

	if len(node.Edges) == 0 {
		return data, nil
	}

	var edges []Edge
	for _, e := range node.Edges {
		if e.Condition != nil && !e.Condition(data) {
			continue
		}

		edges = append(edges, e)
	}

	if len(edges) == 0 {
		return data, nil
	}

	resCh := make(chan any, len(edges))
	var g errgroup.Group
	for _, edge := range edges {
		e := edge
		g.Go(func() error {
			if e.Foreach {
				items, ok := data.([]any)
				if !ok {
					return fmt.Errorf("data is not a slice")
				}

				var out any
				for _, item := range items {
					var err error
					out, err = d.executeFrom(ctx, e.To, item)
					if err != nil {
						return err
					}
				}
				resCh <- out
				return nil
			}

			out, err := d.executeFrom(ctx, e.To, data)
			if err != nil {
				return err
			}

			resCh <- out
			return nil
		})
	}

	if err := g.Wait(); err != nil {
		return nil, err
	}

	close(resCh)

	var last any
	for out := range resCh {
		last = out
	}

	return last, nil
}
