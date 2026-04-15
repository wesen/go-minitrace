/* sqleton
name: top-modified-files
short: Find most frequently modified or read files across sessions
flags:
  - name: operation_type
    type: stringList
    help: Filter by operation types (READ, MODIFY, NEW)
  - name: min_invocations
    type: int
    default: 2
    help: Minimum number of invocations to include
  - name: limit
    type: int
    default: 50
    help: Maximum number of results
*/
SELECT
  REPLACE(CAST(json_extract(tc, '$.input.file_path') AS VARCHAR), '"', '') AS file_path,
  REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation,
  COUNT(*) AS invocation_count,
  COUNT(DISTINCT id) AS session_count
FROM {{TABLE_NAME}},
     UNNEST(tool_calls) AS t(tc)
WHERE REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') IN ('read', 'edit', 'write', 'multiEdit')
  AND json_extract(tc, '$.input.file_path') IS NOT NULL
  {{ if .operation_type -}}
  AND REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') 
      IN ({{ .operation_type | sqlStringIn }})
  {{ end -}}
GROUP BY file_path, operation
HAVING COUNT(*) >= {{ .min_invocations }}
ORDER BY invocation_count DESC, session_count DESC
LIMIT {{ .limit }};
