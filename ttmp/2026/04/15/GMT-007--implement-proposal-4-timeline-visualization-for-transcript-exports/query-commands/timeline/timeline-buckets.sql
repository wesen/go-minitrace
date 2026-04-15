/* sqleton
name: timeline-buckets
short: Bucket one session into fixed-width timeline windows with activity counts
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Bucket width in minutes
  - name: include_empty_buckets
    type: bool
    default: true
    help: Include buckets with zero activity between session start and end
*/
WITH session_row AS (
  SELECT
    id AS session_id,
    CAST(timing->>'started_at' AS TIMESTAMP) AS session_start,
    COALESCE(
      CAST(timing->>'ended_at' AS TIMESTAMP),
      CAST(timing->>'started_at' AS TIMESTAMP) + (CAST(timing->>'duration_seconds' AS DOUBLE) * INTERVAL '1 second')
    ) AS session_end,
    CAST(COALESCE(timing->>'duration_seconds', '0') AS DOUBLE) AS duration_seconds,
    turns,
    tool_calls
  FROM {{TABLE_NAME}}
  WHERE id = {{ .session_id | sqlString }}
),
bucket_span AS (
  SELECT
    session_id,
    session_start,
    session_end,
    GREATEST(1, CAST(CEIL(GREATEST(duration_seconds, 0) / ({{ .bucket_minutes }} * 60.0)) AS BIGINT)) AS bucket_count
  FROM session_row
),
buckets AS (
  SELECT
    bs.session_id,
    gs.bucket_idx,
    bs.session_start + (gs.bucket_idx * {{ .bucket_minutes }}) * INTERVAL '1 minute' AS bucket_start,
    bs.session_start + ((gs.bucket_idx + 1) * {{ .bucket_minutes }}) * INTERVAL '1 minute' AS bucket_end
  FROM bucket_span bs,
       generate_series(0, bs.bucket_count - 1) AS gs(bucket_idx)
),
turn_events AS (
  SELECT
    sr.session_id,
    CAST(
      FLOOR(
        GREATEST(
          0,
          EXTRACT(EPOCH FROM (CAST(json_extract(t, '$.timestamp') AS TIMESTAMP) - sr.session_start)) / 60.0
        ) / {{ .bucket_minutes }}
      ) AS BIGINT
    ) AS bucket_idx,
    CAST(json_extract(t, '$.index') AS BIGINT) AS turn_idx
  FROM session_row sr,
       UNNEST(sr.turns) AS u(t)
  WHERE json_extract(t, '$.timestamp') IS NOT NULL
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
    CASE
      WHEN LOWER(REPLACE(CAST(json_extract(tc, '$.input.file_path') AS VARCHAR), '"', '')) IN ('', 'null') THEN NULL
      ELSE REPLACE(CAST(json_extract(tc, '$.input.file_path') AS VARCHAR), '"', '')
    END AS file_path,
    CAST(COALESCE(CAST(json_extract(tc, '$.output.success') AS BOOLEAN), TRUE) AS BOOLEAN) AS success
  FROM session_row sr,
       UNNEST(sr.tool_calls) AS u(tc)
  WHERE json_extract(tc, '$.timestamp') IS NOT NULL
),
turn_bucket_counts AS (
  SELECT
    session_id,
    bucket_idx,
    COUNT(*) AS turn_count,
    MIN(turn_idx) AS jump_turn_idx
  FROM turn_events
  GROUP BY session_id, bucket_idx
),
tool_bucket_counts AS (
  SELECT
    session_id,
    bucket_idx,
    COUNT(*) AS tool_call_count,
    SUM(CASE WHEN operation_type = 'READ' THEN 1 ELSE 0 END) AS read_count,
    SUM(CASE WHEN operation_type = 'MODIFY' THEN 1 ELSE 0 END) AS modify_count,
    SUM(CASE WHEN operation_type = 'NEW' THEN 1 ELSE 0 END) AS new_count,
    SUM(CASE WHEN operation_type = 'EXECUTE' THEN 1 ELSE 0 END) AS execute_count,
    SUM(CASE WHEN success = FALSE THEN 1 ELSE 0 END) AS error_count,
    COUNT(DISTINCT file_path) FILTER (WHERE file_path IS NOT NULL) AS active_file_count
  FROM tool_events
  GROUP BY session_id, bucket_idx
)
SELECT
  b.session_id,
  b.bucket_idx,
  STRFTIME(b.bucket_start, '%Y-%m-%dT%H:%M:%SZ') AS bucket_start,
  STRFTIME(b.bucket_end, '%Y-%m-%dT%H:%M:%SZ') AS bucket_end,
  COALESCE(tbc.turn_count, 0) AS turn_count,
  COALESCE(tlc.tool_call_count, 0) AS tool_call_count,
  COALESCE(tlc.read_count, 0) AS read_count,
  COALESCE(tlc.modify_count, 0) AS modify_count,
  COALESCE(tlc.new_count, 0) AS new_count,
  COALESCE(tlc.execute_count, 0) AS execute_count,
  COALESCE(tlc.error_count, 0) AS error_count,
  COALESCE(tlc.active_file_count, 0) AS active_file_count,
  tbc.jump_turn_idx AS jump_turn_idx
FROM buckets b
LEFT JOIN turn_bucket_counts tbc
  ON tbc.session_id = b.session_id AND tbc.bucket_idx = b.bucket_idx
LEFT JOIN tool_bucket_counts tlc
  ON tlc.session_id = b.session_id AND tlc.bucket_idx = b.bucket_idx
{{ if not .include_empty_buckets -}}
WHERE COALESCE(tbc.turn_count, 0) > 0 OR COALESCE(tlc.tool_call_count, 0) > 0
{{ end -}}
ORDER BY b.bucket_idx;
