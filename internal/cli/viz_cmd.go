package cli

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/northharbor-dev/waypoint/internal/viz"
	"github.com/spf13/cobra"
)

var (
	vizOutput   string
	vizFormat   string
	vizRenderer string
	vizRenderURL string
)

var vizCmd = &cobra.Command{
	Use:   "viz",
	Short: "Generate a DAG visualization",
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

		phases, err := s.ListPhases(ctx, cfg.Project)
		if err != nil {
			return fmt.Errorf("loading phases: %w", err)
		}

		mermaidText := viz.GenerateMermaid(items, phases)

		switch vizFormat {
		case "md":
			content := "```mermaid\n" + mermaidText + "```\n"
			if vizOutput != "" {
				if err := os.WriteFile(vizOutput, []byte(content), 0644); err != nil {
					return fmt.Errorf("writing markdown: %w", err)
				}
				fmt.Printf("Diagram written to %s\n", vizOutput)
				return nil
			}
			fmt.Print(content)
			return nil

		case "png":
			if vizOutput == "" {
				return fmt.Errorf("--output is required for png format")
			}

			var renderer viz.Renderer
			switch vizRenderer {
			case "cloud":
				renderURL := vizRenderURL
				if renderURL == "" {
					renderURL = cfg.RenderURL
				}
				renderer = viz.NewCloudRenderer(renderURL)
			default:
				renderer = viz.NewLocalRenderer()
			}

			if err := renderer.RenderPNG(mermaidText, vizOutput); err != nil {
				return fmt.Errorf("rendering PNG: %w", err)
			}
			fmt.Printf("Diagram written to %s\n", vizOutput)
			return nil

		default:
			return fmt.Errorf("unsupported format %q (use md or png)", vizFormat)
		}
	},
}

func init() {
	vizCmd.Flags().StringVar(&vizOutput, "output", "", "Output file path")
	vizCmd.Flags().StringVar(&vizFormat, "format", "md", "Output format (md or png)")
	vizCmd.Flags().StringVar(&vizRenderer, "renderer", "local", "PNG renderer (local or cloud)")
	vizCmd.Flags().StringVar(&vizRenderURL, "render-url", "", "Cloud renderer URL")
	rootCmd.AddCommand(vizCmd)
}
