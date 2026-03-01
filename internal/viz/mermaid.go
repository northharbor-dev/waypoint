package viz

import (
	"fmt"
	"strings"

	"github.com/northharbor-dev/waypoint/internal/models"
)

func GenerateMermaid(items []models.WorkItem, phases []models.Phase) string {
	phaseMap := make(map[int]string)
	for _, p := range phases {
		phaseMap[p.Number] = p.Name
	}

	grouped := make(map[int][]models.WorkItem)
	var ungrouped []models.WorkItem
	for _, item := range items {
		if _, ok := phaseMap[item.Phase]; ok {
			grouped[item.Phase] = append(grouped[item.Phase], item)
		} else {
			ungrouped = append(ungrouped, item)
		}
	}

	var b strings.Builder
	b.WriteString("graph TD\n")

	phaseNums := sortedKeys(grouped)
	for _, num := range phaseNums {
		name := phaseMap[num]
		b.WriteString(fmt.Sprintf("  subgraph phase%d[%s]\n", num, name))
		for _, item := range grouped[num] {
			b.WriteString(fmt.Sprintf("    %s\n", nodeDefinition(item)))
		}
		b.WriteString("  end\n")
	}

	for _, item := range ungrouped {
		b.WriteString(fmt.Sprintf("  %s\n", nodeDefinition(item)))
	}

	for _, item := range items {
		for _, dep := range item.Dependencies {
			b.WriteString(fmt.Sprintf("  %s --> %s\n", dep, item.ID))
		}
	}

	return b.String()
}

func nodeDefinition(item models.WorkItem) string {
	label := fmt.Sprintf("%s: %s [%s]", item.ID, item.Title, item.Status)
	switch item.Owner {
	case "human":
		return fmt.Sprintf("%s([%s])", item.ID, label)
	case "collaborative":
		return fmt.Sprintf("%s{{%s}}", item.ID, label)
	default:
		return fmt.Sprintf("%s[%s]", item.ID, label)
	}
}

func sortedKeys(m map[int][]models.WorkItem) []int {
	keys := make([]int, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := range keys {
		for j := i + 1; j < len(keys); j++ {
			if keys[j] < keys[i] {
				keys[i], keys[j] = keys[j], keys[i]
			}
		}
	}
	return keys
}
