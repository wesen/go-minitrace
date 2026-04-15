package export

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/go-go-golems/go-minitrace/pkg/exporthtml"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
)

func newHTMLCommand() (*cobra.Command, error) {
	var archiveGlobs []string
	var sessionID string
	var outputPath string
	var title string
	var webDistDir string

	cmd := &cobra.Command{
		Use:   "html",
		Short: "Export a single session to a self-contained read-only HTML file",
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

			index, err := exporthtml.BuildSessionIndex(archiveGlobs)
			if err != nil {
				return err
			}
			session, err := exporthtml.LoadSessionByID(index, sessionID)
			if err != nil {
				return err
			}
			payload, err := exporthtml.BuildReaderExport(session)
			if err != nil {
				return err
			}
			var html []byte
			if strings.TrimSpace(webDistDir) != "" {
				html, err = exporthtml.RenderHTMLFromBuiltBundle(payload, webDistDir, exporthtml.RenderOptions{PageTitle: title})
			} else {
				html, err = exporthtml.RenderHTML(payload, exporthtml.RenderOptions{PageTitle: title})
			}
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
	cmd.Flags().StringVar(&sessionID, "session-id", "", "Session ID to export")
	cmd.Flags().StringVar(&outputPath, "output", "", "Output HTML path")
	cmd.Flags().StringVar(&title, "title", "", "Optional page title override")
	cmd.Flags().StringVar(&webDistDir, "web-dist-dir", "", "Optional built export-reader bundle directory to inline (e.g. web/dist-export-reader)")
	return cmd, nil
}
