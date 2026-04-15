import type { TranscriptExportPayload } from "./types";

export function loadEmbeddedExport(): TranscriptExportPayload {
  const el = document.getElementById("minitrace-export-data");
  if (!el?.textContent) {
    throw new Error("Missing embedded export payload");
  }
  return JSON.parse(el.textContent) as TranscriptExportPayload;
}
