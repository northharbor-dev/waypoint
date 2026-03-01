package cli

import (
	"context"
	"fmt"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/spf13/cobra"
)

var validateCmd = &cobra.Command{
	Use:   "validate",
	Short: "Check DAG integrity",
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireConfig(); err != nil {
			return err
		}
		s, err := getStore()
		if err != nil {
			return err
		}
		defer s.Close(context.Background())

		items, err := s.ListWorkItems(context.Background(), cfg.Project)
		if err != nil {
			return fmt.Errorf("loading items: %w", err)
		}

		g := dag.Build(items)
		issues := g.Validate()

		if len(issues) == 0 {
			fmt.Println("✓ DAG is valid")
			return nil
		}

		hasError := false
		for _, issue := range issues {
			if issue.Severity == "error" {
				fmt.Printf("✗ %s\n", issue.Message)
				hasError = true
			} else {
				fmt.Printf("⚠ %s\n", issue.Message)
			}
		}

		if hasError {
			return fmt.Errorf("DAG has validation errors")
		}
		return nil
	},
}

func init() { rootCmd.AddCommand(validateCmd) }
