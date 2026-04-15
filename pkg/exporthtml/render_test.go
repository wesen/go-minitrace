package exporthtml

import (
	"bytes"
	"testing"
)

func TestRenderHTML(t *testing.T) {
	payload := &ReaderExport{Version: "reader-export-v1", Session: SessionDetail{ID: "s1", Title: "Example"}}
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
}
