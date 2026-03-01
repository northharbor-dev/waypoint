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

var syncForce bool

var syncCmd = &cobra.Command{
	Use:   "sync [seed-file]",
	Short: "Reconcile an updated YAML file with live state",
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

		data, err := os.ReadFile(args[0])
		if err != nil {
			return fmt.Errorf("reading seed file: %w", err)
		}

		var seed models.SeedFile
		if err := yaml.Unmarshal(data, &seed); err != nil {
			return fmt.Errorf("parsing seed file: %w", err)
		}

		existing, err := s.ListWorkItems(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading existing items: %w", err)
		}

		existingMap := make(map[string]models.WorkItem, len(existing))
		for _, item := range existing {
			existingMap[item.ID] = item
		}

		seedMap := make(map[string]models.SeedWorkItem, len(seed.WorkItems))
		for _, sw := range seed.WorkItems {
			seedMap[sw.ID] = sw
		}

		now := time.Now()
		var added, modified []models.WorkItem
		var flagged []string

		for _, sw := range seed.WorkItems {
			deps := sw.Dependencies
			if deps == nil {
				deps = []string{}
			}
			if ex, ok := existingMap[sw.ID]; ok {
				changed := ex.Title != sw.Title || ex.Role != sw.Role || ex.Phase != sw.Phase || ex.Owner != sw.Owner || !sliceEqual(ex.Dependencies, deps)
				if changed {
					ex.Title = sw.Title
					ex.Role = sw.Role
					ex.Phase = sw.Phase
					ex.Owner = sw.Owner
					ex.Dependencies = deps
					ex.UpdatedAt = now
					modified = append(modified, ex)
				}
			} else {
				added = append(added, models.WorkItem{
					ID:           sw.ID,
					Title:        sw.Title,
					Phase:        sw.Phase,
					Owner:        sw.Owner,
					Role:         sw.Role,
					Status:       models.StatusNotStarted,
					Dependencies: deps,
					Project:      cfg.Project,
					UpdatedAt:    now,
				})
			}
		}

		var removed []models.WorkItem
		for _, ex := range existing {
			if _, ok := seedMap[ex.ID]; !ok {
				removed = append(removed, ex)
			}
		}

		merged := make([]models.WorkItem, 0, len(existing)+len(added))
		for _, ex := range existing {
			if _, ok := seedMap[ex.ID]; ok {
				merged = append(merged, ex)
			}
		}
		for _, m := range modified {
			for i, item := range merged {
				if item.ID == m.ID {
					merged[i] = m
					break
				}
			}
		}
		merged = append(merged, added...)

		g := dag.Build(merged)
		issues := g.Validate()
		for _, issue := range issues {
			if issue.Severity == "error" {
				return fmt.Errorf("validation error after merge: %s", issue.Message)
			}
		}

		for _, item := range added {
			if err := s.UpsertWorkItem(ctx, item); err != nil {
				return fmt.Errorf("adding %s: %w", item.ID, err)
			}
		}
		for _, item := range modified {
			if err := s.UpsertWorkItem(ctx, item); err != nil {
				return fmt.Errorf("updating %s: %w", item.ID, err)
			}
		}

		dependentMap := make(map[string][]string)
		for _, item := range existing {
			for _, dep := range item.Dependencies {
				dependentMap[dep] = append(dependentMap[dep], item.ID)
			}
		}

		for _, item := range removed {
			hasDependents := len(dependentMap[item.ID]) > 0
			safeToRemove := item.Status == models.StatusNotStarted && !hasDependents

			if safeToRemove {
				if err := s.DeleteWorkItem(ctx, cfg.Project, item.ID); err != nil {
					return fmt.Errorf("removing %s: %w", item.ID, err)
				}
			} else if syncForce {
				for _, depID := range dependentMap[item.ID] {
					depItem, getErr := s.GetWorkItem(ctx, cfg.Project, depID)
					if getErr != nil {
						return fmt.Errorf("loading dependent %s: %w", depID, getErr)
					}
					depItem.Dependencies = removeFromSlice(depItem.Dependencies, item.ID)
					depItem.UpdatedAt = now
					if err := s.UpsertWorkItem(ctx, *depItem); err != nil {
						return fmt.Errorf("updating dependent %s: %w", depID, err)
					}
				}
				if err := s.DeleteWorkItem(ctx, cfg.Project, item.ID); err != nil {
					return fmt.Errorf("force-removing %s: %w", item.ID, err)
				}
			} else {
				reason := "status=" + string(item.Status)
				if hasDependents {
					reason = fmt.Sprintf("has dependents: %v", dependentMap[item.ID])
				}
				flagged = append(flagged, fmt.Sprintf("%s (%s)", item.ID, reason))
			}
		}

		existingPhases, err := s.ListPhases(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading phases: %w", err)
		}
		existingPhaseMap := make(map[int]bool, len(existingPhases))
		for _, p := range existingPhases {
			existingPhaseMap[p.Number] = true
		}
		for _, sp := range seed.Phases {
			phase := models.Phase{
				ID:      fmt.Sprintf("%s:%d", cfg.Project, sp.Number),
				Number:  sp.Number,
				Name:    sp.Name,
				Target:  sp.Target,
				Project: cfg.Project,
			}
			if err := s.UpsertPhase(ctx, phase); err != nil {
				return fmt.Errorf("syncing phase %d: %w", sp.Number, err)
			}
		}

		fmt.Printf("Sync complete: %d added, %d modified, %d removed", len(added), len(modified), len(removed)-len(flagged))
		if len(flagged) > 0 {
			fmt.Printf(", %d flagged (use --force to remove):\n", len(flagged))
			for _, f := range flagged {
				fmt.Printf("  ⚠ %s\n", f)
			}
		} else {
			fmt.Println()
		}

		return nil
	},
}

func init() {
	syncCmd.Flags().BoolVar(&syncForce, "force", false, "Force removal of items that are in-progress or have dependents")
	rootCmd.AddCommand(syncCmd)
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

func removeFromSlice(s []string, val string) []string {
	result := make([]string, 0, len(s))
	for _, v := range s {
		if v != val {
			result = append(result, v)
		}
	}
	return result
}
