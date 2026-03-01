package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

var removeForce bool

var removeCmd = &cobra.Command{
	Use:   "remove [WI-ID]",
	Short: "Remove a work item",
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

		targetID := args[0]

		items, err := s.ListWorkItems(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading items: %w", err)
		}

		var target *models.WorkItem
		var dependents []string
		for i, item := range items {
			if item.ID == targetID {
				target = &items[i]
			}
			for _, dep := range item.Dependencies {
				if dep == targetID {
					dependents = append(dependents, item.ID)
					break
				}
			}
		}

		if target == nil {
			return fmt.Errorf("work item %q not found", targetID)
		}

		if len(dependents) > 0 && !removeForce {
			fmt.Printf("Cannot remove %s: the following items depend on it:\n", targetID)
			for _, d := range dependents {
				fmt.Printf("  - %s\n", d)
			}
			return fmt.Errorf("use --force to remove and strip dependency references")
		}

		if removeForce && len(dependents) > 0 {
			now := time.Now()
			for _, depID := range dependents {
				depItem, getErr := s.GetWorkItem(ctx, cfg.Project, depID)
				if getErr != nil {
					return fmt.Errorf("loading dependent %s: %w", depID, getErr)
				}
				depItem.Dependencies = removeFromSlice(depItem.Dependencies, targetID)
				depItem.UpdatedAt = now
				if err := s.UpsertWorkItem(ctx, *depItem); err != nil {
					return fmt.Errorf("updating dependent %s: %w", depID, err)
				}
			}
		}

		if err := s.DeleteWorkItem(ctx, cfg.Project, targetID); err != nil {
			return fmt.Errorf("deleting item: %w", err)
		}

		fmt.Printf("Removed %s\n", targetID)
		return nil
	},
}

func init() {
	removeCmd.Flags().BoolVar(&removeForce, "force", false, "Force removal and strip dependency references from dependents")
	rootCmd.AddCommand(removeCmd)
}
