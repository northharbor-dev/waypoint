package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
)

var blockReason string

var blockCmd = &cobra.Command{
	Use:   "block [WI-ID]",
	Short: "Mark a task as blocked",
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

		if err := s.UpdateStatus(ctx, cfg.Project, id, models.StatusBlocked, blockReason); err != nil {
			return err
		}

		fmt.Printf("Blocked %s: %s\n", id, blockReason)
		return nil
	},
}

func init() {
	blockCmd.Flags().StringVarP(&blockReason, "reason", "r", "", "Reason for blocking (required)")
	blockCmd.MarkFlagRequired("reason")
	rootCmd.AddCommand(blockCmd)
}
