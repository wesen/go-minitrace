package exporttimeline

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

const (
	TimelinePhaseTag  = "timeline-phase"
	TimelineThreadTag = "timeline-thread"
)

type phaseAnnotationDetail struct {
	PhaseID        string `json:"phase_id"`
	Label          string `json:"label"`
	StartBucket    *int   `json:"start_bucket"`
	EndBucket      *int   `json:"end_bucket"`
	StartBucketIdx *int   `json:"start_bucket_idx"`
	EndBucketIdx   *int   `json:"end_bucket_idx"`
	JumpTurn       *int   `json:"jump_turn"`
	JumpTurnIdx    *int   `json:"jump_turn_idx"`
	Source         string `json:"source"`
}

type threadAnnotationDetail struct {
	ThreadID        string                    `json:"thread_id"`
	Label           string                    `json:"label"`
	FilePath        string                    `json:"file_path"`
	StartBucket     *int                      `json:"start_bucket"`
	EndBucket       *int                      `json:"end_bucket"`
	StartBucketIdx  *int                      `json:"start_bucket_idx"`
	EndBucketIdx    *int                      `json:"end_bucket_idx"`
	Segments        []threadAnnotationSegment `json:"segments"`
	JumpTurn        *int                      `json:"jump_turn"`
	JumpTurnIdx     *int                      `json:"jump_turn_idx"`
	Source          string                    `json:"source"`
	TotalOperations int                       `json:"total_operations"`
}

type threadAnnotationSegment struct {
	StartBucket    *int `json:"start_bucket"`
	EndBucket      *int `json:"end_bucket"`
	StartBucketIdx *int `json:"start_bucket_idx"`
	EndBucketIdx   *int `json:"end_bucket_idx"`
}

func ExtractManualTimelineMarkers(annotations []minitrace.Annotation) ([]TimelinePhaseMarker, []TimelineThreadSpan, error) {
	phases := []TimelinePhaseMarker{}
	threads := []TimelineThreadSpan{}
	for _, ann := range annotations {
		switch {
		case hasTag(ann.Content.Tags, TimelinePhaseTag):
			phase, err := parsePhaseAnnotation(ann)
			if err != nil {
				return nil, nil, err
			}
			phases = append(phases, phase)
		case hasTag(ann.Content.Tags, TimelineThreadTag):
			thread, err := parseThreadAnnotation(ann)
			if err != nil {
				return nil, nil, err
			}
			threads = append(threads, thread)
		}
	}
	return phases, threads, nil
}

func parsePhaseAnnotation(ann minitrace.Annotation) (TimelinePhaseMarker, error) {
	if ann.Scope.Type != "session" {
		return TimelinePhaseMarker{}, fmt.Errorf("timeline phase annotation %q must use session scope", ann.ID)
	}
	if strings.TrimSpace(ann.Content.Detail) == "" {
		return TimelinePhaseMarker{}, fmt.Errorf("timeline phase annotation %q must provide JSON detail", ann.ID)
	}
	var detail phaseAnnotationDetail
	if err := json.Unmarshal([]byte(ann.Content.Detail), &detail); err != nil {
		return TimelinePhaseMarker{}, fmt.Errorf("parsing timeline phase annotation %q: %w", ann.ID, err)
	}
	startBucketIdx, endBucketIdx, err := resolveBucketRange(detail.StartBucketIdx, detail.StartBucket, detail.EndBucketIdx, detail.EndBucket)
	if err != nil {
		return TimelinePhaseMarker{}, fmt.Errorf("timeline phase annotation %q: %w", ann.ID, err)
	}
	label := strings.TrimSpace(detail.Label)
	if label == "" {
		label = strings.TrimSpace(ann.Content.Title)
	}
	phaseID := strings.TrimSpace(detail.PhaseID)
	if phaseID == "" {
		phaseID = fmt.Sprintf("phase-annotation-%s", ann.ID)
	}
	source := strings.TrimSpace(detail.Source)
	if source == "" {
		source = "manual"
	}
	jumpTurnIdx := firstInt(detail.JumpTurnIdx, detail.JumpTurn)
	return TimelinePhaseMarker{
		PhaseID:        phaseID,
		Label:          label,
		StartBucketIdx: startBucketIdx,
		EndBucketIdx:   endBucketIdx,
		Source:         source,
		AnnotationID:   ann.ID,
		JumpTurnIdx:    jumpTurnIdx,
	}, nil
}

