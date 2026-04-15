package exporttimeline

import "fmt"

type OperationCounts struct {
	Read    int `json:"read"`
	Modify  int `json:"modify"`
	New     int `json:"new"`
	Execute int `json:"execute"`
}

type TimelineBucket struct {
	BucketIdx       int             `json:"bucket_idx"`
	BucketStart     string          `json:"bucket_start"`
	BucketEnd       string          `json:"bucket_end"`
	TurnCount       int             `json:"turn_count"`
	ToolCallCount   int             `json:"tool_call_count"`
	ErrorCount      int             `json:"error_count"`
	ActiveFileCount int             `json:"active_file_count"`
	OperationCounts OperationCounts `json:"operation_counts"`
	JumpTurnIdx     *int            `json:"jump_turn_idx,omitempty"`
	JumpHash        string          `json:"jump_hash,omitempty"`
}

type TimelineFileSeries struct {
	FilePath                  string   `json:"file_path"`
	FileRank                  int      `json:"file_rank"`
	TotalOperations           int      `json:"total_operations"`
	Series                    []int    `json:"series"`
	DominantOperationByBucket []string `json:"dominant_operation_by_bucket"`
}

type TimelineIdleWindow struct {
	StartBucketIdx        int    `json:"start_bucket_idx"`
	EndBucketIdx          int    `json:"end_bucket_idx"`
	IdleStart             string `json:"idle_start"`
	IdleEnd               string `json:"idle_end"`
	BucketCount           int    `json:"bucket_count"`
	ToolCallCountInWindow int    `json:"tool_call_count_in_window"`
}

type TimelinePhaseSignal struct {
	BucketIdx                 int    `json:"bucket_idx"`
	ToolCallCount             int    `json:"tool_call_count"`
	ErrorCount                int    `json:"error_count"`
	ValidationSignalCount     int    `json:"validation_signal_count"`
	CodeReadSignalCount       int    `json:"code_read_signal_count"`
	ImplementationSignalCount int    `json:"implementation_signal_count"`
	DominantPhaseSignal       string `json:"dominant_phase_signal"`
	JumpTurnIdx               *int   `json:"jump_turn_idx,omitempty"`
	JumpHash                  string `json:"jump_hash,omitempty"`
}

type TimelineThreadSignal struct {
	FilePath        string `json:"file_path"`
	StartBucketIdx  int    `json:"start_bucket_idx"`
	EndBucketIdx    int    `json:"end_bucket_idx"`
	BucketSpan      int    `json:"bucket_span"`
	TotalOperations int    `json:"total_operations"`
	JumpTurnIdx     *int   `json:"jump_turn_idx,omitempty"`
	JumpHash        string `json:"jump_hash,omitempty"`
}

type TimelinePayload struct {
	BucketMinutes int                    `json:"bucket_minutes"`
	BucketCount   int                    `json:"bucket_count"`
	Buckets       []TimelineBucket       `json:"buckets"`
	FileSeries    []TimelineFileSeries   `json:"file_series"`
	IdleWindows   []TimelineIdleWindow   `json:"idle_windows"`
	PhaseSignals  []TimelinePhaseSignal  `json:"phase_signals"`
	ThreadSignals []TimelineThreadSignal `json:"thread_signals"`
}

type BuildPayloadOptions struct {
	BucketMinutes int
}

