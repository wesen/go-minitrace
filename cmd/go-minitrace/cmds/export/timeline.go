package export

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/exporthtml"
	timelineexport "github.com/go-go-golems/go-minitrace/pkg/exporttimeline"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newTimelineCommand() (*cobra.Command, error) {
	var archiveGlobs []string
	var queryRepositories []string
	var sessionID string
	var outputPath string
	var title string
	var readerBaseURL string
	var bucketMinutes int

	cmd := &cobra.Command{
		Use:   "timeline",
		Short: "Export a single session to a self-contained read-only timeline HTML file",
		RunE: func(cmd *cobra.Command, args []string) error {
			if len(archiveGlobs) == 0 {
				return errors.New("--archive-glob is required")
			}
			if strings.TrimSpace(sessionID) == "" {
				return errors.New("--session-id is required")
			}
			if strings.TrimSpace(outputPath) == "" {
				return errors.New("--output is required")
			}
			if len(queryRepositories) == 0 {
				return errors.New("--query-repository is required (timeline-* query commands are not embedded yet)")
			}

			index, err := exporthtml.BuildSessionIndex(archiveGlobs)
			if err != nil {
				return err
			}
			session, err := exporthtml.LoadSessionByID(index, sessionID)
			if err != nil {
				return err
			}

			sqlTimelineData, err := timelineexport.LoadSQLTimelineData(context.Background(), timelineexport.LoadOptions{
				QueryRepositories: queryRepositories,
				ArchiveGlobs:      archiveGlobs,
				SessionID:         sessionID,
				BucketMinutes:     bucketMinutes,
			})
			if err != nil {
				return err
			}

			exportPayload, err := timelineexport.BuildTimelineExport(session, sqlTimelineData, timelineexport.BuildExportOptions{
				BucketMinutes: bucketMinutes,
				ReaderBaseURL: readerBaseURL,
			})
			if err != nil {
				return err
			}

			html, err := timelineexport.RenderHTML(exportPayload, timelineexport.RenderOptions{PageTitle: title})
			if err != nil {
				return err
			}
			if err := os.MkdirAll(filepath.Dir(outputPath), 0o755); err != nil {
				return err
			}
			if err := os.WriteFile(outputPath, html, 0o644); err != nil {
				return err
			}

			fmt.Fprintf(cmd.OutOrStdout(), "Wrote %s\n", outputPath)
			return nil
		},
	}

	cmd.Flags().StringArrayVar(&archiveGlobs, "archive-glob", []string{"./output/active/*/*.minitrace.json"}, "Glob(s) matching .minitrace.json session archives")
	cmd.Flags().StringArrayVar(&queryRepositories, "query-repository", nil, "Repository path(s) containing timeline query commands")
	cmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID to export")
	cmd.Flags().IntVar(&bucketMinutes, "bucket-minutes", 30, "Timeline bucket size in minutes")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output HTML path")
	cmd.Flags().StringVar(&title, "title", "", "Optional page title override")
	cmd.Flags().StringVar(&readerBaseURL, "reader-base-url", "", "Optional chronological reader URL/path used for click navigation")

	return cmd, nil
}
