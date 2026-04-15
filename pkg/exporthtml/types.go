package exporthtml

import "github.com/go-go-golems/go-minitrace/pkg/minitrace"

type SessionTiming struct {
	StartedAt             string  `json:"started_at"`
	EndedAt               *string `json:"ended_at"`
	DurationSeconds       float64 `json:"duration_seconds"`
	ActiveDurationSeconds float64 `json:"active_duration_seconds"`
	HourOfDay             int     `json:"hour_of_day"`
	DayOfWeek             int     `json:"day_of_week"`
}

type SessionMetrics struct {
	TurnCount            int  `json:"turn_count"`
	ToolCallCount        int  `json:"tool_call_count"`
	TotalInputTokens     *int `json:"total_input_tokens,omitempty"`
	TotalOutputTokens    *int `json:"total_output_tokens,omitempty"`
	TotalCacheReadTokens *int `json:"total_cache_read_tokens,omitempty"`
}

type SessionEnvironment struct {
	AgentFramework string `json:"agent_framework"`
	Model          string `json:"model"`
}

type SessionOperationalContext struct {
	WorkingDirectory string `json:"working_directory"`
	AutonomyLevel    string `json:"autonomy_level,omitempty"`
	Sandbox          *bool  `json:"sandbox,omitempty"`
}

type SessionProvenance struct {
	SourceFormat      string `json:"source_format"`
	SourcePath        string `json:"source_path"`
	OriginalSessionID string `json:"original_session_id"`
	ConvertedAt       string `json:"converted_at"`
}

type ToolCallInput struct {
	Command   string         `json:"command,omitempty"`
	Arguments map[string]any `json:"arguments,omitempty"`
	FilePath  string         `json:"file_path,omitempty"`
}

type ToolCallOutput struct {
	Success    bool    `json:"success"`
	Result     *string `json:"result"`
	Error      *string `json:"error"`
	DurationMs int     `json:"duration_ms"`
	Truncated  bool    `json:"truncated"`
}

type BadgeType string

const (
	BadgeCommit       BadgeType = "commit"
	BadgeTicketCreate BadgeType = "ticket-create"
	BadgeDocAdd       BadgeType = "doc-add"
	BadgeDiaryWrite   BadgeType = "diary-write"
	BadgeError        BadgeType = "error"
)

type ToolCall struct {
	ID            string         `json:"id"`
	ToolName      string         `json:"tool_name"`
	Timestamp     string         `json:"timestamp"`
	OperationType string         `json:"operation_type"`
	Input         ToolCallInput  `json:"input"`
	Output        ToolCallOutput `json:"output"`
	Badges        []BadgeType    `json:"badges"`
}

type TurnUsage struct {
	InputTokens     *int `json:"input_tokens,omitempty"`
	OutputTokens    *int `json:"output_tokens,omitempty"`
	CacheReadTokens *int `json:"cache_read_tokens,omitempty"`
	ReasoningTokens *int `json:"reasoning_tokens,omitempty"`
}

type Turn struct {
	Idx             int        `json:"idx"`
	Role            string     `json:"role"`
	Source          string     `json:"source"`
	Content         string     `json:"content"`
	Timestamp       string     `json:"timestamp"`
	Thinking        *string    `json:"thinking,omitempty"`
	Model           *string    `json:"model,omitempty"`
	Usage           *TurnUsage `json:"usage,omitempty"`
	ToolCallsInTurn []ToolCall `json:"tool_calls_in_turn"`
}

type BlockArtifacts struct {
	Commits        []string `json:"commits"`
	TicketsCreated []string `json:"tickets_created"`
	DocsAdded      []string `json:"docs_added"`
	DiaryWrites    int      `json:"diary_writes"`
}

type SessionBlock struct {
	BlockNum    int            `json:"block_num"`
	UserTurnIdx int            `json:"user_turn_idx"`
	UserTs      string         `json:"user_ts"`
	UserContent string         `json:"user_content"`
	AgentTurns  int            `json:"agent_turns"`
	ToolCalls   int            `json:"tool_calls"`
	GapMinutes  *float64       `json:"gap_minutes"`
	Turns       []Turn         `json:"turns"`
	Artifacts   BlockArtifacts `json:"artifacts"`
}

type SessionDetail struct {
	ID                 string                    `json:"id"`
	Title              string                    `json:"title"`
	Summary            *string                   `json:"summary"`
	Classification     string                    `json:"classification"`
	Timing             SessionTiming             `json:"timing"`
	Metrics            SessionMetrics            `json:"metrics"`
	Environment        SessionEnvironment        `json:"environment"`
	OperationalContext SessionOperationalContext `json:"operational_context"`
	Provenance         SessionProvenance         `json:"provenance"`
	Blocks             []SessionBlock            `json:"blocks"`
}

type SearchIndex struct {
	Terms map[string][]int `json:"terms"`
}

type Indices struct {
	SessionAnnotationIDs  []string            `json:"session_annotation_ids"`
	TurnToAnnotations     map[string][]string `json:"turn_to_annotations"`
	ToolCallToAnnotations map[string][]string `json:"tool_call_to_annotations"`
	Search                SearchIndex         `json:"search"`
}

type ReaderExport struct {
	Version     string                 `json:"version"`
	Session     SessionDetail          `json:"session"`
	Annotations []minitrace.Annotation `json:"annotations"`
	Indices     Indices                `json:"indices"`
}
