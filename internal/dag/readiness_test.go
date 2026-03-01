package dag

import (
	"testing"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func TestReadyItemsAllRoles(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusDone),
		makeItem("B", "api_dev", models.StatusNotStarted, "A"),
		makeItem("C", "ui_dev", models.StatusNotStarted, "A"),
		makeItem("D", "backend_dev", models.StatusNotStarted, "B"),
	}
	g := Build(items)

	ready := g.ReadyItems("")
	if len(ready) != 2 {
		t.Fatalf("ReadyItems(\"\") returned %d items, want 2", len(ready))
	}
	ids := []string{ready[0].ID, ready[1].ID}
	if ids[0] != "B" || ids[1] != "C" {
		t.Errorf("ReadyItems(\"\") = %v, want [B, C]", ids)
	}
}

func TestReadyItemsFilterByRole(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusDone),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
		makeItem("C", "ui_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	ready := g.ReadyItems("backend_dev")
	if len(ready) != 1 {
		t.Fatalf("ReadyItems(\"backend_dev\") returned %d items, want 1", len(ready))
	}
	if ready[0].ID != "B" {
		t.Errorf("ReadyItems(\"backend_dev\")[0].ID = %q, want \"B\"", ready[0].ID)
	}
}

func TestReadyItemsNoneReady(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusInProgress),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	ready := g.ReadyItems("")
	if len(ready) != 0 {
		t.Errorf("ReadyItems(\"\") returned %d items, want 0", len(ready))
	}
}

func TestReadyItemsSkipsNonNotStarted(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusDone),
		makeItem("B", "backend_dev", models.StatusInProgress),
		makeItem("C", "backend_dev", models.StatusBlocked),
	}
	g := Build(items)

	ready := g.ReadyItems("")
	if len(ready) != 0 {
		t.Errorf("ReadyItems(\"\") returned %d items, want 0 (all non-not_started)", len(ready))
	}
}

func TestDepsMetForAllDone(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusDone),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	if !g.DepsMetFor("B") {
		t.Error("DepsMetFor(\"B\") = false, want true (dep A is done)")
	}
}

func TestDepsMetForNotDone(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusInProgress),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	if g.DepsMetFor("B") {
		t.Error("DepsMetFor(\"B\") = true, want false (dep A is in_progress)")
	}
}

func TestDepsMetForNoDeps(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
	}
	g := Build(items)

	if !g.DepsMetFor("A") {
		t.Error("DepsMetFor(\"A\") = false, want true (no dependencies)")
	}
}
