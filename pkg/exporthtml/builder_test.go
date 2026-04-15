package exporthtml

import (
	"bytes"
	"encoding/json"
	"testing"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

func strPtr(s string) *string     { return &s }
func intPtr(i int) *int           { return &i }
func floatPtr(f float64) *float64 { return &f }

func TestBuildReaderExport(t *testing.T) {
	session := minitrace.Session{
		ID:                 "session-1",
		Classification:     "internal",
		Title:              strPtr("Example Session"),
		Environment:        minitrace.Environment{AgentFramework: strPtr("pi"), Model: strPtr("gpt-5.4")},
		OperationalContext: minitrace.OperationalContext{WorkingDirectory: strPtr("~/repo")},
		Timing:             minitrace.Timing{StartedAt: strPtr("2026-04-14T10:00:00Z"), DurationSeconds: floatPtr(120)},
		Metrics:            minitrace.Metrics{TurnCount: 2, ToolCallCount: 1},
		Turns: []minitrace.Turn{
			{Index: 0, Role: "user", Timestamp: strPtr("2026-04-14T10:00:00Z"), Content: "Please inspect recording.go"},
			{Index: 1, Role: "assistant", Timestamp: strPtr("2026-04-14T10:00:10Z"), Content: "I am reading the file now", ToolCallsInTurn: []string{"tc-1"}},
		},
		ToolCalls: []minitrace.ToolCall{{
			ID: "tc-1", ToolName: "read", OperationType: "READ", Timestamp: strPtr("2026-04-14T10:00:11Z"),
			Input:  minitrace.ToolCallInput{FilePath: strPtr("pkg/media/gst/recording.go")},
			Output: minitrace.ToolCallOutput{Success: true, Result: strPtr("package gst"), DurationMS: intPtr(12)},
		}},
		Annotations: []minitrace.Annotation{{
			ID: "ann-1", Timestamp: "2026-04-14T10:01:00Z", Annotator: "user",
			Scope:   minitrace.AnnotationScope{Type: "turn", TargetID: "1"},
			Content: minitrace.AnnotationContent{Category: "observation", Title: "Important turn", Detail: "Read the file"},
		}},
	}

	payload, err := BuildReaderExport(session)
	if err != nil {
		t.Fatalf("BuildReaderExport failed: %v", err)
	}
	if payload.Version != "reader-export-v1" {
		t.Fatalf("unexpected version: %s", payload.Version)
	}
	if got := len(payload.Session.Blocks); got != 1 {
		t.Fatalf("expected 1 block, got %d", got)
	}
	if got := len(payload.Session.Blocks[0].Turns); got != 2 {
		t.Fatalf("expected 2 turns in block, got %d", got)
	}
	if got := len(payload.Session.Blocks[0].Turns[1].ToolCallsInTurn); got != 1 {
		t.Fatalf("expected 1 tool call in assistant turn, got %d", got)
	}
	if got := payload.Indices.TurnToAnnotations["1"]; len(got) != 1 || got[0] != "ann-1" {
		t.Fatalf("unexpected turn annotation index: %#v", got)
	}
	if got := payload.Indices.Search.Terms["recording.go"]; len(got) == 0 {
		t.Fatalf("expected search term for file path, got %#v", got)
	}
}

func TestBuildReaderExportUsesEmptyAnnotationArray(t *testing.T) {
	session := minitrace.Session{
		ID:             "session-empty-ann",
		Classification: "internal",
		Turns: []minitrace.Turn{{
			Index: 0,
			Role:  "user",
		}},
	}

	payload, err := BuildReaderExport(session)
	if err != nil {
		t.Fatalf("BuildReaderExport failed: %v", err)
	}
	if payload.Annotations == nil {
		t.Fatalf("expected empty annotation slice, got nil")
	}
	if len(payload.Annotations) != 0 {
		t.Fatalf("expected no annotations, got %d", len(payload.Annotations))
	}
	b, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("json.Marshal failed: %v", err)
	}
	if string(b) == "" || !json.Valid(b) {
		t.Fatalf("expected valid payload json")
	}
	if !bytes.Contains(b, []byte(`"annotations":[]`)) {
		t.Fatalf("expected annotations to marshal as []: %s", string(b))
	}
}
