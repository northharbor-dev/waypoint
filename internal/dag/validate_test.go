package dag

import (
	"strings"
	"testing"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func TestValidateCleanDAG(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusDone),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	issues := g.Validate()
	if len(issues) != 0 {
		t.Errorf("Validate() returned %d issues, want 0: %+v", len(issues), issues)
	}
}

func TestValidateCycle(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted, "B"),
		makeItem("B", "backend_dev", models.StatusNotStarted, "A"),
	}
	g := Build(items)

	issues := g.Validate()
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Message, "cycle") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() should report cycle error, got %+v", issues)
	}
}

func TestValidateDanglingDependency(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted, "nonexistent"),
	}
	g := Build(items)

	issues := g.Validate()
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Message, "non-existent") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() should report dangling dependency, got %+v", issues)
	}
}

func TestValidateInvalidRole(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "invalid_role", models.StatusNotStarted),
	}
	g := Build(items)

	issues := g.Validate()
	found := false
	for _, issue := range issues {
		if issue.Severity == "error" && strings.Contains(issue.Message, "invalid role") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() should report invalid role, got %+v", issues)
	}
}

func TestValidateStatusInconsistency(t *testing.T) {
	items := []models.WorkItem{
		makeItem("A", "backend_dev", models.StatusNotStarted),
		makeItem("B", "backend_dev", models.StatusInProgress, "A"),
	}
	g := Build(items)

	issues := g.Validate()
	found := false
	for _, issue := range issues {
		if issue.Severity == "warning" && strings.Contains(issue.Message, "not_started") {
			found = true
			break
		}
	}
	if !found {
		t.Errorf("Validate() should report status inconsistency warning, got %+v", issues)
	}
}
