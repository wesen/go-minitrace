package exporthtml

import (
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/minitrace"
)

type rawBlockTurn struct {
	Turn        Turn
	ToolCallIDs []string
}

type rawSessionBlock struct {
	BlockNum    int
	UserTurnIdx int
	UserTs      string
	UserContent string
	AgentTurns  int
	ToolCalls   int
	GapMinutes  *float64
	Turns       []rawBlockTurn
}

var tokenPattern = regexp.MustCompile(`[A-Za-z0-9_./:-]+`)

func BuildReaderExport(session minitrace.Session) (*ReaderExport, error) {
	tcByID := make(map[string]minitrace.ToolCall, len(session.ToolCalls))
	for _, tc := range session.ToolCalls {
		tcByID[tc.ID] = tc
	}

	blocks := buildSessionBlocks(session, tcByID)
	indices := buildIndices(session, blocks)

	annotations := make([]minitrace.Annotation, 0, len(session.Annotations))
	annotations = append(annotations, session.Annotations...)

	return &ReaderExport{
		Version:     "reader-export-v1",
		Session:     normalizeSessionDetail(session, blocks),
		Annotations: annotations,
		Indices:     indices,
	}, nil
}

func normalizeSessionDetail(session minitrace.Session, blocks []SessionBlock) SessionDetail {
	return SessionDetail{
		ID:             session.ID,
		Title:          stringValue(session.Title),
		Summary:        session.Summary,
		Classification: session.Classification,
		Timing: SessionTiming{
			StartedAt:             stringValue(session.Timing.StartedAt),
			EndedAt:               session.Timing.EndedAt,
			DurationSeconds:       floatValue(session.Timing.DurationSeconds),
			ActiveDurationSeconds: floatValue(session.Timing.ActiveDurationSeconds),
			HourOfDay:             intValue(session.Timing.HourOfDay),
			DayOfWeek:             intValue(session.Timing.DayOfWeek),
		},
		Metrics: SessionMetrics{
			TurnCount:            session.Metrics.TurnCount,
			ToolCallCount:        session.Metrics.ToolCallCount,
			TotalInputTokens:     session.Metrics.TotalInputTokens,
			TotalOutputTokens:    session.Metrics.TotalOutputTokens,
			TotalCacheReadTokens: session.Metrics.TotalCacheReadTokens,
		},
		Environment: SessionEnvironment{
			AgentFramework: stringValue(session.Environment.AgentFramework),
			Model:          stringValue(session.Environment.Model),
		},
		OperationalContext: SessionOperationalContext{
			WorkingDirectory: stringValue(session.OperationalContext.WorkingDirectory),
			AutonomyLevel:    stringValue(session.OperationalContext.AutonomyLevel),
			Sandbox:          session.OperationalContext.Sandbox,
		},
		Provenance: SessionProvenance{
			SourceFormat:      session.Provenance.SourceFormat,
			SourcePath:        stringValue(session.Provenance.SourcePath),
			OriginalSessionID: stringValue(session.Provenance.OriginalSessionID),
			ConvertedAt:       session.Provenance.ConvertedAt,
		},
		Blocks: blocks,
	}
}

func buildSessionBlocks(session minitrace.Session, tcByID map[string]minitrace.ToolCall) []SessionBlock {
	rawBlocks := buildRawSessionBlocks(session, tcByID)
	blocks := make([]SessionBlock, 0, len(rawBlocks))
	for _, rawBlock := range rawBlocks {
		turns := make([]Turn, 0, len(rawBlock.Turns))
		for _, rawTurn := range rawBlock.Turns {
			turns = append(turns, rawTurn.Turn)
		}

		blocks = append(blocks, SessionBlock{
			BlockNum:    rawBlock.BlockNum,
			UserTurnIdx: rawBlock.UserTurnIdx,
			UserTs:      rawBlock.UserTs,
			UserContent: rawBlock.UserContent,
			AgentTurns:  rawBlock.AgentTurns,
			ToolCalls:   rawBlock.ToolCalls,
			GapMinutes:  rawBlock.GapMinutes,
			Turns:       turns,
			Artifacts:   detectBlockArtifacts(rawBlock, tcByID),
		})
	}
	return blocks
}

func buildRawSessionBlocks(session minitrace.Session, tcByID map[string]minitrace.ToolCall) []rawSessionBlock {
	rawBlocks := make([]rawSessionBlock, 0)
	var current *rawSessionBlock

	for _, turn := range session.Turns {
		if turn.Role == "user" {
			if current != nil {
				rawBlocks = append(rawBlocks, *current)
			}
			current = &rawSessionBlock{
				BlockNum:    len(rawBlocks) + 1,
				UserTurnIdx: turn.Index,
				UserTs:      stringValue(turn.Timestamp),
				UserContent: turn.Content,
			}
		}
		if current == nil {
			current = &rawSessionBlock{BlockNum: 1, UserTurnIdx: -1, UserTs: stringValue(turn.Timestamp)}
		}
		rawTurn := rawBlockTurn{Turn: normalizeTurn(turn, tcByID), ToolCallIDs: append([]string(nil), turn.ToolCallsInTurn...)}
		current.Turns = append(current.Turns, rawTurn)
		if turn.Role != "user" {
			current.AgentTurns++
		}
		current.ToolCalls += len(turn.ToolCallsInTurn)
	}
	if current != nil {
		rawBlocks = append(rawBlocks, *current)
	}
	for i := 1; i < len(rawBlocks); i++ {
		prev, okPrev := minitrace.ParseTimestamp(rawBlocks[i-1].UserTs)
		curr, okCurr := minitrace.ParseTimestamp(rawBlocks[i].UserTs)
		if okPrev && okCurr {
			gap := curr.Sub(prev).Minutes()
			rawBlocks[i].GapMinutes = &gap
		}
	}
	return rawBlocks
}

