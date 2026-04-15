/* sqleton
name: timeline-idle-windows
short: Detect contiguous low-activity windows for one session
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Bucket width in minutes
  - name: max_tool_calls_per_bucket
    type: int
    default: 0
    help: Treat buckets at or below this tool-call count as idle candidates
  - name: min_idle_buckets
    type: int
    default: 2
    help: Minimum contiguous idle buckets required to emit a window
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
    ) AS bucket_idx
  FROM session_row sr,
       UNNEST(sr.tool_calls) AS u(tc)
  WHERE json_extract(tc, '$.timestamp') IS NOT NULL
),
bucket_counts AS (
  SELECT session_id, bucket_idx, COUNT(*) AS tool_call_count
  FROM tool_events
  GROUP BY session_id, bucket_idx
),
annotated_buckets AS (
  SELECT
    b.session_id,
    b.bucket_idx,
    b.bucket_start,
    b.bucket_end,
    COALESCE(bc.tool_call_count, 0) AS tool_call_count,
    CASE WHEN COALESCE(bc.tool_call_count, 0) <= {{ .max_tool_calls_per_bucket }} THEN 1 ELSE 0 END AS is_idle
  FROM buckets b
  LEFT JOIN bucket_counts bc
    ON bc.session_id = b.session_id AND bc.bucket_idx = b.bucket_idx
),
idle_buckets AS (
  SELECT
    session_id,
    bucket_idx,
    bucket_start,
    bucket_end,
    tool_call_count,
    bucket_idx - ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY bucket_idx) AS idle_group
  FROM annotated_buckets
  WHERE is_idle = 1
)
SELECT
  session_id,
  MIN(bucket_idx) AS start_bucket_idx,
  MAX(bucket_idx) AS end_bucket_idx,
  STRFTIME(MIN(bucket_start), '%Y-%m-%dT%H:%M:%SZ') AS idle_start,
  STRFTIME(MAX(bucket_end), '%Y-%m-%dT%H:%M:%SZ') AS idle_end,
  COUNT(*) AS bucket_count,
  SUM(tool_call_count) AS tool_call_count_in_window
FROM idle_buckets
GROUP BY session_id, idle_group
HAVING COUNT(*) >= {{ .min_idle_buckets }}
ORDER BY start_bucket_idx;
