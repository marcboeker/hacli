package cmds

import (
	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newRawCommand() *cobra.Command {
	raw := &cobra.Command{Use: "raw", Short: "Raw API access"}

	getCmd := &cobra.Command{
		Use:   "get <method>",
		Short: "Raw GET",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", args[0], nil)
			if err != nil {
				return err
			}
			return printResult(cmd, result, "data")
		},
	}

	var postJSON string
	postCmd := &cobra.Command{
		Use:   "post <method>",
		Short: "Raw POST",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := app.ParseObject(postJSON)
			if err != nil {
				return err
			}

			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "post", args[0], payload)
			if err != nil {
				return err
			}
			return printResult(cmd, result, "data")
		},
	}
	postCmd.Flags().StringVar(&postJSON, "json", "", "JSON payload")

	var wsJSON string
	wsCmd := &cobra.Command{
		Use:   "ws <wstype>",
		Short: "Raw websocket call",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := app.ParseObject(wsJSON)
			if err != nil {
				return err
			}

			ctx := appContext(cmd)
			result, err := app.WSCall(ctx, args[0], payload)
			if err != nil {
				return err
			}
			return printResult(cmd, result, "data")
		},
	}
	wsCmd.Flags().StringVar(&wsJSON, "json", "", "Websocket payload")

	raw.AddCommand(getCmd, postCmd, wsCmd)
	return raw
}
