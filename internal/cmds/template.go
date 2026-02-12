package cmds

import (
	"fmt"
	"os"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newTemplateCommand() *cobra.Command {
	var local bool

	cmd := &cobra.Command{
		Use:   "template <template> [datafile]",
		Short: "Render templates on server or locally",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			templateFile := args[0]

			templateContent, err := os.ReadFile(templateFile)
			if err != nil {
				return fmt.Errorf("failed to read template file: %w", err)
			}

			var data map[string]any
			if len(args) > 1 {
				dataContent, err := os.ReadFile(args[1])
				if err != nil {
					return fmt.Errorf("failed to read data file: %w", err)
				}
				data, err = app.ParseObject(string(dataContent))
				if err != nil {
					return fmt.Errorf("failed to parse data file: %w", err)
				}
			}

			if local {
				return fmt.Errorf("local template rendering not yet implemented")
			}

			payload := map[string]any{
				"template":  string(templateContent),
				"variables": data,
			}

			result, err := app.RESTCall(ctx, "post", "template", payload)
			if err != nil {
				return err
			}

			if str, ok := result.(string); ok {
				fmt.Fprint(ctx.Stdout, str)
			} else {
				return printResult(cmd, result, "data")
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&local, "local", false, "Render template locally instead of on server")

	return cmd
}
