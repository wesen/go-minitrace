/* sqleton
name: timeline-thread-signals
short: Produce simple contiguous file-based thread candidate spans over timeline buckets
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Bucket width in minutes
  - name: min_operations_per_bucket
    type: int
    default: 2
    help: Minimum operations in a bucket for a file to count as an active thread signal
  - name: min_bucket_span
    type: int
    default: 2
    help: Minimum contiguous buckets required to emit a thread candidate
  - name: top_files
    type: int
    default: 10
    help: Limit candidate generation to the top N files by total operations
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
    CASE
      WHEN LOWER(REPLACE(CAST(json_extract(tc, '$.input.file_path') AS VARCHAR), '"', '')) IN ('', 'null') THEN NULL
      ELSE REPLACE(CAST(json_extract(tc, '$.input.file_path') AS VARCHAR), '"', '')
    END AS file_path,
    CAST(
      FLOOR(
        GREATEST(
          0,
          EXTRACT(EPOCH FROM (CAST(json_extract(tc, '$.timestamp') AS TIMESTAMP) - sr.session_start)) / 60.0
        ) / {{ .bucket_minutes }}
      ) AS BIGINT
    ) AS bucket_idx,
    CAST(COALESCE(CAST(json_extract(tc, '$.emitting_turn_index') AS BIGINT), -1) AS BIGINT) AS turn_idx
  FROM session_row sr,
       UNNEST(sr.tool_calls) AS u(tc)
  WHERE json_extract(tc, '$.timestamp') IS NOT NULL
),
file_events AS (
  SELECT *
  FROM tool_events
  WHERE file_path IS NOT NULL
),
ranked_files AS (
  SELECT
    session_id,
    file_path,
    COUNT(*) AS total_operations,
    ROW_NUMBER() OVER (PARTITION BY session_id ORDER BY COUNT(*) DESC, file_path ASC) AS file_rank
  FROM file_events
  GROUP BY session_id, file_path
),
kept_files AS (
  SELECT session_id, file_path, total_operations, file_rank
  FROM ranked_files
  WHERE file_rank <= {{ .top_files }}
),
per_bucket AS (
  SELECT
    fe.session_id,
    fe.file_path,
    fe.bucket_idx,
    COUNT(*) AS operations,
    MIN(NULLIF(fe.turn_idx, -1)) AS jump_turn_idx
  FROM file_events fe
  INNER JOIN kept_files kf
    ON kf.session_id = fe.session_id AND kf.file_path = fe.file_path
  GROUP BY fe.session_id, fe.file_path, fe.bucket_idx
  HAVING COUNT(*) >= {{ .min_operations_per_bucket }}
),
segmented AS (
  SELECT
    session_id,
    file_path,
    bucket_idx,
    operations,
    jump_turn_idx,
    bucket_idx - ROW_NUMBER() OVER (PARTITION BY session_id, file_path ORDER BY bucket_idx) AS segment_id
  FROM per_bucket
)
SELECT
  s.session_id,
  s.file_path,
  MIN(s.bucket_idx) AS start_bucket_idx,
  MAX(s.bucket_idx) AS end_bucket_idx,
  COUNT(*) AS bucket_span,
  SUM(s.operations) AS total_operations,
  MIN(s.jump_turn_idx) AS jump_turn_idx
FROM segmented s
GROUP BY s.session_id, s.file_path, s.segment_id
HAVING COUNT(*) >= {{ .min_bucket_span }}
ORDER BY total_operations DESC, bucket_span DESC, start_bucket_idx;
