package cmds

import (
	"errors"
	"fmt"
	"os"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func NewRootCommand() *cobra.Command {
	ctx := app.New()

	root := &cobra.Command{
		Use:           "hacli",
		Short:         "Command-line tool for Home Assistant",
		SilenceErrors: true,
		SilenceUsage:  true,
		PersistentPreRunE: func(cmd *cobra.Command, _ []string) error {
			if err := ctx.Resolve(); err != nil {
				if ctx.Server == "auto" || ctx.Server == "" {
					return errors.New("unable to locate Home Assistant (set --server or HA_SERVER)")
				}
				return err
			}
			cmd.SetContext(app.WithContext(cmd.Context(), ctx))
			return nil
		},
	}

	root.PersistentFlags().StringVarP(&ctx.Server, "server", "s", ctx.Server, "Server for Home Assistant")
	root.PersistentFlags().StringVar(&ctx.Token, "token", ctx.Token, "Access token for Home Assistant")
	root.PersistentFlags().StringVar(&ctx.Password, "password", ctx.Password, "Password for Home Assistant")
	root.PersistentFlags().IntVar(&ctx.Timeout, "timeout", ctx.Timeout, "Timeout for all operations")
	root.PersistentFlags().StringVarP(&ctx.Output, "output", "o", ctx.Output, "Output format: json/yaml/table/auto/ndjson")
	root.PersistentFlags().CountVarP(&ctx.Verbose, "verbose", "v", "Enable verbose mode")
	root.PersistentFlags().BoolVarP(&ctx.Insecure, "insecure", "x", false, "Ignore certificate warnings")
	root.PersistentFlags().StringVar(&ctx.Cert, "cert", os.Getenv("HA_CERT"), "Certificate for validating SSL")
	root.PersistentFlags().BoolVar(&ctx.Debug, "debug", false, "Debug mode")
	root.PersistentFlags().StringVar(&ctx.ColumnsRaw, "columns", "", "Comma separated list of columns")
	root.PersistentFlags().BoolVar(&ctx.NoHeaders, "no-headers", false, "Disable table headers")
	root.PersistentFlags().StringVar(&ctx.TableFormat, "table-format", "markdown", "Table format (markdown, plain)")
	root.PersistentFlags().StringVar(&ctx.SortBy, "sort-by", "", "Sort table by path")

	root.AddCommand(newServiceCommand())
	root.AddCommand(newStateCommand())
	root.AddCommand(newRawCommand())
	root.AddCommand(newCompletionCommand())
	root.AddCommand(newInfoCommand())
	root.AddCommand(newAreaCommand())
	root.AddCommand(newConfigCommand())
	root.AddCommand(newDeviceCommand())
	root.AddCommand(newEntityCommand())
	root.AddCommand(newEventCommand())
	root.AddCommand(newHaCommand())
	root.AddCommand(newMapCommand())
	root.AddCommand(newSystemCommand())
	root.AddCommand(newTemplateCommand())

	return root
}

func appContext(cmd *cobra.Command) *app.AppContext {
	return app.FromContext(cmd.Context())
}

func printResult(cmd *cobra.Command, value any, defaultOutput string) error {
	ctx := appContext(cmd)
	if err := app.PrintData(ctx, value, defaultOutput); err != nil {
		return fmt.Errorf("render output: %w", err)
	}
	return nil
}
