package exporttimeline

import "github.com/go-go-golems/go-minitrace/pkg/minitrace"

type SessionSummary struct {
	ID              string `json:"id"`
	Title           string `json:"title"`
	StartedAt       string `json:"started_at,omitempty"`
	EndedAt         string `json:"ended_at,omitempty"`
	TurnCount       int    `json:"turn_count"`
	ToolCallCount   int    `json:"tool_call_count"`
	AnnotationCount int    `json:"annotation_count"`
}

type ReaderLinks struct {
	BaseURL string `json:"base_url,omitempty"`
}

type TimelineExport struct {
	Version     string                 `json:"version"`
	Session     SessionSummary         `json:"session"`
	Timeline    *TimelinePayload       `json:"timeline"`
	ReaderLinks ReaderLinks            `json:"reader_links"`
	Annotations []minitrace.Annotation `json:"annotations"`
}

type BuildExportOptions struct {
	BucketMinutes      int
	ReaderBaseURL      string
	ManualPhaseMarkers []TimelinePhaseMarker
	ManualThreadSpans  []TimelineThreadSpan
}
