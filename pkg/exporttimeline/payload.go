package exporttimeline

import (
	"fmt"
	"path/filepath"
	"sort"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

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

type TimelinePhaseMarker struct {
	PhaseID        string `json:"phase_id"`
	Label          string `json:"label"`
	StartBucketIdx int    `json:"start_bucket_idx"`
	EndBucketIdx   int    `json:"end_bucket_idx"`
	Source         string `json:"source"`
	AnnotationID   string `json:"annotation_id,omitempty"`
	PhaseSignal    string `json:"phase_signal,omitempty"`
	JumpTurnIdx    *int   `json:"jump_turn_idx,omitempty"`
	JumpHash       string `json:"jump_hash,omitempty"`
}

type TimelineThreadSegment struct {
	StartBucketIdx int `json:"start_bucket_idx"`
	EndBucketIdx   int `json:"end_bucket_idx"`
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

type TimelineThreadSpan struct {
	ThreadID        string                  `json:"thread_id"`
	Label           string                  `json:"label"`
	FilePath        string                  `json:"file_path,omitempty"`
	Segments        []TimelineThreadSegment `json:"segments"`
	Source          string                  `json:"source"`
	AnnotationID    string                  `json:"annotation_id,omitempty"`
	TotalOperations int                     `json:"total_operations,omitempty"`
	JumpTurnIdx     *int                    `json:"jump_turn_idx,omitempty"`
	JumpHash        string                  `json:"jump_hash,omitempty"`
}

type TimelinePayload struct {
	BucketMinutes int                    `json:"bucket_minutes"`
	BucketCount   int                    `json:"bucket_count"`
	Buckets       []TimelineBucket       `json:"buckets"`
	FileSeries    []TimelineFileSeries   `json:"file_series"`
	IdleWindows   []TimelineIdleWindow   `json:"idle_windows"`
	PhaseSignals  []TimelinePhaseSignal  `json:"phase_signals"`
	PhaseMarkers  []TimelinePhaseMarker  `json:"phase_markers"`
	ThreadSignals []TimelineThreadSignal `json:"thread_signals"`
	ThreadSpans   []TimelineThreadSpan   `json:"thread_spans"`
}

type BuildPayloadOptions struct {
	BucketMinutes      int
	Annotations        []minitrace.Annotation
	ManualPhaseMarkers []TimelinePhaseMarker
	ManualThreadSpans  []TimelineThreadSpan
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
		PhaseMarkers:  []TimelinePhaseMarker{},
		ThreadSignals: make([]TimelineThreadSignal, 0, len(data.ThreadSignals)),
		ThreadSpans:   []TimelineThreadSpan{},
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

	annotationPhases, annotationThreads, err := ExtractManualTimelineMarkers(opts.Annotations)
	if err != nil {
		return nil, err
	}

	manualPhases := append([]TimelinePhaseMarker{}, opts.ManualPhaseMarkers...)
	manualPhases = append(manualPhases, annotationPhases...)
	manualThreads := append([]TimelineThreadSpan{}, opts.ManualThreadSpans...)
	manualThreads = append(manualThreads, annotationThreads...)

	ret.PhaseMarkers = mergePhaseMarkers(
		normalizeManualPhaseMarkers(manualPhases, jumpByBucket),
		buildDerivedPhaseMarkers(data.PhaseSignals, jumpByBucket),
	)
	ret.ThreadSpans = mergeThreadSpans(
		normalizeManualThreadSpans(manualThreads, jumpByBucket),
		buildDerivedThreadSpans(data.ThreadSignals),
	)

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

func buildDerivedPhaseMarkers(rows []PhaseSignalRow, jumpByBucket map[int]*int) []TimelinePhaseMarker {
	if len(rows) == 0 {
		return []TimelinePhaseMarker{}
	}
	ordered := append([]PhaseSignalRow(nil), rows...)
	sort.Slice(ordered, func(i, j int) bool {
		return ordered[i].BucketIdx < ordered[j].BucketIdx
	})

	ret := []TimelinePhaseMarker{}
	var current *TimelinePhaseMarker
	for _, row := range ordered {
		signal := strings.TrimSpace(row.DominantPhaseSignal)
		if signal == "" || signal == "mixed" {
			if current != nil {
				ret = append(ret, finalizeDerivedPhaseMarker(*current, jumpByBucket))
				current = nil
			}
			continue
		}
		if current == nil {
			current = &TimelinePhaseMarker{
				PhaseID:        fmt.Sprintf("phase-%s-%d-%d", signal, row.BucketIdx, row.BucketIdx),
				Label:          humanizeSignal(signal),
				StartBucketIdx: row.BucketIdx,
				EndBucketIdx:   row.BucketIdx,
				Source:         "sql-derived",
				PhaseSignal:    signal,
			}
			continue
		}
		if current.PhaseSignal == signal && row.BucketIdx == current.EndBucketIdx+1 {
			current.EndBucketIdx = row.BucketIdx
			current.PhaseID = fmt.Sprintf("phase-%s-%d-%d", signal, current.StartBucketIdx, current.EndBucketIdx)
			continue
		}
		ret = append(ret, finalizeDerivedPhaseMarker(*current, jumpByBucket))
		current = &TimelinePhaseMarker{
			PhaseID:        fmt.Sprintf("phase-%s-%d-%d", signal, row.BucketIdx, row.BucketIdx),
			Label:          humanizeSignal(signal),
			StartBucketIdx: row.BucketIdx,
			EndBucketIdx:   row.BucketIdx,
			Source:         "sql-derived",
			PhaseSignal:    signal,
		}
	}
	if current != nil {
		ret = append(ret, finalizeDerivedPhaseMarker(*current, jumpByBucket))
	}
	return ret
}

func finalizeDerivedPhaseMarker(marker TimelinePhaseMarker, jumpByBucket map[int]*int) TimelinePhaseMarker {
	marker.JumpTurnIdx = firstJumpInBucketRange(jumpByBucket, marker.StartBucketIdx, marker.EndBucketIdx)
	marker.JumpHash = turnHash(marker.JumpTurnIdx)
	return marker
}

func buildDerivedThreadSpans(rows []ThreadSignalRow) []TimelineThreadSpan {
	if len(rows) == 0 {
		return []TimelineThreadSpan{}
	}
	byFile := map[string]*TimelineThreadSpan{}
	order := []string{}
	for _, row := range rows {
		span, ok := byFile[row.FilePath]
		if !ok {
			label := filepath.Base(row.FilePath)
			if label == "." || label == string(filepath.Separator) {
				label = row.FilePath
			}
			span = &TimelineThreadSpan{
				ThreadID:    fmt.Sprintf("thread-file-%s", slugify(row.FilePath)),
				Label:       label,
				FilePath:    row.FilePath,
				Segments:    []TimelineThreadSegment{},
				Source:      "sql-derived",
				JumpTurnIdx: row.JumpTurnIdx,
				JumpHash:    turnHash(row.JumpTurnIdx),
			}
			byFile[row.FilePath] = span
			order = append(order, row.FilePath)
		}
		span.Segments = append(span.Segments, TimelineThreadSegment{StartBucketIdx: row.StartBucketIdx, EndBucketIdx: row.EndBucketIdx})
		span.TotalOperations += row.TotalOperations
		if span.JumpTurnIdx == nil && row.JumpTurnIdx != nil {
			span.JumpTurnIdx = row.JumpTurnIdx
			span.JumpHash = turnHash(row.JumpTurnIdx)
		}
	}
	ret := make([]TimelineThreadSpan, 0, len(order))
	for _, filePath := range order {
		span := *byFile[filePath]
		normalizeSegments(&span)
		ret = append(ret, span)
	}
	sort.Slice(ret, func(i, j int) bool {
		left, right := earliestSegmentStart(ret[i].Segments), earliestSegmentStart(ret[j].Segments)
		if left != right {
			return left < right
		}
		return ret[i].Label < ret[j].Label
	})
	return ret
}

func normalizeManualPhaseMarkers(markers []TimelinePhaseMarker, jumpByBucket map[int]*int) []TimelinePhaseMarker {
	ret := make([]TimelinePhaseMarker, 0, len(markers))
	for _, marker := range markers {
		if marker.Source == "" {
			marker.Source = "manual"
		}
		if marker.PhaseID == "" {
			marker.PhaseID = fmt.Sprintf("phase-%s-%d-%d", slugify(marker.Label), marker.StartBucketIdx, marker.EndBucketIdx)
		}
		if marker.Label == "" {
			marker.Label = marker.PhaseID
		}
		if marker.JumpTurnIdx == nil {
			marker.JumpTurnIdx = firstJumpInBucketRange(jumpByBucket, marker.StartBucketIdx, marker.EndBucketIdx)
		}
		marker.JumpHash = turnHash(marker.JumpTurnIdx)
		ret = append(ret, marker)
	}
	return ret
}

func normalizeManualThreadSpans(spans []TimelineThreadSpan, jumpByBucket map[int]*int) []TimelineThreadSpan {
	ret := make([]TimelineThreadSpan, 0, len(spans))
	for _, span := range spans {
		normalizeSegments(&span)
		if span.Source == "" {
			span.Source = "manual"
		}
		if span.ThreadID == "" {
			span.ThreadID = fmt.Sprintf("thread-%s", slugify(span.Label))
		}
		if span.Label == "" {
			if span.FilePath != "" {
				span.Label = filepath.Base(span.FilePath)
			} else {
				span.Label = span.ThreadID
			}
		}
		if span.JumpTurnIdx == nil {
			span.JumpTurnIdx = firstJumpForSegments(jumpByBucket, span.Segments)
		}
		span.JumpHash = turnHash(span.JumpTurnIdx)
		ret = append(ret, span)
	}
	return ret
}

func mergePhaseMarkers(manual, derived []TimelinePhaseMarker) []TimelinePhaseMarker {
	ret := append([]TimelinePhaseMarker{}, manual...)
	for _, marker := range derived {
		skip := false
		for _, existing := range manual {
			if marker.PhaseID == existing.PhaseID || rangesOverlap(marker.StartBucketIdx, marker.EndBucketIdx, existing.StartBucketIdx, existing.EndBucketIdx) {
				skip = true
				break
			}
		}
		if !skip {
			ret = append(ret, marker)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		if ret[i].StartBucketIdx != ret[j].StartBucketIdx {
			return ret[i].StartBucketIdx < ret[j].StartBucketIdx
		}
		if ret[i].EndBucketIdx != ret[j].EndBucketIdx {
			return ret[i].EndBucketIdx < ret[j].EndBucketIdx
		}
		if ret[i].Source != ret[j].Source {
			return ret[i].Source < ret[j].Source
		}
		return ret[i].Label < ret[j].Label
	})
	return ret
}

func mergeThreadSpans(manual, derived []TimelineThreadSpan) []TimelineThreadSpan {
	ret := append([]TimelineThreadSpan{}, manual...)
	for _, span := range derived {
		skip := false
		for _, existing := range manual {
			if span.ThreadID == existing.ThreadID {
				skip = true
				break
			}
			if span.FilePath != "" && span.FilePath == existing.FilePath && segmentsEqual(span.Segments, existing.Segments) {
				skip = true
				break
			}
		}
		if !skip {
			ret = append(ret, span)
		}
	}
	sort.Slice(ret, func(i, j int) bool {
		left, right := earliestSegmentStart(ret[i].Segments), earliestSegmentStart(ret[j].Segments)
		if left != right {
			return left < right
		}
		if ret[i].Source != ret[j].Source {
			return ret[i].Source < ret[j].Source
		}
		return ret[i].Label < ret[j].Label
	})
	return ret
}

func normalizeSegments(span *TimelineThreadSpan) {
	if len(span.Segments) == 0 {
		span.Segments = []TimelineThreadSegment{}
		return
	}
	segments := append([]TimelineThreadSegment(nil), span.Segments...)
	sort.Slice(segments, func(i, j int) bool {
		if segments[i].StartBucketIdx != segments[j].StartBucketIdx {
			return segments[i].StartBucketIdx < segments[j].StartBucketIdx
		}
		return segments[i].EndBucketIdx < segments[j].EndBucketIdx
	})
	span.Segments = segments
}

func earliestSegmentStart(segments []TimelineThreadSegment) int {
	if len(segments) == 0 {
		return 0
	}
	return segments[0].StartBucketIdx
}

func segmentsEqual(left, right []TimelineThreadSegment) bool {
	if len(left) != len(right) {
		return false
	}
	for i := range left {
		if left[i] != right[i] {
			return false
		}
	}
	return true
}

func firstJumpForSegments(jumpByBucket map[int]*int, segments []TimelineThreadSegment) *int {
	for _, segment := range segments {
		if jumpTurnIdx := firstJumpInBucketRange(jumpByBucket, segment.StartBucketIdx, segment.EndBucketIdx); jumpTurnIdx != nil {
			return jumpTurnIdx
		}
	}
	return nil
}

func firstJumpInBucketRange(jumpByBucket map[int]*int, start, end int) *int {
	for bucketIdx := start; bucketIdx <= end; bucketIdx++ {
		if jumpTurnIdx, ok := jumpByBucket[bucketIdx]; ok && jumpTurnIdx != nil {
			return jumpTurnIdx
		}
	}
	return nil
}

func rangesOverlap(leftStart, leftEnd, rightStart, rightEnd int) bool {
	return leftStart <= rightEnd && rightStart <= leftEnd
}

func humanizeSignal(signal string) string {
	parts := strings.Split(signal, "-")
	for i := range parts {
		if parts[i] == "" {
			continue
		}
		parts[i] = strings.ToUpper(parts[i][:1]) + parts[i][1:]
	}
	return strings.Join(parts, " ")
}

func slugify(value string) string {
	value = strings.TrimSpace(strings.ToLower(value))
	if value == "" {
		return "unknown"
	}
	replacer := strings.NewReplacer("/", "-", "\\", "-", " ", "-", "_", "-", ".", "-", ":", "-", "#", "-")
	value = replacer.Replace(value)
	for strings.Contains(value, "--") {
		value = strings.ReplaceAll(value, "--", "-")
	}
	value = strings.Trim(value, "-")
	if value == "" {
		return "unknown"
	}
	return value
}

func turnHash(turnIdx *int) string {
	if turnIdx == nil {
		return ""
	}
	return fmt.Sprintf("#turn-%d", *turnIdx)
}
