package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/northharbor-dev/waypoint/internal/dag"
	"github.com/northharbor-dev/waypoint/internal/models"
	"github.com/spf13/cobra"
	"gopkg.in/yaml.v3"
)

var initCmd = &cobra.Command{
	Use:   "init [seed-file]",
	Short: "Seed a new project from a YAML file",
	Args:  cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		if err := requireConfig(); err != nil {
			return err
		}

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading seed file: %w", err)
		}

		var seed models.SeedFile
		if err := yaml.Unmarshal(data, &seed); err != nil {
			return fmt.Errorf("parsing seed file: %w", err)
		}

		now := time.Now()
		items := make([]models.WorkItem, len(seed.WorkItems))
		for i, sw := range seed.WorkItems {
			deps := sw.Dependencies
			if deps == nil {
				deps = []string{}
			}
			items[i] = models.WorkItem{
				ID:           sw.ID,
				Title:        sw.Title,
				Phase:        sw.Phase,
				Owner:        sw.Owner,
				Role:         sw.Role,
				Status:       models.StatusNotStarted,
				Dependencies: deps,
				Project:      cfg.Project,
				UpdatedAt:    now,
			}
		}

		phases := make([]models.Phase, len(seed.Phases))
		for i, sp := range seed.Phases {
			phases[i] = models.Phase{
				ID:      fmt.Sprintf("%s:%d", cfg.Project, sp.Number),
				Number:  sp.Number,
				Name:    sp.Name,
				Target:  sp.Target,
				Project: cfg.Project,
			}
		}

		g := dag.Build(items)
		issues := g.Validate()
		if len(issues) > 0 {
			hasError := false
			for _, issue := range issues {
				prefix := "⚠"
				if issue.Severity == "error" {
					prefix = "✗"
					hasError = true
				}
				fmt.Printf("  %s %s\n", prefix, issue.Message)
			}
			if hasError {
				return fmt.Errorf("seed file has validation errors; aborting")
			}
		}

		s, err := getStore()
		if err != nil {
			return err
		}
		defer s.Close(context.Background())

		if err := s.SeedProject(context.Background(), cfg.Project, items, phases); err != nil {
			return fmt.Errorf("seeding project: %w", err)
		}

		fmt.Printf("Seeded project %q: %d work items, %d phases\n", cfg.Project, len(items), len(phases))
		return nil
	},
}

func init() { rootCmd.AddCommand(initCmd) }
