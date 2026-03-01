package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

var unblockCmd = &cobra.Command{
	Use:   "unblock [WI-ID]",
	Short: "Clear blocker and reset to not_started",
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

		if err := s.UpdateStatus(ctx, cfg.Project, id, models.StatusNotStarted, ""); err != nil {
			return err
		}

		fmt.Printf("Unblocked %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(unblockCmd)
}