func normalizeTurn(turn minitrace.Turn, tcByID map[string]minitrace.ToolCall) Turn {
	toolCalls := make([]ToolCall, 0, len(turn.ToolCallsInTurn))
	for _, toolCallID := range turn.ToolCallsInTurn {
		if tc, ok := tcByID[toolCallID]; ok {
			toolCalls = append(toolCalls, normalizeToolCall(tc))
		}
	}
	return Turn{
		Idx:             turn.Index,
		Role:            turn.Role,
		Source:          stringValue(turn.Source),
		Content:         turn.Content,
		Timestamp:       stringValue(turn.Timestamp),
		Thinking:        turn.Thinking,
		Model:           turn.Model,
		Usage:           normalizeUsage(turn.Usage),
		ToolCallsInTurn: toolCalls,
	}
}

func normalizeUsage(u *minitrace.Usage) *TurnUsage {
	if u == nil {
		return nil
	}
	return &TurnUsage{InputTokens: u.InputTokens, OutputTokens: u.OutputTokens, CacheReadTokens: u.CacheReadTokens, ReasoningTokens: u.ReasoningTokens}
}

func normalizeToolCall(toolCall minitrace.ToolCall) ToolCall {
	return ToolCall{
		ID:            toolCall.ID,
		ToolName:      toolCall.ToolName,
		Timestamp:     stringValue(toolCall.Timestamp),
		OperationType: toolCall.OperationType,
		Input:         ToolCallInput{Command: stringValue(toolCall.Input.Command), Arguments: normalizeArguments(toolCall.Input.Arguments), FilePath: stringValue(toolCall.Input.FilePath)},
		Output:        ToolCallOutput{Success: toolCall.Output.Success, Result: toolCall.Output.Result, Error: toolCall.Output.Error, DurationMs: intValue(toolCall.Output.DurationMS), Truncated: toolCall.Output.Truncated},
		Badges:        detectBadges(toolCall),
	}
}

func normalizeArguments(arguments any) map[string]any {
	if arguments == nil {
		return nil
	}
	if argMap, ok := arguments.(map[string]any); ok {
		return argMap
	}
	return map[string]any{"value": arguments}
}

func buildIndices(session minitrace.Session, blocks []SessionBlock) Indices {
	idx := Indices{
		SessionAnnotationIDs:  []string{},
		TurnToAnnotations:     map[string][]string{},
		ToolCallToAnnotations: map[string][]string{},
		Search:                SearchIndex{Terms: map[string][]int{}},
	}
	for _, ann := range session.Annotations {
		switch ann.Scope.Type {
		case "session":
			idx.SessionAnnotationIDs = append(idx.SessionAnnotationIDs, ann.ID)
		case "turn":
			idx.TurnToAnnotations[ann.Scope.TargetID] = append(idx.TurnToAnnotations[ann.Scope.TargetID], ann.ID)
		case "tool_call":
			idx.ToolCallToAnnotations[ann.Scope.TargetID] = append(idx.ToolCallToAnnotations[ann.Scope.TargetID], ann.ID)
		}
	}
	for _, block := range blocks {
		for _, turn := range block.Turns {
			indexText(idx.Search.Terms, turn.Content, turn.Idx)
			indexText(idx.Search.Terms, stringValue(turn.Thinking), turn.Idx)
			for _, tc := range turn.ToolCallsInTurn {
				indexText(idx.Search.Terms, tc.ToolName, turn.Idx)
				indexText(idx.Search.Terms, tc.Input.FilePath, turn.Idx)
				indexText(idx.Search.Terms, tc.Input.Command, turn.Idx)
				if tc.Output.Result != nil {
					indexText(idx.Search.Terms, *tc.Output.Result, turn.Idx)
				}
				if tc.Output.Error != nil {
					indexText(idx.Search.Terms, *tc.Output.Error, turn.Idx)
				}
			}
		}
	}
	for term, values := range idx.Search.Terms {
		sort.Ints(values)
		idx.Search.Terms[term] = dedupeInts(values)
	}
	return idx
}

func indexText(dst map[string][]int, text string, turnIdx int) {
	for _, token := range tokenPattern.FindAllString(strings.ToLower(text), -1) {
		if len(token) < 2 {
			continue
		}
		dst[token] = append(dst[token], turnIdx)
	}
}

