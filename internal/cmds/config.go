package cmds

import (
	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newConfigCommand() *cobra.Command {
	config := &cobra.Command{Use: "config", Short: "Get configuration from Home Assistant"}

	config.AddCommand(newConfigFullCommand())
	config.AddCommand(newConfigComponentsCommand())
	config.AddCommand(newConfigWhitelistDirsCommand())
	config.AddCommand(newConfigReleaseCommand())

	return config
}

func newConfigFullCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "full",
		Short: "Get full configuration details",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", "config", nil)
			if err != nil {
				return err
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "VERSION", Path: "version"},
					{Header: "CONFIG", Path: "config_dir"},
					{Header: "TZ", Path: "time_zone"},
					{Header: "LOCATION", Path: "location_name"},
					{Header: "LONGITUDE", Path: "longitude"},
					{Header: "LATITUDE", Path: "latitude"},
					{Header: "ELEVATION", Path: "elevation"},
					{Header: "UNITS", Path: "unit_system"},
				}
			}
			err = printResult(cmd, result, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newConfigComponentsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "components",
		Short: "Get loaded components",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", "config", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			components, ok := resultMap["components"].([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			formatted := make([]any, 0, len(components))
			for _, comp := range components {
				formatted = append(formatted, map[string]any{"component": comp})
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "COMPONENT", Path: "component"},
				}
			}
			err = printResult(cmd, formatted, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newConfigWhitelistDirsCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "whitelist_dirs",
		Short: "Get whitelisted directories",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", "config", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			dirs, ok := resultMap["whitelist_external_dirs"].([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			formatted := make([]any, 0, len(dirs))
			for _, dir := range dirs {
				formatted = append(formatted, map[string]any{"directory": dir})
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "DIRECTORY", Path: "directory"},
				}
			}
			err = printResult(cmd, formatted, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newConfigReleaseCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "release",
		Short: "Get Home Assistant release version",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", "config", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			version := resultMap["version"]
			formatted := []any{map[string]any{"version": version}}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "VERSION", Path: "version"},
				}
			}
			err = printResult(cmd, formatted, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}
