/* sqleton
name: web-searches
short: Analyze web search queries during GStreamer sessions
flags:
  - name: session_id
    type: string
    help: Filter to a specific session ID
  - name: query_pattern
    type: string
    help: Filter search queries by pattern
  - name: limit
    type: int
    default: 50
    help: Maximum number of results
*/
SELECT
  id AS session_id,
  REPLACE(CAST(json_extract(tc, '$.input.query') AS VARCHAR), '"', '') AS search_query,
  REPLACE(CAST(json_extract(tc, '$.output.result_summary') AS VARCHAR), '"', '') AS result_summary,
  REPLACE(CAST(json_extract(tc, '$.timestamp') AS VARCHAR), '"', '') AS timestamp
FROM {{TABLE_NAME}},
     UNNEST(tool_calls) AS t(tc)
WHERE REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') = 'web_search'
  {{ if .session_id -}}
  AND id = {{ .session_id | sqlString }}
  {{ end -}}
  {{ if .query_pattern -}}
  AND REPLACE(CAST(json_extract(tc, '$.input.query') AS VARCHAR), '"', '') 
      LIKE {{ .query_pattern | sqlLike }}
  {{ end -}}
ORDER BY timestamp
LIMIT {{ .limit }};
