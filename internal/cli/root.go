package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/northharbor-dev/waypoint/internal/store"
	"github.com/northharbor-dev/waypoint/internal/store/mongo"
	"github.com/spf13/cobra"
)

type Config struct {
	MongoURI       string
	Database       string
	Project        string
	RenderURL      string
	StaleThreshold time.Duration
}

var cfg Config
var agentInfo bool

var rootCmd = &cobra.Command{
	Use:   "waypoint",
	Short: "Graph-based work tracking for AI-assisted development",
	PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
		_ = godotenv.Load()

		cfg.MongoURI = os.Getenv("WAYPOINT_MONGO_URI")
		cfg.Database = os.Getenv("WAYPOINT_MONGO_DATABASE")
		cfg.Project = os.Getenv("WAYPOINT_PROJECT")
		cfg.RenderURL = os.Getenv("WAYPOINT_RENDER_URL")

		if cfg.Database == "" {
			cfg.Database = "waypoint"
		}
		if cfg.RenderURL == "" {
			cfg.RenderURL = "https://mermaid.ink"
		}
		cfg.StaleThreshold = 24 * time.Hour

		return nil
	},
	RunE: func(cmd *cobra.Command, args []string) error {
		if agentInfo {
			printAgentInfo()
			return nil
		}
		return cmd.Help()
	},
}

func init() {
	rootCmd.PersistentFlags().BoolVar(&agentInfo, "agent-info", false, "Print agent onboarding information")
}

func Execute() error {
	return rootCmd.Execute()
}

func printAgentInfo() {
	fmt.Println(`WAYPOINT - Work Tracking CLI for AI-Assisted Development

OVERVIEW:
Waypoint tracks work items as a dependency graph (DAG). Each item has a
status, an assigned role, and dependencies on other items. Use this tool
to find your next task, claim it, and mark it done.

WORKFLOW (follow this sequence):
1. Find work:    waypoint next --role <your-role>
2. Claim it:     waypoint start <WI-ID>
3. Do the work.
4. Complete it:  waypoint done <WI-ID>
If blocked:      waypoint block <WI-ID> -r "reason"
If claim fails:  Another agent took it. Run 'next' again.

COMMANDS:
  next [--role <role>]         Show tasks ready to start (all deps done)
  start <WI-ID>                Claim a task (atomic, prevents double-pickup)
  done <WI-ID>                 Mark task complete
  block <WI-ID> -r "reason"   Mark task blocked
  unblock <WI-ID>              Clear blocker
  status [--phase N]           Show all tasks and their state
  status --output json         Full DAG state as JSON (for programmatic inspection)
  report [--output file]       Generate markdown status report
  viz [--output f] [--format]  Generate DAG diagram (md or png)
  critical-path                Show longest dependency chain

ROLES:
  lead, api_dev, ui_dev, backend_dev, security,
  unit_test, integration, data_model, scenario

STATUSES:
  not_started -> in_progress (via start, deps must be done)
  in_progress -> done (via done)
  in_progress -> blocked (via block)
  blocked -> not_started (via unblock)

RULES:
- Always run 'next' before starting work.
- Always run 'start' before doing work. Never skip claiming.
- If 'start' fails, the task is taken. Run 'next' for alternatives.
- Run 'done' immediately when work is complete.
- Never modify the database directly. Always use the CLI.

ERROR HANDLING:
- If 'start' says "already claimed": run 'next' again.
- If 'done' says "task not found": the DAG was updated and your task
  was removed. Run 'next' to find new work.
- If 'next' returns nothing: all available tasks for your role are
  either done, in progress by others, or blocked. Run 'status' to
  see what's pending and report to the project lead.

ENVIRONMENT:
  WAYPOINT_MONGO_URI       MongoDB connection string (required)
  WAYPOINT_MONGO_DATABASE  Database name (default: waypoint)
  WAYPOINT_PROJECT         Project scope (required)`)
}

func getStore() (store.Store, error) {
	if err := requireConfig(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	return mongo.New(ctx, cfg.MongoURI, cfg.Database)
}

func requireConfig() error {
	if cfg.MongoURI == "" {
		return fmt.Errorf("WAYPOINT_MONGO_URI is required; set it in the environment or in a .env file")
	}
	if cfg.Project == "" {
		return fmt.Errorf("WAYPOINT_PROJECT is required; set it in the environment or in a .env file")
	}
	return nil
}
