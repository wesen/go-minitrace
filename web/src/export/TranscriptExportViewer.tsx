import { useEffect, useMemo, useState } from "react";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { Annotation } from "../types";
import { BlockCard } from "../components/TranscriptViewer/BlockCard";
import type { FocusedTranscriptTarget } from "../components/TranscriptViewer/types";
import type { TranscriptExportPayload } from "./types";

function buildAnnotationIndex(annotations: Annotation[]) {
  const byTurn: Record<string, Annotation[]> = {};
  const byToolCall: Record<string, Annotation[]> = {};
  const sessionScoped: Annotation[] = [];

  for (const ann of annotations) {
    if (ann.scope.type === "session") sessionScoped.push(ann);
    if (ann.scope.type === "turn") (byTurn[ann.scope.target_id] ??= []).push(ann);
    if (ann.scope.type === "tool_call") (byToolCall[ann.scope.target_id] ??= []).push(ann);
  }

  return { byTurn, byToolCall, sessionScoped };
}

function blockMatchesQuery(block: TranscriptExportPayload["session"]["blocks"][number], query: string) {
  const q = query.toLowerCase();
  if ((block.user_content || "").toLowerCase().includes(q)) return true;
  return block.turns.some((turn) => {
    if ((turn.content || "").toLowerCase().includes(q)) return true;
    if ((turn.thinking || "").toLowerCase().includes(q)) return true;
    return turn.tool_calls_in_turn.some((tc) => {
      return (tc.tool_name || "").toLowerCase().includes(q)
        || (tc.input.file_path || "").toLowerCase().includes(q)
        || (tc.input.command || "").toLowerCase().includes(q)
        || (tc.output.result || "").toLowerCase().includes(q)
        || (tc.output.error || "").toLowerCase().includes(q);
    });
  });
}

function parseFocusedTarget(hash: string): FocusedTranscriptTarget | null {
  const normalized = hash.replace(/^#/, "").trim();
  if (!normalized) return null;

  if (normalized.startsWith("turn-")) {
    const targetId = normalized.slice("turn-".length);
    if (targetId) return { scopeType: "turn", targetId, nonce: Date.now() };
  }

  if (normalized.startsWith("tool-call-")) {
    const targetId = normalized.slice("tool-call-".length);
    if (targetId) return { scopeType: "tool_call", targetId, nonce: Date.now() };
  }

  return null;
}

function blockContainsTarget(
  block: TranscriptExportPayload["session"]["blocks"][number],
  focusedTarget: FocusedTranscriptTarget | null,
) {
  if (!focusedTarget) return false;
  if (focusedTarget.scopeType === "turn") {
    return block.turns.some((turn) => String(turn.idx) === focusedTarget.targetId);
  }
  if (focusedTarget.scopeType === "tool_call") {
    return block.turns.some((turn) =>
      turn.tool_calls_in_turn.some((toolCall) => toolCall.id === focusedTarget.targetId),
    );
  }
  return false;
}

export function TranscriptExportViewer({ data }: { data: TranscriptExportPayload }) {
  const [query, setQuery] = useState("");
  const [focusedTarget, setFocusedTarget] = useState<FocusedTranscriptTarget | null>(null);
  const annotationIndex = useMemo(() => buildAnnotationIndex(data.annotations), [data.annotations]);

  useEffect(() => {
    const syncHash = () => setFocusedTarget(parseFocusedTarget(window.location.hash));
    syncHash();
    window.addEventListener("hashchange", syncHash);
    return () => window.removeEventListener("hashchange", syncHash);
  }, []);

  useEffect(() => {
    if (!focusedTarget) return;
    const elementId = focusedTarget.scopeType === "turn"
      ? `turn-${focusedTarget.targetId}`
      : `tool-call-${focusedTarget.targetId}`;
    const timer = window.setTimeout(() => {
      document.getElementById(elementId)?.scrollIntoView({ block: "center", behavior: "smooth" });
    }, 50);
    return () => window.clearTimeout(timer);
  }, [focusedTarget]);

  const visibleBlocks = useMemo(() => {
    const q = query.trim();
    if (!q) return data.session.blocks;
    return data.session.blocks.filter((block) => blockMatchesQuery(block, q));
  }, [data.session.blocks, query]);

  return (
    <Box sx={{ maxWidth: 1100, mx: "auto", p: 2, display: "flex", flexDirection: "column", gap: 2 }}>
      <Box>
        <Typography variant="h3">{data.session.title || data.session.id}</Typography>
        <Stack direction="row" spacing={1} sx={{ mt: 1, flexWrap: "wrap" }}>
          <Chip size="small" label={data.session.id.slice(0, 8)} />
          <Chip size="small" label={data.session.environment.agent_framework || "unknown"} />
          <Chip size="small" label={data.session.environment.model || "unknown-model"} />
          {data.session.operational_context.working_directory && (
            <Chip size="small" label={data.session.operational_context.working_directory} />
          )}
        </Stack>
      </Box>

      {annotationIndex.sessionScoped.length > 0 && (
        <Box>
          <Typography variant="overline">Session annotations</Typography>
          <Stack direction="row" spacing={1} sx={{ flexWrap: "wrap", mt: 0.5 }}>
            {annotationIndex.sessionScoped.map((ann) => (
              <Chip key={ann.id} size="small" label={`${ann.content.category}: ${ann.content.title}`} />
            ))}
          </Stack>
        </Box>
      )}

      <TextField
        type="search"
        label="Search transcript"
        value={query}
        onChange={(e) => setQuery(e.target.value)}
        placeholder="Search turns, commands, file paths, outputs"
        size="small"
        fullWidth
      />

      <Box>
        {visibleBlocks.map((block, i) => (
          <BlockCard
            key={block.block_num}
            block={block}
            defaultExpanded={focusedTarget == null && i === 0}
            forceExpanded={blockContainsTarget(block, focusedTarget)}
            focusedTarget={focusedTarget}
            turnAnnotations={annotationIndex.byTurn}
            toolCallAnnotations={annotationIndex.byToolCall}
            showAnnotationActions={false}
          />
        ))}
      </Box>
    </Box>
  );
}
