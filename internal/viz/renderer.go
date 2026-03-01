package viz

type Renderer interface {
	RenderPNG(mermaidText string, outputPath string) error
}
