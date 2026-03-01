package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/northharbor-dev/waypoint/internal/store/mongo"
	"github.com/spf13/cobra"
)

var startCmd = &cobra.Command{
	Use:   "start [WI-ID]",
	Short: "Claim a task to work on",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		id := args[0]

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

		graph := dag.Build(items)

		if !graph.DepsMetFor(id) {
			var pending []string
			for _, dep := range graph.Dependencies[id] {
				if item, ok := graph.Items[dep]; ok && item.Status != models.StatusDone {
					pending = append(pending, dep)
				}
			}
			return fmt.Errorf("cannot start %s: dependencies not done: %v", id, pending)
		}

		item, err := s.ClaimWorkItem(ctx, cfg.Project, id, "agent")
		if err != nil {
			if errors.Is(err, mongo.ErrAlreadyClaimed) {
				fmt.Printf("%s already claimed. Run 'waypoint next' for available tasks.\n", id)
				return nil
			}
			return err
		}

		fmt.Printf("Started %s: %s\n", item.ID, item.Title)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(startCmd)
}
