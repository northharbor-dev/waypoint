package dag

import (
	"fmt"

	"github.com/northharbor-dev/waypoint/internal/models"
)

type ValidationIssue struct {
	Severity string
	Message  string
}

func (g *Graph) Validate() []ValidationIssue {
	var issues []ValidationIssue

	if _, err := g.TopologicalSort(); err != nil {
		issues = append(issues, ValidationIssue{
			Severity: "error",
			Message:  fmt.Sprintf("cycle detected in dependency graph: %v", err),
		})
	}

	for id, deps := range g.Dependencies {
		for _, dep := range deps {
			if _, exists := g.Items[dep]; !exists {
				issues = append(issues, ValidationIssue{
					Severity: "error",
					Message:  fmt.Sprintf("item %q depends on non-existent item %q", id, dep),
				})
			}
		}
	}

	for _, item := range g.Items {
		if item.Role != "" && !models.IsValidRole(item.Role) {
			issues = append(issues, ValidationIssue{
				Severity: "error",
				Message:  fmt.Sprintf("item %q has invalid role %q", item.ID, item.Role),
			})
		}
	}

	for _, item := range g.Items {
		if item.Status == models.StatusInProgress || item.Status == models.StatusDone {
			for _, dep := range item.Dependencies {
				depItem, exists := g.Items[dep]
				if exists && depItem.Status == models.StatusNotStarted {
					issues = append(issues, ValidationIssue{
						Severity: "warning",
						Message:  fmt.Sprintf("item %q is %s but dependency %q is not_started", item.ID, item.Status, dep),
					})
				}
			}
		}
	}

	return issues
}
