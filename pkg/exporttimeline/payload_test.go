package exporttimeline

import "testing"

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
	if len(payload.ThreadSignals) != 1 || payload.ThreadSignals[0].JumpHash != "#turn-5" {
		t.Fatalf("unexpected thread signals %#v", payload.ThreadSignals)
	}
}
