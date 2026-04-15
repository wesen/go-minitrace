/* sqleton
name: bash-commands
short: Analyze bash commands executed during GStreamer sessions
flags:
  - name: session_id
    type: string
    help: Filter to a specific session ID
  - name: command_pattern
    type: string
    help: Filter commands by pattern (e.g., go test, make%)
  - name: limit
    type: int
    default: 50
    help: Maximum number of results
*/
SELECT
  REPLACE(CAST(json_extract(tc, '$.input.command') AS VARCHAR), '"', '') AS command,
  REPLACE(CAST(json_extract(tc, '$.input.description') AS VARCHAR), '"', '') AS description,
  COUNT(*) AS execution_count,
  COUNT(DISTINCT id) AS session_count,
  MAX(REPLACE(CAST(json_extract(tc, '$.timestamp') AS VARCHAR), '"', '')) AS last_executed
FROM {{TABLE_NAME}},
     UNNEST(tool_calls) AS t(tc)
WHERE REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') = 'bash'
  AND json_extract(tc, '$.input.command') IS NOT NULL
  {{ if .session_id -}}
  AND id = {{ .session_id | sqlString }}
  {{ end -}}
  {{ if .command_pattern -}}
  AND REPLACE(CAST(json_extract(tc, '$.input.command') AS VARCHAR), '"', '') 
      LIKE {{ .command_pattern | sqlLike }}
  {{ end -}}
GROUP BY command, description
ORDER BY execution_count DESC, last_executed DESC
LIMIT {{ .limit }};
