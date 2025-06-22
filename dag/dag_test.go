package dag

import (
	"context"
	"fmt"
	"log/slog"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDAGMultiNodes(t *testing.T) {
	t.Run("failed", func(t *testing.T) {
		visited := []string{}
		n1 := &Node{ID: "n1", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n1")
			slog.Info("n1")
			return input, nil
		}}

		n2 := &Node{ID: "n2", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n2")
			slog.Info("n2")
			return input, fmt.Errorf("failed")
		}}

		n3 := &Node{ID: "n3", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n3")
			slog.Info("n3")
			return input, nil
		}}

		n4 := &Node{ID: "n4", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n4")
			slog.Info("n4")
			return input, nil
		}}

		n1.Edges = []Edge{{To: "n2"}}
		n2.Edges = []Edge{{To: "n3"}}
		n3.Edges = []Edge{{To: "n4"}}

		d := New(n1)
		d.AddNode(n2)
		d.AddNode(n3)
		d.AddNode(n4)

		_, err := d.Execute(context.Background(), nil)
		assert.NotNil(t, err)
		assert.Equal(t, []string{"n1", "n2"}, visited)
	})

	t.Run("success", func(t *testing.T) {
		visited := []string{}
		n1 := &Node{ID: "n1", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n1")
			slog.Info("n1")
			return input, nil
		}}

		n2 := &Node{ID: "n2", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n2")
			slog.Info("n2")
			return input, nil
		}}

		n3 := &Node{ID: "n3", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n3")
			slog.Info("n3")
			return input, nil
		}}

		n4 := &Node{ID: "n4", Run: func(ctx context.Context, input any) (any, error) {
			visited = append(visited, "n4")
			slog.Info("n4")
			return input, nil
		}}

		n1.Edges = []Edge{{To: "n4"}, {To: "n2"}}
		n2.Edges = []Edge{{To: "n3"}}

		d := New(n1)
		d.AddNode(n2)
		d.AddNode(n3)
		d.AddNode(n4)

		_, err := d.Execute(context.Background(), nil)
		if err != nil {
			t.Fatal(err)
		}

		assert.NoError(t, err)
		assert.Equal(t, []string{"n1", "n2", "n3", "n4"}, visited)
	})
}

func TestSubDag(t *testing.T) {
	subVisited := []string{}
	subN1 := &Node{ID: "sub1", Run: func(ctx context.Context, input any) (any, error) {
		subVisited = append(subVisited, "sub1")
		slog.Info("sub1")
		return input, nil
	}}

	subN2 := &Node{ID: "sub2", Run: func(ctx context.Context, input any) (any, error) {
		subVisited = append(subVisited, "sub2")
		slog.Info("sub2")
		return input, nil
	}}

	subN1.Edges = []Edge{{To: "sub2"}}
	subDag := New(subN1)
	subDag.AddNode(subN2)

	n1 := &Node{ID: "n1", SubDag: subDag}

	n2 := &Node{ID: "n2", Run: func(ctx context.Context, input any) (any, error) {
		subVisited = append(subVisited, "n2")
		slog.Info("n2")
		return input, nil
	}}

	n1.Edges = []Edge{{To: "n2"}}
	d := New(n1)
	d.AddNode(n2)

	_, err := d.Execute(context.Background(), nil)
	assert.NoError(t, err)
	assert.Equal(t, []string{"sub1", "sub2", "n2"}, subVisited)
}

func TestDAGCinditional(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		n1 := &Node{ID: "n1", Run: func(ctx context.Context, input any) (any, error) {
			t.Log("n1", input)
			return input, nil
		}}

		n2 := &Node{ID: "n2", Run: func(ctx context.Context, input any) (any, error) {
			t.Log("n2", input)
			return input, nil
		}}

		n3 := &Node{ID: "n3", Run: func(ctx context.Context, input any) (any, error) {
			t.Log("n3", input)
			return input, nil
		}}

		n1.Edges = []Edge{{To: "n2", Condition: func(input any) bool {
			return input.(int) == 1
		}}, {To: "n3", Condition: func(input any) bool {
			return input.(int) == 2
		}}}

		d := New(n1)
		d.AddNode(n2)
		d.AddNode(n3)

		out, err := d.Execute(context.Background(), 1)
		assert.NoError(t, err)
		assert.Equal(t, 1, out)

		out, err = d.Execute(context.Background(), -1)
		assert.NoError(t, err)
		assert.Equal(t, -1, out)
	})
}

func TestDAGForeach(t *testing.T) {
	type User struct {
		Name string
		Age  int
	}

	users := []User{
		{
			Name: "John",
			Age:  20,
		},
		{
			Name: "Jane",
			Age:  20,
		},
	}
	t.Run("success", func(t *testing.T) {
		var count = 0
		eachN := &Node{ID: "each-n", Run: func(ctx context.Context, input any) (any, error) {
			var eachAny []any
			for _, user := range users {
				eachAny = append(eachAny, user)
			}
			return eachAny, nil
		}}

		n2 := &Node{ID: "worker", Run: func(ctx context.Context, input any) (any, error) {
			count++
			return input, nil
		}}

		eachN.Edges = []Edge{{To: "worker", Foreach: true}}

		d := New(eachN)
		d.AddNode(n2)

		_, err := d.Execute(context.Background(), nil)
		assert.NoError(t, err)
		assert.Equal(t, 2, count)
	})
}
