/* sqleton
name: timeline-phase-signals
short: Compute bucket-level temporal signals that can feed phase labeling
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Bucket width in minutes
*/
WITH session_row AS (
  SELECT
    id AS session_id,
    CAST(timing->>'started_at' AS TIMESTAMP) AS session_start,
    tool_calls
  FROM {{TABLE_NAME}}
  WHERE id = {{ .session_id | sqlString }}
),
tool_events AS (
  SELECT
    sr.session_id,
    CAST(
      FLOOR(
        GREATEST(
          0,
          EXTRACT(EPOCH FROM (CAST(json_extract(tc, '$.timestamp') AS TIMESTAMP) - sr.session_start)) / 60.0
        ) / {{ .bucket_minutes }}
      ) AS BIGINT
    ) AS bucket_idx,
    REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation_type,
    REPLACE(CAST(json_extract(tc, '$.tool_name') AS VARCHAR), '"', '') AS tool_name,
    LOWER(COALESCE(REPLACE(CAST(json_extract(tc, '$.input.command') AS VARCHAR), '"', ''), '')) AS command_text,
    CAST(COALESCE(CAST(json_extract(tc, '$.output.success') AS BOOLEAN), TRUE) AS BOOLEAN) AS success
  FROM session_row sr,
       UNNEST(sr.tool_calls) AS u(tc)
  WHERE json_extract(tc, '$.timestamp') IS NOT NULL
)
SELECT
  session_id,
  bucket_idx,
  COUNT(*) AS tool_call_count,
  SUM(CASE WHEN operation_type = 'READ' THEN 1 ELSE 0 END) AS read_count,
  SUM(CASE WHEN operation_type = 'MODIFY' THEN 1 ELSE 0 END) AS modify_count,
  SUM(CASE WHEN operation_type = 'NEW' THEN 1 ELSE 0 END) AS new_count,
  SUM(CASE WHEN operation_type = 'EXECUTE' THEN 1 ELSE 0 END) AS execute_count,
  SUM(CASE WHEN success = FALSE THEN 1 ELSE 0 END) AS error_count,
  SUM(CASE WHEN tool_name = 'bash' AND (command_text LIKE '%go test%' OR command_text LIKE '%pnpm test%' OR command_text LIKE '%pytest%' OR command_text LIKE '%validate%') THEN 1 ELSE 0 END) AS validation_signal_count,
  SUM(CASE WHEN tool_name = 'read' THEN 1 ELSE 0 END) AS code_read_signal_count,
  SUM(CASE WHEN tool_name = 'edit' OR tool_name = 'write' THEN 1 ELSE 0 END) AS implementation_signal_count,
  CASE
    WHEN SUM(CASE WHEN success = FALSE THEN 1 ELSE 0 END) >= GREATEST(1, COUNT(*) / 4) THEN 'error-heavy'
    WHEN SUM(CASE WHEN tool_name = 'bash' AND (command_text LIKE '%go test%' OR command_text LIKE '%pnpm test%' OR command_text LIKE '%pytest%' OR command_text LIKE '%validate%') THEN 1 ELSE 0 END) >= GREATEST(1, COUNT(*) / 5) THEN 'validation-heavy'
    WHEN SUM(CASE WHEN operation_type = 'MODIFY' OR operation_type = 'NEW' THEN 1 ELSE 0 END) >= GREATEST(1, COUNT(*) / 2) THEN 'implementation-heavy'
    WHEN SUM(CASE WHEN operation_type = 'READ' THEN 1 ELSE 0 END) >= GREATEST(1, COUNT(*) / 2) THEN 'discovery-heavy'
    WHEN SUM(CASE WHEN operation_type = 'EXECUTE' THEN 1 ELSE 0 END) >= GREATEST(1, COUNT(*) / 2) THEN 'execution-heavy'
    ELSE 'mixed'
  END AS dominant_phase_signal
FROM tool_events
GROUP BY session_id, bucket_idx
ORDER BY bucket_idx;
