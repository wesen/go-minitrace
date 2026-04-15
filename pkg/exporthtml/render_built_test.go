package exporthtml

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestRenderHTMLFromBuiltBundle(t *testing.T) {
	dir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir, "static"), 0o755); err != nil {
		t.Fatal(err)
	}
	html := `<!doctype html><html><head><title>Old</title><script type="module" crossorigin src="/static/export-reader.js"></script></head><body><div id="root"></div><script id="minitrace-export-data" type="application/json">{"old":true}</script></body></html>`
	js := `console.log("reader bundle");`
	if err := os.WriteFile(filepath.Join(dir, "export-reader.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "static", "export-reader.js"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	out, err := RenderHTMLFromBuiltBundle(&ReaderExport{Version: "reader-export-v1", Session: SessionDetail{ID: "s1", Title: "Example"}}, dir, RenderOptions{})
	if err != nil {
		t.Fatalf("RenderHTMLFromBuiltBundle failed: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, `console.log("reader bundle");`) {
		t.Fatalf("expected inlined module script")
	}
	if !strings.Contains(text, `id="minitrace-export-data"`) || !strings.Contains(text, `reader-export-v1`) {
		t.Fatalf("expected replaced payload json")
	}
	if !strings.Contains(text, `<title>Example</title>`) {
		t.Fatalf("expected replaced title")
	}
}
