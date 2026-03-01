package dag

import (
	"fmt"

	"github.com/northharbor-dev/waypoint/internal/models"
)

type Graph struct {
	Items        map[string]*models.WorkItem
	Adjacency    map[string][]string
	Dependencies map[string][]string
}

func Build(items []models.WorkItem) *Graph {
	g := &Graph{
		Items:        make(map[string]*models.WorkItem, len(items)),
		Adjacency:    make(map[string][]string, len(items)),
		Dependencies: make(map[string][]string, len(items)),
	}

	for i := range items {
		item := &items[i]
		g.Items[item.ID] = item
		if g.Adjacency[item.ID] == nil {
			g.Adjacency[item.ID] = []string{}
		}
	}

	for _, item := range g.Items {
		for _, dep := range item.Dependencies {
			g.Dependencies[item.ID] = append(g.Dependencies[item.ID], dep)
			g.Adjacency[dep] = append(g.Adjacency[dep], item.ID)
		}
	}

	return g
}

func (g *Graph) TopologicalSort() ([]string, error) {
	inDegree := make(map[string]int, len(g.Items))
	for id := range g.Items {
		inDegree[id] = 0
	}
	for id := range g.Items {
		for _, dep := range g.Dependencies[id] {
			_ = dep
			inDegree[id]++
		}
	}

	var queue []string
	for id, deg := range inDegree {
		if deg == 0 {
			queue = append(queue, id)
		}
	}

	var order []string
	for len(queue) > 0 {
		node := queue[0]
		queue = queue[1:]
		order = append(order, node)

		for _, neighbor := range g.Adjacency[node] {
			inDegree[neighbor]--
			if inDegree[neighbor] == 0 {
				queue = append(queue, neighbor)
			}
		}
	}

	if len(order) != len(g.Items) {
		return nil, fmt.Errorf("cycle detected: processed %d of %d items", len(order), len(g.Items))
	}

	return order, nil
}
