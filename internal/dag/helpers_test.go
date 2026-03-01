package dag

import "github.com/northharbor-dev/waypoint/internal/models"

func makeItem(id, role string, status models.Status, deps ...string) models.WorkItem {
	if deps == nil {
		deps = []string{}
	}
	return models.WorkItem{
		ID:           id,
		Title:        id + " title",
		Phase:        1,
		Owner:        "agent",
		Role:         role,
		Status:       status,
		Dependencies: deps,
	}
}

func contains(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}
