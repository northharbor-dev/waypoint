package viz

import (
	"strings"
	"testing"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func TestGenerateMermaid(t *testing.T) {
	items := []models.WorkItem{
		{
			ID:     "A",
			Title:  "Setup",
			Phase:  1,
			Owner:  "human",
			Role:   "lead",
			Status: models.StatusDone,
		},
		{
			ID:           "B",
			Title:        "API",
			Phase:        2,
			Owner:        "agent",
			Role:         "api_dev",
			Status:       models.StatusNotStarted,
			Dependencies: []string{"A"},
		},
		{
			ID:           "C",
			Title:        "Collab Task",
			Phase:        2,
			Owner:        "collaborative",
			Role:         "ui_dev",
			Status:       models.StatusInProgress,
			Dependencies: []string{"A"},
		},
	}
	phases := []models.Phase{
		{Number: 1, Name: "Foundation"},
		{Number: 2, Name: "Development"},
	}

	output := GenerateMermaid(items, phases)

	checks := []struct {
		name     string
		contains string
	}{
		{"graph directive", "graph TD"},
		{"phase 1 label", "Foundation"},
		{"phase 2 label", "Development"},
		{"human node stadium shape", "A(["},
		{"agent node rectangle shape", "B["},
		{"collaborative node hexagon shape", "C{{"},
		{"dep edge A to B", "A --> B"},
		{"dep edge A to C", "A --> C"},
	}
	for _, c := range checks {
		if !strings.Contains(output, c.contains) {
			t.Errorf("%s: output should contain %q\nGot:\n%s", c.name, c.contains, output)
		}
	}

	if strings.Contains(output, "B([") {
		t.Errorf("agent-owned node B should not use stadium shape\nGot:\n%s", output)
	}
}

func TestGenerateMermaidUngroupedItems(t *testing.T) {
	items := []models.WorkItem{
		{
			ID:     "X",
			Title:  "Orphan",
			Phase:  99,
			Owner:  "agent",
			Role:   "backend_dev",
			Status: models.StatusNotStarted,
		},
	}

	output := GenerateMermaid(items, nil)

	if !strings.Contains(output, "graph TD") {
		t.Error("output should contain 'graph TD'")
	}
	if !strings.Contains(output, "X[") {
		t.Errorf("ungrouped item X should appear as rectangle node\nGot:\n%s", output)
	}
	if strings.Contains(output, "subgraph") {
		t.Errorf("should have no subgraphs with no matching phases\nGot:\n%s", output)
	}
}
