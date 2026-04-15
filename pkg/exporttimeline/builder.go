package exporttimeline

import (
	"fmt"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func BuildTimelineExport(session minitrace.Session, data *SQLTimelineData, opts BuildExportOptions) (*TimelineExport, error) {
	if strings.TrimSpace(session.ID) == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	if data == nil {
		return nil, fmt.Errorf("sql timeline data is required")
	}

	annotations := session.Annotations
	if annotations == nil {
		annotations = []minitrace.Annotation{}
	}

	timelinePayload, err := BuildTimelinePayload(data, BuildPayloadOptions{
		BucketMinutes:      opts.BucketMinutes,
		Annotations:        annotations,
		ManualPhaseMarkers: opts.ManualPhaseMarkers,
		ManualThreadSpans:  opts.ManualThreadSpans,
	})
	if err != nil {
		return nil, err
	}

	ret := &TimelineExport{
		Version: "timeline-export-v1",
		Session: SessionSummary{
			ID:              session.ID,
			Title:           stringOr(session.Title),
			StartedAt:       stringOr(session.Timing.StartedAt),
			EndedAt:         stringOr(session.Timing.EndedAt),
			TurnCount:       len(session.Turns),
			ToolCallCount:   len(session.ToolCalls),
			AnnotationCount: len(annotations),
		},
		Timeline: timelinePayload,
		ReaderLinks: ReaderLinks{
			BaseURL: strings.TrimSpace(opts.ReaderBaseURL),
		},
		Annotations: annotations,
	}
	if strings.TrimSpace(ret.Session.Title) == "" {
		ret.Session.Title = ret.Session.ID
	}
	return ret, nil
}

func stringOr(value *string) string {
	if value == nil {
		return ""
	}
	return strings.TrimSpace(*value)
}
