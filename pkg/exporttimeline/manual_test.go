package exporttimeline

import (
	"testing"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func TestExtractManualTimelineMarkers(t *testing.T) {
	annotations := []minitrace.Annotation{
		{
			ID:        "phase-annotation",
			Timestamp: "2026-04-15T01:00:00Z",
			Annotator: "tester",
			Scope:     minitrace.AnnotationScope{Type: "session", TargetID: "sess-1"},
			Content: minitrace.AnnotationContent{
				Category: "observation",
				Title:    "Debugging loop",
				Tags:     []string{TimelinePhaseTag},
				Detail:   `{"phase_id":"phase-debugging","start_bucket":4,"end_bucket":8,"jump_turn_idx":345}`,
			},
		},
		{
			ID:        "thread-annotation",
			Timestamp: "2026-04-15T01:05:00Z",
			Annotator: "tester",
			Scope:     minitrace.AnnotationScope{Type: "session", TargetID: "sess-1"},
			Content: minitrace.AnnotationContent{
				Category: "observation",
				Title:    "Preview path",
				Tags:     []string{TimelineThreadTag},
				Detail:   `{"thread_id":"thread-preview","file_path":"pkg/preview.go","segments":[{"start_bucket":1,"end_bucket":2},{"start_bucket":4,"end_bucket":5}],"jump_turn_idx":266}`,
			},
		},
	}

	phases, threads, err := ExtractManualTimelineMarkers(annotations)
	if err != nil {
		t.Fatalf("ExtractManualTimelineMarkers returned error: %v", err)
	}
	if len(phases) != 1 {
		t.Fatalf("len(phases) = %d, want 1", len(phases))
	}
	if phases[0].PhaseID != "phase-debugging" || phases[0].StartBucketIdx != 4 || phases[0].EndBucketIdx != 8 {
		t.Fatalf("unexpected phase %#v", phases[0])
	}
	if phases[0].JumpTurnIdx == nil || *phases[0].JumpTurnIdx != 345 {
		t.Fatalf("unexpected phase jump %#v", phases[0].JumpTurnIdx)
	}
	if len(threads) != 1 {
		t.Fatalf("len(threads) = %d, want 1", len(threads))
	}
	if threads[0].ThreadID != "thread-preview" || threads[0].FilePath != "pkg/preview.go" {
		t.Fatalf("unexpected thread %#v", threads[0])
	}
	if got := len(threads[0].Segments); got != 2 {
		t.Fatalf("len(segments) = %d, want 2", got)
	}
	if threads[0].JumpTurnIdx == nil || *threads[0].JumpTurnIdx != 266 {
		t.Fatalf("unexpected thread jump %#v", threads[0].JumpTurnIdx)
	}
}

func TestExtractManualTimelineMarkers_ErrorsOnWrongScope(t *testing.T) {
	annotations := []minitrace.Annotation{
		{
			ID:        "phase-annotation",
			Timestamp: "2026-04-15T01:00:00Z",
			Annotator: "tester",
			Scope:     minitrace.AnnotationScope{Type: "turn", TargetID: "12"},
			Content: minitrace.AnnotationContent{
				Category: "observation",
				Title:    "Debugging loop",
				Tags:     []string{TimelinePhaseTag},
				Detail:   `{"phase_id":"phase-debugging","start_bucket":4,"end_bucket":8}`,
			},
		},
	}

	_, _, err := ExtractManualTimelineMarkers(annotations)
	if err == nil {
		t.Fatal("expected error, got nil")
	}
}
