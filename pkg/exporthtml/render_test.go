package exporthtml

import (
	"bytes"
	"encoding/json"
	"testing"
)

func TestRenderHTML(t *testing.T) {
	payload := &ReaderExport{
		Version: "reader-export-v1",
		Session: SessionDetail{
			ID:    "s1",
			Title: "Example",
			Blocks: []SessionBlock{
				{
					Turns: []Turn{
						{
							Idx:  1,
							Role: "assistant",
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
	out, err := RenderHTML(payload, RenderOptions{})
	if err != nil {
		t.Fatalf("RenderHTML failed: %v", err)
	}
	if !bytes.Contains(out, []byte("minitrace-export-data")) {
		t.Fatalf("expected embedded payload script tag")
	}
	if !bytes.Contains(out, []byte("Example")) {
		t.Fatalf("expected title in rendered html")
	}
	if bytes.Contains(out, []byte(`</script><script>bad()</script>`)) {
		t.Fatalf("expected payload json to be escaped for script-tag embedding")
	}
	if !bytes.Contains(out, []byte(`\u003c/script\u003e\u003cscript\u003ebad()\u003c/script\u003e`)) {
		t.Fatalf("expected escaped script-like payload content")
	}
}
