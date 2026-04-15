package exporttimeline

import (
	"testing"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func intPtr(v int) *int { return &v }

func TestBuildTimelinePayload(t *testing.T) {
	data := &SQLTimelineData{
		Buckets: []BucketRow{
			{BucketIdx: 0, BucketStart: "2026-04-15T00:00:00Z", BucketEnd: "2026-04-15T00:30:00Z", TurnCount: 2, ToolCallCount: 3, ReadCount: 1, ExecuteCount: 2, JumpTurnIdx: intPtr(5)},
			{BucketIdx: 1, BucketStart: "2026-04-15T00:30:00Z", BucketEnd: "2026-04-15T01:00:00Z", TurnCount: 1, ToolCallCount: 2, ModifyCount: 2, ActiveFileCount: 1, JumpTurnIdx: intPtr(9)},
		},
		FileActivity: []FileActivityRow{
			{FilePath: "pkg/a.go", FileRank: 1, TotalOperations: 3, BucketIdx: 0, Operations: 1, DominantOperation: "READ"},
			{FilePath: "pkg/a.go", FileRank: 1, TotalOperations: 3, BucketIdx: 1, Operations: 2, DominantOperation: "MODIFY"},
			{FilePath: "pkg/b.go", FileRank: 2, TotalOperations: 1, BucketIdx: 1, Operations: 1, DominantOperation: "READ"},
		},
		IdleWindows:   []IdleWindowRow{{StartBucketIdx: 3, EndBucketIdx: 4, BucketCount: 2, IdleStart: "2026-04-15T02:00:00Z", IdleEnd: "2026-04-15T03:00:00Z"}},
		PhaseSignals:  []PhaseSignalRow{{BucketIdx: 1, ToolCallCount: 2, DominantPhaseSignal: "implementation-heavy"}},
		ThreadSignals: []ThreadSignalRow{{FilePath: "pkg/a.go", StartBucketIdx: 0, EndBucketIdx: 1, BucketSpan: 2, TotalOperations: 3, JumpTurnIdx: intPtr(5)}},
	}

	payload, err := BuildTimelinePayload(data, BuildPayloadOptions{BucketMinutes: 30})
	if err != nil {
		t.Fatalf("BuildTimelinePayload returned error: %v", err)
	}
	if payload.BucketMinutes != 30 {
		t.Fatalf("BucketMinutes = %d, want 30", payload.BucketMinutes)
	}
	if len(payload.Buckets) != 2 {
		t.Fatalf("len(Buckets) = %d, want 2", len(payload.Buckets))
	}
	if payload.Buckets[0].JumpHash != "#turn-5" {
		t.Fatalf("bucket 0 JumpHash = %q, want #turn-5", payload.Buckets[0].JumpHash)
	}
	if len(payload.FileSeries) != 2 {
		t.Fatalf("len(FileSeries) = %d, want 2", len(payload.FileSeries))
	}
	if got := payload.FileSeries[0].Series[1]; got != 2 {
		t.Fatalf("file series bucket 1 = %d, want 2", got)
	}
	if got := payload.FileSeries[0].DominantOperationByBucket[1]; got != "MODIFY" {
		t.Fatalf("dominant op bucket 1 = %q, want MODIFY", got)
	}
	if len(payload.PhaseSignals) != 1 || payload.PhaseSignals[0].JumpHash != "#turn-9" {
		t.Fatalf("unexpected phase signals %#v", payload.PhaseSignals)
	}
	if len(payload.PhaseMarkers) != 1 || payload.PhaseMarkers[0].JumpHash != "#turn-9" {
		t.Fatalf("unexpected phase markers %#v", payload.PhaseMarkers)
	}
	if len(payload.ThreadSignals) != 1 || payload.ThreadSignals[0].JumpHash != "#turn-5" {
		t.Fatalf("unexpected thread signals %#v", payload.ThreadSignals)
	}
	if len(payload.ThreadSpans) != 1 || payload.ThreadSpans[0].JumpHash != "#turn-5" {
		t.Fatalf("unexpected thread spans %#v", payload.ThreadSpans)
	}
}

func TestBuildTimelinePayload_MergesManualAndDerivedTimelineMarkers(t *testing.T) {
	annotations := []minitrace.Annotation{
		{
			ID:        "ann-phase-1",
			Timestamp: "2026-04-15T01:10:00Z",
			Annotator: "tester",
			Scope:     minitrace.AnnotationScope{Type: "session", TargetID: "sess-1"},
			Content: minitrace.AnnotationContent{
				Category: "observation",
				Title:    "Manual implementation phase",
				Tags:     []string{TimelinePhaseTag},
				Detail:   `{"phase_id":"phase-manual-impl","start_bucket":2,"end_bucket":3}`,
			},
		},
		{
			ID:        "ann-thread-1",
			Timestamp: "2026-04-15T01:20:00Z",
			Annotator: "tester",
			Scope:     minitrace.AnnotationScope{Type: "session", TargetID: "sess-1"},
			Content: minitrace.AnnotationContent{
				Category: "observation",
				Title:    "Manual file thread",
				Tags:     []string{TimelineThreadTag},
				Detail:   `{"thread_id":"thread-manual-b","file_path":"pkg/b.go","segments":[{"start_bucket":2,"end_bucket":3}],"jump_turn_idx":99}`,
			},
		},
	}

	data := &SQLTimelineData{
		Buckets: []BucketRow{
			{BucketIdx: 0, JumpTurnIdx: intPtr(0)},
			{BucketIdx: 1, JumpTurnIdx: intPtr(10)},
			{BucketIdx: 2, JumpTurnIdx: intPtr(20)},
			{BucketIdx: 3, JumpTurnIdx: intPtr(30)},
			{BucketIdx: 4, JumpTurnIdx: intPtr(40)},
		},
		PhaseSignals: []PhaseSignalRow{
			{BucketIdx: 0, DominantPhaseSignal: "discovery-heavy"},
			{BucketIdx: 1, DominantPhaseSignal: "discovery-heavy"},
			{BucketIdx: 2, DominantPhaseSignal: "implementation-heavy"},
			{BucketIdx: 3, DominantPhaseSignal: "implementation-heavy"},
		},
		ThreadSignals: []ThreadSignalRow{
			{FilePath: "pkg/a.go", StartBucketIdx: 0, EndBucketIdx: 1, BucketSpan: 2, TotalOperations: 3, JumpTurnIdx: intPtr(0)},
			{FilePath: "pkg/a.go", StartBucketIdx: 3, EndBucketIdx: 4, BucketSpan: 2, TotalOperations: 4, JumpTurnIdx: intPtr(30)},
			{FilePath: "pkg/b.go", StartBucketIdx: 2, EndBucketIdx: 3, BucketSpan: 2, TotalOperations: 2, JumpTurnIdx: intPtr(20)},
		},
	}

	payload, err := BuildTimelinePayload(data, BuildPayloadOptions{BucketMinutes: 30, Annotations: annotations})
	if err != nil {
		t.Fatalf("BuildTimelinePayload returned error: %v", err)
	}
	if len(payload.PhaseMarkers) != 2 {
		t.Fatalf("len(PhaseMarkers) = %d, want 2", len(payload.PhaseMarkers))
	}
	if payload.PhaseMarkers[0].PhaseID != "phase-discovery-heavy-0-1" {
		t.Fatalf("phase 0 id = %q", payload.PhaseMarkers[0].PhaseID)
	}
	if payload.PhaseMarkers[1].PhaseID != "phase-manual-impl" {
		t.Fatalf("phase 1 id = %q", payload.PhaseMarkers[1].PhaseID)
	}
	if payload.PhaseMarkers[1].JumpHash != "#turn-20" {
		t.Fatalf("manual phase jump hash = %q, want #turn-20", payload.PhaseMarkers[1].JumpHash)
	}
	if len(payload.ThreadSpans) != 2 {
		t.Fatalf("len(ThreadSpans) = %d, want 2", len(payload.ThreadSpans))
	}
	if payload.ThreadSpans[0].ThreadID != "thread-file-pkg-a-go" {
		t.Fatalf("thread 0 id = %q", payload.ThreadSpans[0].ThreadID)
	}
	if got := len(payload.ThreadSpans[0].Segments); got != 2 {
		t.Fatalf("derived thread segment count = %d, want 2", got)
	}
	if payload.ThreadSpans[1].ThreadID != "thread-manual-b" {
		t.Fatalf("thread 1 id = %q", payload.ThreadSpans[1].ThreadID)
	}
	if payload.ThreadSpans[1].JumpHash != "#turn-99" {
		t.Fatalf("manual thread jump hash = %q, want #turn-99", payload.ThreadSpans[1].JumpHash)
	}
}
