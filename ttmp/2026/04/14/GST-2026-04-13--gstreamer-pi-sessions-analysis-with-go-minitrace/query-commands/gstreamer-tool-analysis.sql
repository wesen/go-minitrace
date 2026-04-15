/* sqleton
name: gstreamer-tool-analysis
short: Analyze tool calls in GStreamer sessions by operation type and tool name
flags:
  - name: session_id
    type: string
    help: Filter to a specific session ID
  - name: operation_type
    type: stringList
    help: Filter by operation types (READ, MODIFY, NEW, EXECUTE, DELEGATE, OTHER)
  - name: limit
    type: int
    default: 100
    help: Maximum number of results
*/
SELECT
  REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') AS tool_name,
  REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation_type,
  COUNT(*) AS call_count,
  SUM(CASE WHEN json_extract(tc, '$.output.success') = true THEN 1 ELSE 0 END) AS success_count,
  ROUND(AVG(CAST(json_extract(tc, '$.output.duration_ms') AS DOUBLE)), 0) AS avg_duration_ms
FROM {{TABLE_NAME}},
     UNNEST(tool_calls) AS t(tc)
WHERE 1=1
  {{ if .session_id -}}
  AND id = {{ .session_id | sqlString }}
  {{ end -}}
  {{ if .operation_type -}}
  AND REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') 
      IN ({{ .operation_type | sqlStringIn }})
  {{ end -}}
GROUP BY tool_name, operation_type
ORDER BY call_count DESC
LIMIT {{ .limit }};
