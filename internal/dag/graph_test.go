package dag

import (
	"testing"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func TestBuild(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "backend_dev", models.StatusNotStarted, "A", "B"),
	}

	g := Build(items)

	if len(g.Items) != 3 {
		t.Fatalf("Items map has %d entries, want 3", len(g.Items))
	}
	for _, id := range []string{"A", "B", "C"} {
		if _, exists := g.Items[id]; !exists {
			t.Errorf("Items map missing %q", id)
		}
	}

	if adj := g.Adjacency["A"]; !contains(adj, "B") || !contains(adj, "C") {
		t.Errorf("Adjacency[A] = %v, want to contain B and C", adj)
	}
	if adj := g.Adjacency["B"]; !contains(adj, "C") {
		t.Errorf("Adjacency[B] = %v, want to contain C", adj)
	}

	if deps := g.Dependencies["B"]; len(deps) != 1 || deps[0] != "A" {
		t.Errorf("Dependencies[B] = %v, want [A]", deps)
	}
	if deps := g.Dependencies["C"]; len(deps) != 2 {
		t.Errorf("Dependencies[C] has %d entries, want 2", len(deps))
	}
}

func TestTopologicalSortValid(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "backend_dev", models.StatusNotStarted, "A", "B"),
	}
	g := Build(items)

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() returned error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("TopologicalSort() returned %d items, want 3", len(order))
	}

	pos := make(map[string]int)
	for i, id := range order {
		pos[id] = i
	}
	for _, item := range items {
		for _, dep := range item.Dependencies {
			if pos[dep] >= pos[item.ID] {
				t.Errorf("dependency %q (pos %d) should come before %q (pos %d)",
					dep, pos[dep], item.ID, pos[item.ID])
			}
		}
	}
}

func TestTopologicalSortCycle(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted, "B"),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	_, err := g.TopologicalSort()
	if err == nil {
		t.Fatal("TopologicalSort() should return error for cyclic graph")
	}
}

func TestTopologicalSortNoDeps(t *testing.T) {
	items := []models.WorkItem{
		makeItem("X", "backend_dev", models.StatusNotStarted),
		makeItem("Y", "backend_dev", models.StatusNotStarted),
		makeItem("Z", "backend_dev", models.StatusNotStarted),
	}
	g := Build(items)

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() returned error: %v", err)
	}
	if len(order) != 3 {
		t.Fatalf("got %d items, want 3", len(order))
	}

	seen := make(map[string]bool)
	for _, id := range order {
		seen[id] = true
	}
	for _, id := range []string{"X", "Y", "Z"} {
		if !seen[id] {
			t.Errorf("missing item %q in order", id)
		}
	}
}

func TestTopologicalSortLinearChain(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "backend_dev", models.StatusNotStarted, "B"),
	}
	g := Build(items)

	order, err := g.TopologicalSort()
	if err != nil {
		t.Fatalf("TopologicalSort() returned error: %v", err)
	}

	want := []string{"A", "B", "C"}
	if len(order) != len(want) {
		t.Fatalf("got %v, want %v", order, want)
	}
	for i := range want {
		if order[i] != want[i] {
			t.Errorf("order[%d] = %q, want %q", i, order[i], want[i])
		}
	}
}
