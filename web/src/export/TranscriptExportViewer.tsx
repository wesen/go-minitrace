import { useMemo, useState } from "react";
import Box from "@mui/material/Box";
import Chip from "@mui/material/Chip";
import Stack from "@mui/material/Stack";
import TextField from "@mui/material/TextField";
import Typography from "@mui/material/Typography";
import type { Annotation } from "../types";
import { BlockCard } from "../components/TranscriptViewer/BlockCard";
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

export function TranscriptExportViewer({ data }: { data: TranscriptExportPayload }) {
  const [query, setQuery] = useState("");
  const annotationIndex = useMemo(() => buildAnnotationIndex(data.annotations), [data.annotations]);

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
            defaultExpanded={i === 0}
            turnAnnotations={annotationIndex.byTurn}
            toolCallAnnotations={annotationIndex.byToolCall}
            showAnnotationActions={false}
          />
        ))}
      </Box>
    </Box>
  );
}