func BuildTimelinePayload(data *SQLTimelineData, opts BuildPayloadOptions) (*TimelinePayload, error) {
	if data == nil {
		return nil, fmt.Errorf("sql timeline data is required")
	}
	if opts.BucketMinutes <= 0 {
		opts.BucketMinutes = 30
	}

	ret := &TimelinePayload{
		BucketMinutes: opts.BucketMinutes,
		BucketCount:   len(data.Buckets),
		Buckets:       make([]TimelineBucket, 0, len(data.Buckets)),
		FileSeries:    []TimelineFileSeries{},
		IdleWindows:   make([]TimelineIdleWindow, 0, len(data.IdleWindows)),
		PhaseSignals:  make([]TimelinePhaseSignal, 0, len(data.PhaseSignals)),
		ThreadSignals: make([]TimelineThreadSignal, 0, len(data.ThreadSignals)),
	}

	jumpByBucket := map[int]*int{}
	for _, row := range data.Buckets {
		ret.Buckets = append(ret.Buckets, TimelineBucket{
			BucketIdx:       row.BucketIdx,
			BucketStart:     row.BucketStart,
			BucketEnd:       row.BucketEnd,
			TurnCount:       row.TurnCount,
			ToolCallCount:   row.ToolCallCount,
			ErrorCount:      row.ErrorCount,
			ActiveFileCount: row.ActiveFileCount,
			OperationCounts: OperationCounts{Read: row.ReadCount, Modify: row.ModifyCount, New: row.NewCount, Execute: row.ExecuteCount},
			JumpTurnIdx:     row.JumpTurnIdx,
			JumpHash:        turnHash(row.JumpTurnIdx),
		})
		if row.JumpTurnIdx != nil {
			jumpByBucket[row.BucketIdx] = row.JumpTurnIdx
		}
	}

	ret.FileSeries = buildFileSeries(data.FileActivity, len(data.Buckets))
	for _, row := range data.IdleWindows {
		ret.IdleWindows = append(ret.IdleWindows, TimelineIdleWindow{
			StartBucketIdx:        row.StartBucketIdx,
			EndBucketIdx:          row.EndBucketIdx,
			IdleStart:             row.IdleStart,
			IdleEnd:               row.IdleEnd,
			BucketCount:           row.BucketCount,
			ToolCallCountInWindow: row.ToolCallCountInWindow,
		})
	}
	for _, row := range data.PhaseSignals {
		jumpTurnIdx := jumpByBucket[row.BucketIdx]
		ret.PhaseSignals = append(ret.PhaseSignals, TimelinePhaseSignal{
			BucketIdx:                 row.BucketIdx,
			ToolCallCount:             row.ToolCallCount,
			ErrorCount:                row.ErrorCount,
			ValidationSignalCount:     row.ValidationSignalCount,
			CodeReadSignalCount:       row.CodeReadSignalCount,
			ImplementationSignalCount: row.ImplementationSignalCount,
			DominantPhaseSignal:       row.DominantPhaseSignal,
			JumpTurnIdx:               jumpTurnIdx,
			JumpHash:                  turnHash(jumpTurnIdx),
		})
	}
	for _, row := range data.ThreadSignals {
		ret.ThreadSignals = append(ret.ThreadSignals, TimelineThreadSignal{
			FilePath:        row.FilePath,
			StartBucketIdx:  row.StartBucketIdx,
			EndBucketIdx:    row.EndBucketIdx,
			BucketSpan:      row.BucketSpan,
			TotalOperations: row.TotalOperations,
			JumpTurnIdx:     row.JumpTurnIdx,
			JumpHash:        turnHash(row.JumpTurnIdx),
		})
	}

	return ret, nil
}

func buildFileSeries(rows []FileActivityRow, bucketCount int) []TimelineFileSeries {
	byFile := map[string]*TimelineFileSeries{}
	order := []string{}
	for _, row := range rows {
		series, ok := byFile[row.FilePath]
		if !ok {
			series = &TimelineFileSeries{
				FilePath:                  row.FilePath,
				FileRank:                  row.FileRank,
				TotalOperations:           row.TotalOperations,
				Series:                    make([]int, bucketCount),
				DominantOperationByBucket: make([]string, bucketCount),
			}
			byFile[row.FilePath] = series
			order = append(order, row.FilePath)
		}
		if row.BucketIdx >= 0 && row.BucketIdx < bucketCount {
			series.Series[row.BucketIdx] = row.Operations
			series.DominantOperationByBucket[row.BucketIdx] = row.DominantOperation
		}
	}
	ret := make([]TimelineFileSeries, 0, len(order))
	for _, filePath := range order {
		ret = append(ret, *byFile[filePath])
	}
	return ret
}

func turnHash(turnIdx *int) string {
	if turnIdx == nil {
		return ""
	}
	return fmt.Sprintf("#turn-%d", *turnIdx)
}