func dedupeInts(values []int) []int {
	if len(values) == 0 {
		return values
	}
	out := values[:1]
	for _, v := range values[1:] {
		if out[len(out)-1] != v {
			out = append(out, v)
		}
	}
	return out
}

func stringValue(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func intValue(value *int) int {
	if value == nil {
		return 0
	}
	return *value
}

func floatValue(value *float64) float64 {
	if value == nil {
		return 0
	}
	return *value
}

var (
	commitMessagePattern = regexp.MustCompile(`git commit(?:\s+-[^\s]+\s+)*\s+-m\s+["']([^"']+)["']`)
	ticketPattern        = regexp.MustCompile(`--ticket\s+([A-Za-z0-9._-]+)`)
	titlePattern         = regexp.MustCompile(`--title\s+["']([^"']+)["']`)
)

func detectBadges(toolCall minitrace.ToolCall) []BadgeType {
	badges := make([]BadgeType, 0, 2)
	command := strings.ToLower(extractCommand(toolCall))
	if !toolCall.Output.Success {
		badges = append(badges, BadgeError)
	}
	if strings.Contains(command, "git commit") {
		badges = append(badges, BadgeCommit)
	}
	if strings.Contains(command, "docmgr ticket create") || strings.Contains(command, "docmgr ticket create-ticket") {
		badges = append(badges, BadgeTicketCreate)
	}
	if strings.Contains(command, "docmgr doc add") {
		badges = append(badges, BadgeDocAdd)
	}
	if isDiaryWrite(toolCall, command) {
		badges = append(badges, BadgeDiaryWrite)
	}
	return badges
}

func detectBlockArtifacts(block rawSessionBlock, tcByID map[string]minitrace.ToolCall) BlockArtifacts {
	artifacts := BlockArtifacts{Commits: []string{}, TicketsCreated: []string{}, DocsAdded: []string{}}
	commitSeen := map[string]struct{}{}
	ticketSeen := map[string]struct{}{}
	docSeen := map[string]struct{}{}
	for _, turn := range block.Turns {
		for _, toolCallID := range turn.ToolCallIDs {
			toolCall, ok := tcByID[toolCallID]
			if !ok {
				continue
			}
			for _, badge := range detectBadges(toolCall) {
				switch badge {
				case BadgeCommit:
					message := extractCommitMessage(toolCall)
					if message == "" {
						continue
					}
					if _, ok := commitSeen[message]; ok {
						continue
					}
					commitSeen[message] = struct{}{}
					artifacts.Commits = append(artifacts.Commits, message)
				case BadgeTicketCreate:
					ticketID := extractTicketID(toolCall)
					if ticketID == "" {
						continue
					}
					if _, ok := ticketSeen[ticketID]; ok {
						continue
					}
					ticketSeen[ticketID] = struct{}{}
					artifacts.TicketsCreated = append(artifacts.TicketsCreated, ticketID)
				case BadgeDocAdd:
					title := extractDocTitle(toolCall)
					if title == "" {
						continue
					}
					if _, ok := docSeen[title]; ok {
						continue
					}
					docSeen[title] = struct{}{}
					artifacts.DocsAdded = append(artifacts.DocsAdded, title)
				case BadgeDiaryWrite:
					artifacts.DiaryWrites++
				case BadgeError:
					continue
				}
			}
		}
	}
	return artifacts
}

func extractCommand(toolCall minitrace.ToolCall) string {
	if toolCall.Input.Command != nil {
		return *toolCall.Input.Command
	}
	if argMap, ok := toolCall.Input.Arguments.(map[string]any); ok {
		if command, ok := argMap["cmd"].(string); ok {
			return command
		}
	}
	return ""
}

func extractCommitMessage(toolCall minitrace.ToolCall) string {
	m := commitMessagePattern.FindStringSubmatch(extractCommand(toolCall))
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func extractTicketID(toolCall minitrace.ToolCall) string {
	m := ticketPattern.FindStringSubmatch(extractCommand(toolCall))
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func extractDocTitle(toolCall minitrace.ToolCall) string {
	m := titlePattern.FindStringSubmatch(extractCommand(toolCall))
	if len(m) == 2 {
		return m[1]
	}
	return ""
}

func isDiaryWrite(toolCall minitrace.ToolCall, command string) bool {
	filePath := strings.ToLower(stringValue(toolCall.Input.FilePath))
	if strings.Contains(filePath, "diary") && (strings.EqualFold(toolCall.OperationType, "modify") || strings.EqualFold(toolCall.OperationType, "new") || strings.EqualFold(toolCall.OperationType, "create")) {
		return true
	}
	if strings.Contains(command, "diary") || strings.Contains(command, "diary.md") {
		if strings.Contains(command, "apply_patch") || strings.Contains(command, "tee ") || strings.Contains(command, "cat >") || strings.Contains(command, "cat >>") {
			return true
		}
	}
	return false
}

func MarshalPayload(payload *ReaderExport) ([]byte, error) {
	if payload == nil {
		return nil, fmt.Errorf("payload is required")
	}
	return json.Marshal(payload)
}
