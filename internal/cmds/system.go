package cmds

import (
	"fmt"
	"io"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newSystemCommand() *cobra.Command {
	system := &cobra.Command{Use: "system", Short: "System operations"}

	system.AddCommand(newSystemLogCommand())
	system.AddCommand(newSystemHealthCommand())

	return system
}

func newSystemLogCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "log",
		Short: "Get errors from Home Assistant",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)

			apiURL, err := app.BuildAPIURL(ctx.Server, "error_log")
			if err != nil {
				return err
			}

			req, err := ctx.HTTPClient.Get(apiURL)
			if err != nil {
				return fmt.Errorf("request failed: %w", err)
			}
			defer req.Body.Close()

			body, err := io.ReadAll(req.Body)
			if err != nil {
				return fmt.Errorf("read response: %w", err)
			}

			fmt.Fprint(ctx.Stdout, string(body))
			return nil
		},
	}
}

func newSystemHealthCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "health",
		Short: "Get system health",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)

			result, err := app.WSCall(ctx, "system_health/info", nil)
			if err != nil {
				return err
			}

			return printResult(cmd, result, "data")
		},
	}
}