func parseThreadAnnotation(ann minitrace.Annotation) (TimelineThreadSpan, error) {
	if ann.Scope.Type != "session" {
		return TimelineThreadSpan{}, fmt.Errorf("timeline thread annotation %q must use session scope", ann.ID)
	}
	if strings.TrimSpace(ann.Content.Detail) == "" {
		return TimelineThreadSpan{}, fmt.Errorf("timeline thread annotation %q must provide JSON detail", ann.ID)
	}
	var detail threadAnnotationDetail
	if err := json.Unmarshal([]byte(ann.Content.Detail), &detail); err != nil {
		return TimelineThreadSpan{}, fmt.Errorf("parsing timeline thread annotation %q: %w", ann.ID, err)
	}
	segments := make([]TimelineThreadSegment, 0, len(detail.Segments))
	for _, segment := range detail.Segments {
		startBucketIdx, endBucketIdx, err := resolveBucketRange(segment.StartBucketIdx, segment.StartBucket, segment.EndBucketIdx, segment.EndBucket)
		if err != nil {
			return TimelineThreadSpan{}, fmt.Errorf("timeline thread annotation %q: %w", ann.ID, err)
		}
		segments = append(segments, TimelineThreadSegment{StartBucketIdx: startBucketIdx, EndBucketIdx: endBucketIdx})
	}
	if len(segments) == 0 {
		startBucketIdx, endBucketIdx, err := resolveBucketRange(detail.StartBucketIdx, detail.StartBucket, detail.EndBucketIdx, detail.EndBucket)
		if err != nil {
			return TimelineThreadSpan{}, fmt.Errorf("timeline thread annotation %q: %w", ann.ID, err)
		}
		segments = append(segments, TimelineThreadSegment{StartBucketIdx: startBucketIdx, EndBucketIdx: endBucketIdx})
	}
	label := strings.TrimSpace(detail.Label)
	if label == "" {
		label = strings.TrimSpace(ann.Content.Title)
	}
	threadID := strings.TrimSpace(detail.ThreadID)
	if threadID == "" {
		threadID = fmt.Sprintf("thread-annotation-%s", ann.ID)
	}
	source := strings.TrimSpace(detail.Source)
	if source == "" {
		source = "manual"
	}
	jumpTurnIdx := firstInt(detail.JumpTurnIdx, detail.JumpTurn)
	return TimelineThreadSpan{
		ThreadID:        threadID,
		Label:           label,
		FilePath:        strings.TrimSpace(detail.FilePath),
		Segments:        segments,
		Source:          source,
		AnnotationID:    ann.ID,
		TotalOperations: detail.TotalOperations,
		JumpTurnIdx:     jumpTurnIdx,
	}, nil
}

func resolveBucketRange(startBucketIdx, startBucket, endBucketIdx, endBucket *int) (int, int, error) {
	start := firstInt(startBucketIdx, startBucket)
	end := firstInt(endBucketIdx, endBucket)
	if start == nil || end == nil {
		return 0, 0, fmt.Errorf("start and end bucket values are required")
	}
	if *end < *start {
		return 0, 0, fmt.Errorf("end bucket %d is before start bucket %d", *end, *start)
	}
	return *start, *end, nil
}

func firstInt(values ...*int) *int {
	for _, value := range values {
		if value != nil {
			return value
		}
	}
	return nil
}

func hasTag(tags []string, needle string) bool {
	for _, tag := range tags {
		if strings.EqualFold(strings.TrimSpace(tag), needle) {
			return true
		}
	}
	return false
}
