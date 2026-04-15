/* sqleton
name: timeline-file-activity
short: Show per-file per-bucket activity for one session
flags:
  - name: session_id
    type: string
    required: true
    help: Session ID to analyze
  - name: bucket_minutes
    type: int
    default: 30
    help: Bucket width in minutes
  - name: min_operations
    type: int
    default: 1
    help: Minimum operations required to keep a file/bucket row
  - name: top_files
    type: int
    default: 20
    help: Limit output to the top N files by total operations across the session
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
    REPLACE(CAST(json_extract(tc, '$.operation_type') AS VARCHAR), '"', '') AS operation_type,
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
    SUM(CASE WHEN fe.operation_type = 'READ' THEN 1 ELSE 0 END) AS read_count,
    SUM(CASE WHEN fe.operation_type = 'MODIFY' THEN 1 ELSE 0 END) AS modify_count,
    SUM(CASE WHEN fe.operation_type = 'NEW' THEN 1 ELSE 0 END) AS new_count,
    SUM(CASE WHEN fe.operation_type = 'EXECUTE' THEN 1 ELSE 0 END) AS execute_count
  FROM file_events fe
  INNER JOIN kept_files kf
    ON kf.session_id = fe.session_id AND kf.file_path = fe.file_path
  GROUP BY fe.session_id, fe.file_path, fe.bucket_idx
  HAVING COUNT(*) >= {{ .min_operations }}
)
SELECT
  pb.session_id,
  pb.file_path,
  kf.file_rank,
  kf.total_operations,
  pb.bucket_idx,
  pb.operations,
  pb.read_count,
  pb.modify_count,
  pb.new_count,
  pb.execute_count,
  CASE
    WHEN pb.modify_count >= pb.read_count AND pb.modify_count >= pb.new_count AND pb.modify_count >= pb.execute_count THEN 'MODIFY'
    WHEN pb.read_count >= pb.new_count AND pb.read_count >= pb.execute_count THEN 'READ'
    WHEN pb.execute_count >= pb.new_count THEN 'EXECUTE'
    ELSE 'NEW'
  END AS dominant_operation
FROM per_bucket pb
INNER JOIN kept_files kf
  ON kf.session_id = pb.session_id AND kf.file_path = pb.file_path
ORDER BY kf.file_rank, pb.bucket_idx, pb.file_path;
