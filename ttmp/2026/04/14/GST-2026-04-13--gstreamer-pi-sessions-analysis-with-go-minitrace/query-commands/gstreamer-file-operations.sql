/* sqleton
name: gstreamer-file-operations
short: Analyze file read/modify patterns in GStreamer sessions
flags:
  - name: session_id
    type: string
    help: Filter to a specific session ID
  - name: file_pattern
    type: string
    help: Filter files by pattern (e.g., %.go, %.rs, %.ts)
  - name: limit
    type: int
    default: 50
    help: Maximum number of results
*/
SELECT
  REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') AS tool_name,
  REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation,
  REPLACE(CAST(json_extract(tc, '$.input.path') AS VARCHAR), '"', '') AS file_path,
  REPLACE(CAST(json_extract(tc, '$.input.description') AS VARCHAR), '"', '') AS description,
  COUNT(*) AS invocation_count,
  MAX(REPLACE(CAST(json_extract(tc, '$.timestamp') AS VARCHAR), '"', '')) AS last_accessed
FROM {{TABLE_NAME}},
     UNNEST(tool_calls) AS t(tc)
WHERE 1=1
  AND REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') IN (
    'read', 'edit', 'write', 'multiEdit', 'bash'
  )
  {{ if .session_id -}}
  AND id = {{ .session_id | sqlString }}
  {{ end -}}
  {{ if .file_pattern -}}
  AND REPLACE(CAST(json_extract(tc, '$.input.path') AS VARCHAR), '"', '') 
      LIKE {{ .file_pattern | sqlLike }}
  {{ end -}}
GROUP BY tool_name, operation, file_path, description
ORDER BY invocation_count DESC, last_accessed DESC
LIMIT {{ .limit }};
