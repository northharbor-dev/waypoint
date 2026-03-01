package viz

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

type CloudRenderer struct {
	BaseURL string
}

func NewCloudRenderer(baseURL string) *CloudRenderer {
	if baseURL == "" {
		baseURL = "https://mermaid.ink"
	}
	return &CloudRenderer{BaseURL: baseURL}
}

func (r *CloudRenderer) RenderPNG(mermaidText string, outputPath string) error {
	encoded := base64.URLEncoding.EncodeToString([]byte(mermaidText))
	url := fmt.Sprintf("%s/img/%s", r.BaseURL, encoded)

	resp, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("fetching diagram: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("mermaid.ink returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("reading response: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, body, 0644); err != nil {
		return fmt.Errorf("writing PNG: %w", err)
	}

	return nil
}
