import type { Annotation, SessionDetail } from "../types";

export interface TranscriptExportPayload {
  version: string;
  session: SessionDetail;
  annotations: Annotation[];
  indices: {
    session_annotation_ids: string[];
    turn_to_annotations: Record<string, string[]>;
    tool_call_to_annotations: Record<string, string[]>;
    search: {
      terms: Record<string, number[]>;
    };
  };
}
