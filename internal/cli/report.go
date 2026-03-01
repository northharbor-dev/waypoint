package cli

import (
	"context"
	"fmt"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

var reportOutput string

var reportCmd = &cobra.Command{
	Use:   "report",
	Short: "Generate a markdown status report",
	RunE: func(cmd *cobra.Command, args []string) error {
		s, err := getStore()
		if err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		defer s.Close(ctx)

		items, err := s.ListWorkItems(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading work items: %w", err)
		}

		phases, err := s.ListPhases(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading phases: %w", err)
		}

		graph := dag.Build(items)
		readyItems := graph.ReadyItems("")
		criticalPath := graph.CriticalPath()
		now := time.Now()

		var b strings.Builder

		fmt.Fprintf(&b, "# %s -- Project Status Report\n\n", cfg.Project)
		fmt.Fprintf(&b, "_Generated: %s_\n\n", now.Format(time.RFC3339))

		var total, done, inProgress, blocked, notStarted int
		for _, item := range items {
			total++
			switch item.Status {
			case models.StatusDone:
				done++
			case models.StatusInProgress:
				inProgress++
			case models.StatusBlocked:
				blocked++
			case models.StatusNotStarted:
				notStarted++
			}
		}

		progress := 0.0
		if total > 0 {
			progress = float64(done) / float64(total) * 100
		}

		fmt.Fprintf(&b, "## Summary\n\n")
		fmt.Fprintf(&b, "| Metric | Count |\n")
		fmt.Fprintf(&b, "|--------|-------|\n")
		fmt.Fprintf(&b, "| Total | %d |\n", total)
		fmt.Fprintf(&b, "| Done | %d |\n", done)
		fmt.Fprintf(&b, "| In Progress | %d |\n", inProgress)
		fmt.Fprintf(&b, "| Blocked | %d |\n", blocked)
		fmt.Fprintf(&b, "| Not Started | %d |\n", notStarted)
		fmt.Fprintf(&b, "| Progress | %.1f%% |\n\n", progress)

		if len(phases) > 0 {
			fmt.Fprintf(&b, "## Phase Progress\n\n")
			fmt.Fprintf(&b, "| Phase | Name | Done | Total | Progress |\n")
			fmt.Fprintf(&b, "|-------|------|------|-------|----------|\n")

			sort.Slice(phases, func(i, j int) bool {
				return phases[i].Number < phases[j].Number
			})

			for _, phase := range phases {
				var pTotal, pDone int
				for _, item := range items {
					if item.Phase == phase.Number {
						pTotal++
						if item.Status == models.StatusDone {
							pDone++
						}
					}
				}
				pPct := 0.0
				if pTotal > 0 {
					pPct = float64(pDone) / float64(pTotal) * 100
				}
				fmt.Fprintf(&b, "| %d | %s | %d | %d | %.0f%% |\n",
					phase.Number, phase.Name, pDone, pTotal, pPct)
			}
			b.WriteString("\n")
		}

		completed := filterByStatus(items, models.StatusDone)
		sort.Slice(completed, func(i, j int) bool {
			if completed[i].CompletedAt == nil {
				return false
			}
			if completed[j].CompletedAt == nil {
				return true
			}
			return completed[i].CompletedAt.After(*completed[j].CompletedAt)
		})

		if len(completed) > 0 {
			fmt.Fprintf(&b, "## Recently Completed\n\n")
			fmt.Fprintf(&b, "| ID | Title | Started | Completed | Duration |\n")
			fmt.Fprintf(&b, "|----|-------|---------|-----------|----------|\n")
			for _, item := range completed {
				started := ""
				if item.StartedAt != nil {
					started = item.StartedAt.Format("2006-01-02 15:04")
				}
				completedAt := ""
				duration := ""
				if item.CompletedAt != nil {
					completedAt = item.CompletedAt.Format("2006-01-02 15:04")
					if item.StartedAt != nil {
						duration = models.FormatDuration(item.CompletedAt.Sub(*item.StartedAt))
					}
				}
				fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
					item.ID, item.Title, started, completedAt, duration)
			}
			b.WriteString("\n")
		}

		inProgressItems := filterByStatus(items, models.StatusInProgress)
		sort.Slice(inProgressItems, func(i, j int) bool {
			return inProgressItems[i].ID < inProgressItems[j].ID
		})

		if len(inProgressItems) > 0 {
			fmt.Fprintf(&b, "## In Progress\n\n")
			fmt.Fprintf(&b, "| ID | Title | Claimed By | Started | Elapsed |\n")
			fmt.Fprintf(&b, "|----|-------|------------|---------|--------|\n")
			for _, item := range inProgressItems {
				claimed := ""
				if item.ClaimedBy != nil {
					claimed = *item.ClaimedBy
				}
				started := ""
				elapsed := ""
				if item.StartedAt != nil {
					started = item.StartedAt.Format("2006-01-02 15:04")
					elapsed = models.FormatDuration(now.Sub(*item.StartedAt))
				}
				fmt.Fprintf(&b, "| %s | %s | %s | %s | %s |\n",
					item.ID, item.Title, claimed, started, elapsed)
			}
			b.WriteString("\n")
		}

		blockedItems := filterByStatus(items, models.StatusBlocked)
		sort.Slice(blockedItems, func(i, j int) bool {
			return blockedItems[i].ID < blockedItems[j].ID
		})

		if len(blockedItems) > 0 {
			fmt.Fprintf(&b, "## Blocked\n\n")
			fmt.Fprintf(&b, "| ID | Title | Reason |\n")
			fmt.Fprintf(&b, "|----|-------|--------|\n")
			for _, item := range blockedItems {
				reason := ""
				if item.BlockerNote != nil {
					reason = *item.BlockerNote
				}
				fmt.Fprintf(&b, "| %s | %s | %s |\n", item.ID, item.Title, reason)
			}
			b.WriteString("\n")
		}

		if len(readyItems) > 0 {
			fmt.Fprintf(&b, "## Ready to Start\n\n")
			fmt.Fprintf(&b, "| ID | Title | Role | Phase |\n")
			fmt.Fprintf(&b, "|----|-------|------|-------|\n")
			for _, item := range readyItems {
				fmt.Fprintf(&b, "| %s | %s | %s | %d |\n",
					item.ID, item.Title, item.Role, item.Phase)
			}
			b.WriteString("\n")
		}

		var staleItems []models.WorkItem
		for _, item := range inProgressItems {
			if item.StartedAt != nil && now.Sub(*item.StartedAt) > cfg.StaleThreshold {
				staleItems = append(staleItems, item)
			}
		}

		if len(staleItems) > 0 {
			fmt.Fprintf(&b, "## Stale Items\n\n")
			fmt.Fprintf(&b, "| ID | Title | Claimed By | Elapsed |\n")
			fmt.Fprintf(&b, "|----|-------|------------|--------|\n")
			for _, item := range staleItems {
				claimed := ""
				if item.ClaimedBy != nil {
					claimed = *item.ClaimedBy
				}
				elapsed := models.FormatDuration(now.Sub(*item.StartedAt))
				fmt.Fprintf(&b, "| %s | %s | %s | %s |\n",
					item.ID, item.Title, claimed, elapsed)
			}
			b.WriteString("\n")
		}

		if len(criticalPath) > 0 {
			fmt.Fprintf(&b, "## Critical Path\n\n")
			fmt.Fprintf(&b, "%s\n\n", strings.Join(criticalPath, " → "))
		}

		report := b.String()

		if reportOutput != "" {
			if err := os.WriteFile(reportOutput, []byte(report), 0644); err != nil {
				return fmt.Errorf("writing report: %w", err)
			}
			fmt.Printf("Report written to %s\n", reportOutput)
			return nil
		}

		fmt.Print(report)
		return nil
	},
}

func filterByStatus(items []models.WorkItem, status models.Status) []models.WorkItem {
	var result []models.WorkItem
	for _, item := range items {
		if item.Status == status {
			result = append(result, item)
		}
	}
	return result
}

func init() {
	reportCmd.Flags().StringVar(&reportOutput, "output", "", "Write report to file")
	rootCmd.AddCommand(reportCmd)
}
