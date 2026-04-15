package exporttimeline

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"

	minitracecmd "github.com/go-go-golems/go-minitrace/pkg/minitracecmd"
	queryengine "github.com/go-go-golems/go-minitrace/pkg/query"
)

func LoadSQLTimelineData(ctx context.Context, opts LoadOptions) (*SQLTimelineData, error) {
	if strings.TrimSpace(opts.SessionID) == "" {
		return nil, fmt.Errorf("session ID is required")
	}
	if len(opts.ArchiveGlobs) == 0 {
		return nil, fmt.Errorf("at least one archive glob is required")
	}
	if len(opts.QueryRepositories) == 0 {
		return nil, fmt.Errorf("at least one query repository is required")
	}
	if opts.BucketMinutes <= 0 {
		opts.BucketMinutes = 30
	}
	if strings.TrimSpace(opts.DBPath) == "" {
		opts.DBPath = ":memory:"
	}
	if strings.TrimSpace(opts.TableName) == "" {
		opts.TableName = "sessions_base"
	}

	catalog, err := minitracecmd.LoadConfiguredCatalog("go-minitrace", opts.QueryRepositories)
	if err != nil {
		return nil, err
	}

	db, conn, err := queryengine.OpenConnection(ctx, opts.DBPath)
	if err != nil {
		return nil, err
	}
	defer func() { _ = conn.Close() }()
	defer func() { _ = db.Close() }()

	if err := queryengine.LoadArchive(ctx, conn, queryengine.LoadOptions{
		ArchiveGlobs:  opts.ArchiveGlobs,
		TableName:     opts.TableName,
		PersistLoaded: false,
	}); err != nil {
		return nil, err
	}

	ret := &SQLTimelineData{}
	if err := loadCommandInto(ctx, conn, catalog, "timeline-buckets", map[string]any{
		"session_id":            opts.SessionID,
		"bucket_minutes":        opts.BucketMinutes,
		"include_empty_buckets": true,
	}, &ret.Buckets, opts.TableName); err != nil {
		return nil, err
	}
	if err := loadCommandInto(ctx, conn, catalog, "timeline-file-activity", map[string]any{
		"session_id":     opts.SessionID,
		"bucket_minutes": opts.BucketMinutes,
		"min_operations": 1,
		"top_files":      20,
	}, &ret.FileActivity, opts.TableName); err != nil {
		return nil, err
	}
	if err := loadCommandInto(ctx, conn, catalog, "timeline-idle-windows", map[string]any{
		"session_id":                opts.SessionID,
		"bucket_minutes":            opts.BucketMinutes,
		"max_tool_calls_per_bucket": 0,
		"min_idle_buckets":          2,
	}, &ret.IdleWindows, opts.TableName); err != nil {
		return nil, err
	}
	if err := loadCommandInto(ctx, conn, catalog, "timeline-phase-signals", map[string]any{
		"session_id":     opts.SessionID,
		"bucket_minutes": opts.BucketMinutes,
	}, &ret.PhaseSignals, opts.TableName); err != nil {
		return nil, err
	}
	if err := loadCommandInto(ctx, conn, catalog, "timeline-thread-signals", map[string]any{
		"session_id":                opts.SessionID,
		"bucket_minutes":            opts.BucketMinutes,
		"min_operations_per_bucket": 2,
		"min_bucket_span":           2,
		"top_files":                 10,
	}, &ret.ThreadSignals, opts.TableName); err != nil {
		return nil, err
	}

	return ret, nil
}

func loadCommandInto[T any](ctx context.Context, conn *sql.Conn, catalog *minitracecmd.Catalog, commandName string, values map[string]any, target *[]T, tableName string) error {
	cmd, ok := catalog.ByName[commandName]
	if !ok {
		return fmt.Errorf("command %q not found in loaded catalog", commandName)
	}
	resolvedCmd, resolvedValues, err := minitracecmd.ResolveAliasCommand(catalog, cmd, values)
	if err != nil {
		return err
	}
	sqlText, err := minitracecmd.RenderCommand(resolvedCmd, minitracecmd.RenderContext{
		TableName: tableName,
		Values:    resolvedValues,
	})
	if err != nil {
		return err
	}
	if err := queryengine.ValidateReadOnlyQuery(sqlText); err != nil {
		return err
	}
	rows, err := queryRowsToMaps(ctx, conn, sqlText)
	if err != nil {
		return err
	}
	decoded, err := decodeRows[T](rows)
	if err != nil {
		return fmt.Errorf("decoding rows for %s: %w", commandName, err)
	}
	*target = decoded
	return nil
}

func queryRowsToMaps(ctx context.Context, conn *sql.Conn, sqlText string) ([]map[string]any, error) {
	rows, err := conn.QueryContext(ctx, sqlText)
	if err != nil {
		return nil, fmt.Errorf("executing query: %w", err)
	}
	defer func() { _ = rows.Close() }()

	columns, err := rows.Columns()
	if err != nil {
		return nil, fmt.Errorf("reading columns: %w", err)
	}

	ret := make([]map[string]any, 0)
	for rows.Next() {
		values := make([]any, len(columns))
		scanArgs := make([]any, len(columns))
		for i := range columns {
			scanArgs[i] = &values[i]
		}
		if err := rows.Scan(scanArgs...); err != nil {
			return nil, fmt.Errorf("scanning row: %w", err)
		}
		row := make(map[string]any, len(columns))
		for i, column := range columns {
			row[column] = queryengine.NormalizeValue(values[i])
		}
		ret = append(ret, row)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterating rows: %w", err)
	}
	return ret, nil
}

func decodeRows[T any](rows []map[string]any) ([]T, error) {
	payload, err := json.Marshal(rows)
	if err != nil {
		return nil, err
	}
	ret := make([]T, 0, len(rows))
	if err := json.Unmarshal(payload, &ret); err != nil {
		return nil, err
	}
	return ret, nil
}
