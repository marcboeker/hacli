package cmds

import (
	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newInfoCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "info",
		Short: "Display server information",
		RunE: func(cmd *cobra.Command, _ []string) error {
			apiData, err := getAPIInfo(cmd)
			if err != nil {
				return err
			}
			return printResult(cmd, apiData, "data")
		},
	}
}

func getAPIInfo(cmd *cobra.Command) (any, error) {
	ctx := appContext(cmd)
	return appCallInfo(ctx)
}

func appCallInfo(ctx *app.AppContext) (any, error) {
	return app.RESTCall(ctx, "get", "", nil)
}
