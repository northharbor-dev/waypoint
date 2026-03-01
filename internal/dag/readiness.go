package dag

import (
	"sort"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func (g *Graph) ReadyItems(role string) []*models.WorkItem {
	var ready []*models.WorkItem

	for _, item := range g.Items {
		if item.Status != models.StatusNotStarted {
			continue
		}
		if !g.DepsMetFor(item.ID) {
			continue
		}
		if role != "" && item.Role != role {
			continue
		}
		ready = append(ready, item)
	}

	sort.Slice(ready, func(i, j int) bool {
		return ready[i].ID < ready[j].ID
	})

	return ready
}

func (g *Graph) DepsMetFor(id string) bool {
	for _, dep := range g.Dependencies[id] {
		depItem, exists := g.Items[dep]
		if !exists || depItem.Status != models.StatusDone {
			return false
		}
	}
	return true
}
