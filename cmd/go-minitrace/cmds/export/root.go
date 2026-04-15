package export

import "github.com/spf13/cobra"

func NewCommand() (*cobra.Command, error) {
	root := &cobra.Command{
		Use:   "export",
		Short: "Export minitrace sessions to read-only artifacts",
	}
	htmlCmd, err := newHTMLCommand()
	if err != nil {
		return nil, err
	}
	timelineCmd, err := newTimelineCommand()
	if err != nil {
		return nil, err
	}
	root.AddCommand(htmlCmd)
	root.AddCommand(timelineCmd)
	return root, nil
}
