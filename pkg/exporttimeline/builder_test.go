package exporttimeline

import (
	"testing"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func TestBuildTimelineExport(t *testing.T) {
	title := "Session title"
	started := "2026-04-15T00:00:00Z"
	ended := "2026-04-15T00:30:00Z"
	session := minitrace.Session{
		ID:    "session-1",
		Title: &title,
		Timing: minitrace.Timing{
			StartedAt: &started,
			EndedAt:   &ended,
		},
		Turns:       []minitrace.Turn{{Index: 0}},
		ToolCalls:   []minitrace.ToolCall{{ID: "tc-1"}},
		Annotations: []minitrace.Annotation{},
	}
	data := &SQLTimelineData{Buckets: []BucketRow{{BucketIdx: 0}}}

	exportPayload, err := BuildTimelineExport(session, data, BuildExportOptions{BucketMinutes: 30, ReaderBaseURL: "reader.html"})
	if err != nil {
		t.Fatalf("BuildTimelineExport returned error: %v", err)
	}
	if exportPayload.Version != "timeline-export-v1" {
		t.Fatalf("Version = %q, want timeline-export-v1", exportPayload.Version)
	}
	if exportPayload.Session.ID != "session-1" {
		t.Fatalf("Session.ID = %q, want session-1", exportPayload.Session.ID)
	}
	if exportPayload.Session.Title != "Session title" {
		t.Fatalf("Session.Title = %q, want Session title", exportPayload.Session.Title)
	}
	if exportPayload.ReaderLinks.BaseURL != "reader.html" {
		t.Fatalf("ReaderLinks.BaseURL = %q, want reader.html", exportPayload.ReaderLinks.BaseURL)
	}
	if exportPayload.Annotations == nil {
		t.Fatalf("Annotations must not be nil")
	}
}
