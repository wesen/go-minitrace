package exporthtml

import (
	"encoding/json"
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
	html := `<!doctype html><html><head><title>Old</title><link rel="icon" href="/favicon.svg"><script type="module" crossorigin src="/static/export-reader.js"></script><script src="/extra.js"></script></head><body><div id="root"></div><script id="minitrace-export-data" type="application/json">{"old":true}</script></body></html>`
	js := `console.log("reader bundle $1");`
	if err := os.WriteFile(filepath.Join(dir, "export-reader.html"), []byte(html), 0o644); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(dir, "static", "export-reader.js"), []byte(js), 0o644); err != nil {
		t.Fatal(err)
	}
	payload := &ReaderExport{
		Version: "reader-export-v1",
		Session: SessionDetail{
			ID:    "s1",
			Title: "Example",
			Blocks: []SessionBlock{
				{
					Turns: []Turn{
						{
							Idx:     1,
							Role:    "assistant",
							Content: `cd \"$(dirname \"$0\")\"`,
							ToolCallsInTurn: []ToolCall{
								{
									ID:       "tc1",
									ToolName: "write",
									Input: ToolCallInput{
										Arguments: map[string]any{
											"content": json.RawMessage(`"</script><script>bad()</script>"`),
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}
	out, err := RenderHTMLFromBuiltBundle(payload, dir, RenderOptions{})
	if err != nil {
		t.Fatalf("RenderHTMLFromBuiltBundle failed: %v", err)
	}
	text := string(out)
	if !strings.Contains(text, `console.log("reader bundle $1");`) {
		t.Fatalf("expected inlined module script to keep literal $ sequences")
	}
	if strings.Contains(text, `/favicon.svg`) || strings.Contains(text, `/extra.js`) {
		t.Fatalf("expected external asset references to be stripped")
	}
	if strings.Count(text, `id="minitrace-export-data"`) != 1 {
		t.Fatalf("expected exactly one embedded payload script tag")
	}
	if strings.Contains(text, `</script><script>bad()</script>`) {
		t.Fatalf("expected payload json to be escaped for script-tag embedding")
	}
	if !strings.Contains(text, `\u003c/script\u003e\u003cscript\u003ebad()\u003c/script\u003e`) {
		t.Fatalf("expected escaped script-like payload content")
	}
	if !strings.Contains(text, `$(dirname \\\"$0\\\")`) {
		t.Fatalf("expected payload content to keep literal $0 sequences")
	}
	if !strings.Contains(text, `<title>Example</title>`) {
		t.Fatalf("expected replaced title")
	}
}
