package cli

import (
	"context"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/spf13/cobra"
)

var criticalPathCmd = &cobra.Command{
	Use:   "critical-path",
	Short: "Show the longest dependency chain",
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
		path := graph.CriticalPath()

		if len(path) == 0 {
			fmt.Println("No critical path found.")
			return nil
		}

		fmt.Println(strings.Join(path, " → "))
		fmt.Printf("\nLength: %d items\n\n", len(path))

		w := tabwriter.NewWriter(os.Stdout, 0, 4, 2, ' ', 0)
		fmt.Fprintln(w, "ID\tTITLE\tSTATUS")
		for _, id := range path {
			item := graph.Items[id]
			fmt.Fprintf(w, "%s\t%s\t%s\n", item.ID, item.Title, item.Status)
		}
		return w.Flush()
	},
}

func init() { rootCmd.AddCommand(criticalPathCmd) }
