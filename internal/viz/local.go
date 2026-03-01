package viz

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/chromedp/chromedp"
)

type LocalRenderer struct{}

func NewLocalRenderer() *LocalRenderer {
	return &LocalRenderer{}
}

func (r *LocalRenderer) RenderPNG(mermaidText string, outputPath string) error {
	html := buildHTML(mermaidText)

	tmpDir, err := os.MkdirTemp("", "mermaid-render-*")
	if err != nil {
		return fmt.Errorf("creating temp dir: %w", err)
	}
	defer os.RemoveAll(tmpDir)

	htmlPath := filepath.Join(tmpDir, "diagram.html")
	if err := os.WriteFile(htmlPath, []byte(html), 0644); err != nil {
		return fmt.Errorf("writing temp HTML: %w", err)
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	var buf []byte
	fileURL := "file://" + htmlPath

	err = chromedp.Run(ctx,
		chromedp.Navigate(fileURL),
		chromedp.WaitVisible(`.mermaid svg`, chromedp.ByQuery),
		chromedp.Screenshot(`.mermaid svg`, &buf, chromedp.NodeVisible, chromedp.ByQuery),
	)
	if err != nil {
		return fmt.Errorf("rendering diagram: %w", err)
	}

	if err := os.MkdirAll(filepath.Dir(outputPath), 0755); err != nil {
		return fmt.Errorf("creating output directory: %w", err)
	}

	if err := os.WriteFile(outputPath, buf, 0644); err != nil {
		return fmt.Errorf("writing PNG: %w", err)
	}

	return nil
}

func buildHTML(mermaidText string) string {
	escaped := strings.ReplaceAll(mermaidText, "<", "&lt;")
	escaped = strings.ReplaceAll(escaped, ">", "&gt;")

	return `<!DOCTYPE html>
<html><body>
<pre class="mermaid">` + escaped + `</pre>
<script src="https://cdn.jsdelivr.net/npm/mermaid/dist/mermaid.min.js"></script>
<script>mermaid.initialize({startOnLoad:true});</script>
</body></html>`
}
