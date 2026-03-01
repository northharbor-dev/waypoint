package cli

import (
	"context"
	"fmt"
	"os"
	"text/tabwriter"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/spf13/cobra"
)

var nextRole string

var nextCmd = &cobra.Command{
	Use:   "next",
	Short: "Show tasks ready to start",
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

		graph := dag.Build(items)
		ready := graph.ReadyItems(nextRole)

		if len(ready) == 0 {
			fmt.Println("No tasks are ready to start.")
			return nil
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tROLE\tPHASE")
		for _, item := range ready {
			fmt.Fprintf(w, "%s\t%s\t%s\t%d\n", item.ID, item.Title, item.Role, item.Phase)
		}
		return w.Flush()
	},
}

func init() {
	nextCmd.Flags().StringVar(&nextRole, "role", "", "Filter by role")
	rootCmd.AddCommand(nextCmd)
}
