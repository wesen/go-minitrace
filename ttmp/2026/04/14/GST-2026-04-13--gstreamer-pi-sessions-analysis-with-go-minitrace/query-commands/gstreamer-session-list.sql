/* sqleton
name: gstreamer-session-list
short: List GStreamer-related sessions with detailed metrics
flags:
  - name: date_from
    type: string
    help: Filter sessions starting from this date (YYYY-MM-DD)
  - name: date_to
    type: string
    help: Filter sessions up to this date (YYYY-MM-DD)
  - name: min_duration
    type: int
    help: Minimum session duration in seconds
  - name: limit
    type: int
    default: 50
    help: Maximum number of sessions to return
*/
SELECT
  id AS session_id,
  environment->>'agent_framework' AS framework,
  environment->>'model' AS model,
  title,
  CAST(metrics->>'turn_count' AS INT) AS turns,
  CAST(metrics->>'tool_call_count' AS INT) AS tools,
  ROUND(CAST(timing->>'duration_seconds' AS DOUBLE), 0) AS duration_s,
  ROUND(CAST(metrics->>'read_ratio' AS DOUBLE), 2) AS read_ratio,
  timing->>'started_at' AS started_at,
  provenance->>'source_format' AS source_format
FROM {{TABLE_NAME}}
WHERE 1=1
  AND (title LIKE '%gstreamer%' 
       OR title LIKE '%GStreamer%' 
       OR title LIKE '%gst%'
       OR title LIKE '%pipeline%'
       OR title LIKE '%screencast%')
  {{ if .date_from -}}
  AND timing->>'started_at' >= {{ .date_from | sqlString }}
  {{ end -}}
  {{ if .date_to -}}
  AND timing->>'started_at' < {{ .date_to | sqlString }}
  {{ end -}}
  {{ if .min_duration -}}
  AND CAST(timing->>'duration_seconds' AS DOUBLE) >= {{ .min_duration }}
  {{ end -}}
ORDER BY timing->>'started_at' DESC
LIMIT {{ .limit }};
