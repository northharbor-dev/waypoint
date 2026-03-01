package dag

import (
	"testing"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func TestCriticalPathLinearChain(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "backend_dev", models.StatusNotStarted, "B"),
	}
	g := Build(items)

	path := g.CriticalPath()
	want := []string{"A", "B", "C"}
	if len(path) != len(want) {
		t.Fatalf("CriticalPath() = %v, want %v", path, want)
	}
	for i := range want {
		if path[i] != want[i] {
			t.Errorf("path[%d] = %q, want %q", i, path[i], want[i])
		}
	}
}

func TestCriticalPathDiamondReturnsLongest(t *testing.T) {
	// Long arm: A→B→E→D (4 nodes), short arm: A→C→D (3 nodes)
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("E", "backend_dev", models.StatusNotStarted, "B"),
		makeItem("D", "backend_dev", models.StatusNotStarted, "E", "C"),
	}
	g := Build(items)

	path := g.CriticalPath()
	want := []string{"A", "B", "E", "D"}
	if len(path) != len(want) {
		t.Fatalf("CriticalPath() length = %d, want %d: %v", len(path), len(want), path)
	}
	for i := range want {
		if path[i] != want[i] {
			t.Errorf("path[%d] = %q, want %q", i, path[i], want[i])
		}
	}
}

func TestCriticalPathSingleItem(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
	}
	g := Build(items)

	path := g.CriticalPath()
	if len(path) != 1 || path[0] != "A" {
		t.Errorf("CriticalPath() = %v, want [A]", path)
	}
}

func TestCriticalPathEmptyGraph(t *testing.T) {
	g := Build([]models.WorkItem{})

	path := g.CriticalPath()
	if len(path) != 0 {
		t.Errorf("CriticalPath() = %v, want empty", path)
	}
}
