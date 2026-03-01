package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"text/tabwriter"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

type StatusJSON struct {
	Project      string           `json:"project"`
	GeneratedAt  time.Time        `json:"generated_at"`
	Summary      StatusSummary    `json:"summary"`
	Phases       []models.Phase   `json:"phases"`
	WorkItems    []StatusWorkItem `json:"work_items"`
	ReadyItems   []string         `json:"ready_items"`
	CriticalPath []string         `json:"critical_path"`
}

type StatusSummary struct {
	Total      int `json:"total"`
	NotStarted int `json:"not_started"`
	InProgress int `json:"in_progress"`
	Done       int `json:"done"`
	Blocked    int `json:"blocked"`
}

type StatusWorkItem struct {
	models.WorkItem
	Ready   bool   `json:"ready"`
	DepsMet bool   `json:"deps_met"`
	Elapsed string `json:"elapsed,omitempty"`
}

var (
	statusPhase  int
	statusOutput string
)

var statusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show all tasks with current state",
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

		if statusOutput == "json" {
			return printStatusJSON(items, phases, graph)
		}

		return printStatusTable(items, graph)
	},
}

func printStatusJSON(items []models.WorkItem, phases []models.Phase, graph *dag.Graph) error {
	now := time.Now()
	summary := StatusSummary{Total: len(items)}
	var workItems []StatusWorkItem

	for _, item := range items {
		switch item.Status {
		case models.StatusNotStarted:
			summary.NotStarted++
		case models.StatusInProgress:
			summary.InProgress++
		case models.StatusDone:
			summary.Done++
		case models.StatusBlocked:
			summary.Blocked++
		}

		swi := StatusWorkItem{
			WorkItem: item,
			Ready:    item.Status == models.StatusNotStarted && graph.DepsMetFor(item.ID),
			DepsMet:  graph.DepsMetFor(item.ID),
		}
		if item.Status == models.StatusInProgress && item.StartedAt != nil {
			swi.Elapsed = models.FormatDuration(now.Sub(*item.StartedAt))
		}
		workItems = append(workItems, swi)
	}

	sort.Slice(workItems, func(i, j int) bool {
		return workItems[i].ID < workItems[j].ID
	})

	readyItems := graph.ReadyItems("")
	readyIDs := make([]string, len(readyItems))
	for i, r := range readyItems {
		readyIDs[i] = r.ID
	}

	out := StatusJSON{
		Project:      cfg.Project,
		GeneratedAt:  now,
		Summary:      summary,
		Phases:       phases,
		WorkItems:    workItems,
		ReadyItems:   readyIDs,
		CriticalPath: graph.CriticalPath(),
	}

	data, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		return fmt.Errorf("marshalling JSON: %w", err)
	}
	fmt.Println(string(data))
	return nil
}

func printStatusTable(items []models.WorkItem, graph *dag.Graph) error {
	now := time.Now()

	filtered := items
	if statusPhase > 0 {
		filtered = nil
		for _, item := range items {
			if item.Phase == statusPhase {
				filtered = append(filtered, item)
			}
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		return filtered[i].ID < filtered[j].ID
	})

	w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
	fmt.Fprintln(w, "ID\tTITLE\tROLE\tSTATUS\tCLAIMED BY\tELAPSED")
	for _, item := range filtered {
		claimed := ""
		if item.ClaimedBy != nil {
			claimed = *item.ClaimedBy
		}

		elapsed := ""
		stale := ""
		if item.Status == models.StatusInProgress && item.StartedAt != nil {
			dur := now.Sub(*item.StartedAt)
			elapsed = models.FormatDuration(dur)
			if dur > cfg.StaleThreshold {
				stale = " [STALE]"
			}
		}

		fmt.Fprintf(w, "%s\t%s\t%s\t%s%s\t%s\t%s\n",
			item.ID, item.Title, item.Role, item.Status, stale, claimed, elapsed)
	}
	return w.Flush()
}

func init() {
	statusCmd.Flags().IntVar(&statusPhase, "phase", 0, "Filter by phase number")
	statusCmd.Flags().StringVar(&statusOutput, "output", "", "Output format (json)")
	rootCmd.AddCommand(statusCmd)
}
