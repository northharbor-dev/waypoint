package cli

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/northharbor-dev/waypoint/internal/store/mongo"
	"github.com/spf13/cobra"
)

var doneCmd = &cobra.Command{
	Use:   "done [WI-ID]",
	Short: "Mark a task as complete",
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

		if err := s.CompleteWorkItem(ctx, cfg.Project, id); err != nil {
			if errors.Is(err, mongo.ErrNotFound) {
				return fmt.Errorf("work item %s not found", id)
			}
			return err
		}

		fmt.Printf("Completed %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(doneCmd)
}
