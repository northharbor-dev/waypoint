package cli

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

var (
	addTitle string
	addPhase int
	addRole  string
	addOwner string
	addDeps  string
)

var addCmd = &cobra.Command{
	Use:   "add [WI-ID]",
	Short: "Add a single work item",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireConfig(); err != nil {
			return err
		}
		s, err := getStore()
		if err != nil {
			return err
		}
		defer s.Close(context.Background())
		ctx := context.Background()

		var deps []string
		if addDeps != "" {
			for _, d := range strings.Split(addDeps, ",") {
				d = strings.TrimSpace(d)
				if d != "" {
					deps = append(deps, d)
				}
			}
		}
		if deps == nil {
			deps = []string{}
		}

		item := models.WorkItem{
			ID:           args[0],
			Title:        addTitle,
			Phase:        addPhase,
			Owner:        addOwner,
			Role:         addRole,
			Status:       models.StatusNotStarted,
			Dependencies: deps,
			Project:      cfg.Project,
			UpdatedAt:    time.Now(),
		}

		existing, err := s.ListWorkItems(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading items: %w", err)
		}

		all := append(existing, item)
		g := dag.Build(all)
		issues := g.Validate()
		for _, issue := range issues {
			if issue.Severity == "error" {
				return fmt.Errorf("validation error: %s", issue.Message)
			}
		}

		if err := s.UpsertWorkItem(ctx, item); err != nil {
			return fmt.Errorf("adding item: %w", err)
		}

		fmt.Printf("Added %s (%s)\n", item.ID, item.Title)
		return nil
	},
}

func init() {
	addCmd.Flags().StringVar(&addTitle, "title", "", "Work item title (required)")
	addCmd.Flags().IntVar(&addPhase, "phase", 0, "Phase number (required)")
	addCmd.Flags().StringVar(&addRole, "role", "", "Assigned role (required)")
	addCmd.Flags().StringVar(&addOwner, "owner", "agent", "Owner")
	addCmd.Flags().StringVar(&addDeps, "deps", "", "Comma-separated dependency IDs")
	_ = addCmd.MarkFlagRequired("title")
	_ = addCmd.MarkFlagRequired("phase")
	_ = addCmd.MarkFlagRequired("role")
	rootCmd.AddCommand(addCmd)
}
