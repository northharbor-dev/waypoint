package cli

import (
	"context"
	"fmt"
	"time"

	"github.com/spf13/cobra"
)

var resetCmd = &cobra.Command{
	Use:   "reset [WI-ID]",
	Short: "Release an orphaned task (crash recovery)",
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

		if err := s.ReleaseWorkItem(ctx, cfg.Project, id); err != nil {
			return err
		}

		fmt.Printf("Released %s\n", id)
		return nil
	},
}

func init() {
	rootCmd.AddCommand(resetCmd)
}
