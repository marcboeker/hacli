package cmds

import (
	"fmt"
	"regexp"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newAreaCommand() *cobra.Command {
	area := &cobra.Command{Use: "area", Short: "Area registry operations (EXPERIMENTAL)"}

	area.AddCommand(newAreaListCommand())
	area.AddCommand(newAreaCreateCommand())
	area.AddCommand(newAreaDeleteCommand())
	area.AddCommand(newAreaRenameCommand())

	return area
}

func newAreaListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [areafilter]",
		Short: "List all areas",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			filter := ".*"
			if len(args) > 0 {
				filter = args[0]
			}

			re, err := regexp.Compile(filter)
			if err != nil {
				return fmt.Errorf("invalid filter regex: %w", err)
			}

			ctx := appContext(cmd)
			result, err := app.WSCall(ctx, "config/area_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			areas, ok := resultMap["result"].([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			filtered := make([]any, 0, len(areas))
			for _, item := range areas {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}
				name, _ := obj["name"].(string)
				if re.MatchString(name) {
					filtered = append(filtered, obj)
				}
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ID", Path: "area_id"},
					{Header: "NAME", Path: "name"},
				}
			}
			err = printResult(cmd, filtered, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newAreaCreateCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "create <names...>",
		Short: "Create one or more areas",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			results := make([]any, 0, len(args))

			for _, name := range args {
				result, err := app.WSCall(ctx, "config/area_registry/create", map[string]any{"name": name})
				if err != nil {
					return fmt.Errorf("failed to create area %s: %w", name, err)
				}
				results = append(results, result)
			}

			return printResult(cmd, results, "data")
		},
	}
}

func newAreaDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <names...>",
		Short: "Delete areas by name or ID",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			results := make([]any, 0, len(args))

			for _, name := range args {
				area, err := findArea(ctx, name)
				if err != nil {
					return fmt.Errorf("could not find area %s: %w", name, err)
				}

				result, err := app.WSCall(ctx, "config/area_registry/delete", map[string]any{"area_id": area["area_id"]})
				if err != nil {
					return fmt.Errorf("failed to delete area %s: %w", name, err)
				}
				results = append(results, result)
			}

			return printResult(cmd, results, "data")
		},
	}
}

func newAreaRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <oldname> <newname>",
		Short: "Rename an area",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			oldname := args[0]
			newname := args[1]

			area, err := findArea(ctx, oldname)
			if err != nil {
				return fmt.Errorf("could not find area %s: %w", oldname, err)
			}

			result, err := app.WSCall(ctx, "config/area_registry/update", map[string]any{
				"area_id": area["area_id"],
				"name":    newname,
			})
			if err != nil {
				return fmt.Errorf("failed to rename area: %w", err)
			}

			return printResult(cmd, result, "data")
		},
	}
}

func findArea(ctx *app.AppContext, idOrName string) (map[string]any, error) {
	result, err := app.WSCall(ctx, "config/area_registry/list", nil)
	if err != nil {
		return nil, err
	}

	resultMap, ok := result.(map[string]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	areas, ok := resultMap["result"].([]any)
	if !ok {
		return nil, fmt.Errorf("unexpected response format")
	}

	for _, item := range areas {
		area, ok := item.(map[string]any)
		if !ok {
			continue
		}

		areaID, _ := area["area_id"].(string)
		name, _ := area["name"].(string)

		if areaID == idOrName || name == idOrName {
			return area, nil
		}
	}

	return nil, fmt.Errorf("area not found: %s", idOrName)
}
