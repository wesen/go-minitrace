package exporttimeline

import (
	"bytes"
	"testing"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func TestRenderHTML(t *testing.T) {
	exportPayload := &TimelineExport{
		Version:  "timeline-export-v1",
		Session:  SessionSummary{ID: "s1", Title: "Timeline Test"},
		Timeline: &TimelinePayload{BucketMinutes: 30, BucketCount: 1},
		Annotations: []minitrace.Annotation{
			{
				ID: "ann-1",
				Content: minitrace.AnnotationContent{
					Detail: `"</script><script>alert('oops')</script>"`,
				},
			},
		},
	}

	out, err := RenderHTML(exportPayload, RenderOptions{})
	if err != nil {
		t.Fatalf("RenderHTML returned error: %v", err)
	}
	if !bytes.Contains(out, []byte("minitrace-timeline-export-data")) {
		t.Fatalf("expected embedded timeline payload script tag")
	}
	if !bytes.Contains(out, []byte("Timeline Test")) {
		t.Fatalf("expected session title in rendered html")
	}
	if bytes.Contains(out, []byte(`</script><script>alert('oops')</script>`)) {
		t.Fatalf("expected annotation payload to be escaped for script-tag embedding")
	}
	if !bytes.Contains(out, []byte(`\u003c/script\u003e\u003cscript\u003ealert('oops')\u003c/script\u003e`)) {
		t.Fatalf("expected escaped script-like payload content")
	}
}
