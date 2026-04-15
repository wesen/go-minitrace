/* sqleton
name: session-timeline
short: Show timeline of tool call operations over session duration
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Time bucket size in minutes for aggregating activity
*/
WITH tool_timeline AS (
  SELECT
    id AS session_id,
    REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') AS tool_name,
    REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation_type,
    REPLACE(CAST(json_extract(tc, '$.timestamp') AS VARCHAR), '"', '') AS timestamp_str,
    CAST(timing->>'started_at' AS TIMESTAMP) AS session_start
  FROM {{TABLE_NAME}},
       UNNEST(tool_calls) AS t(tc)
  WHERE id = {{ .session_id | sqlString }}
)
SELECT
  session_id,
  tool_name,
  operation_type,
  COUNT(*) AS operations,
  MIN(timestamp_str) AS first_in_bucket,
  MAX(timestamp_str) AS last_in_bucket,
  ROUND((EXTRACT(EPOCH FROM CAST(MIN(timestamp_str) AS TIMESTAMP) - session_start) / 60.0)::DOUBLE, 0) AS minutes_from_start
FROM tool_timeline
GROUP BY session_id, tool_name, operation_type, 
         ROUND((EXTRACT(EPOCH FROM CAST(timestamp_str AS TIMESTAMP) - session_start) / 60.0 / {{ .bucket_minutes }}))::INT
ORDER BY first_in_bucket;
