package cmds

import (
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newStateCommand() *cobra.Command {
	state := &cobra.Command{Use: "state", Short: "Entity state commands"}

	state.AddCommand(newStateGetCommand())
	state.AddCommand(newStateDeleteCommand())
	state.AddCommand(newStateListCommand())
	state.AddCommand(newStateEditCommand())
	state.AddCommand(newStateToggleCommand("toggle"))
	state.AddCommand(newStateToggleCommand("turn_on"))
	state.AddCommand(newStateToggleCommand("turn_off"))
	state.AddCommand(newStateHistoryCommand())

	return state
}

func newStateGetCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "get <entity>",
		Short: "Get state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "get", "states/"+args[0], nil)
			if err != nil {
				return err
			}
			return printResult(cmd, result, "table")
		},
	}
}

func newStateDeleteCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "delete <entity>",
		Short: "Delete state",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "delete", "states/"+args[0], nil)
			if err != nil {
				return err
			}
			if result != nil {
				fmt.Fprintf(ctx.Stdout, "State for entity %s deleted.\n", args[0])
			} else {
				fmt.Fprintf(ctx.Stdout, "Entity %s not found.\n", args[0])
			}
			return nil
		},
	}
}

func newStateListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [entityfilter]",
		Short: "List states",
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
			result, err := app.RESTCall(ctx, "get", "states", nil)
			if err != nil {
				return err
			}

			items, ok := result.([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			filtered := make([]any, 0, len(items))
			for _, item := range items {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}
				entityID, _ := obj["entity_id"].(string)
				if re.MatchString(entityID) {
					filtered = append(filtered, obj)
				}
			}

			// Set default columns to match Python COLUMNS_ENTITIES
			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ENTITY", Path: "entity_id"},
					{Header: "DESCRIPTION", Path: "attributes.friendly_name"},
					{Header: "STATE", Path: "state"},
					{Header: "CHANGED", Path: "last_changed"},
				}
			}
			err = printResult(cmd, filtered, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newStateEditCommand() *cobra.Command {
	var attributes string
	var attributesJSON string
	var merge bool

	cmd := &cobra.Command{
		Use:   "edit <entity> [newstate]",
		Short: "Edit state",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			entity := args[0]

			payload := map[string]any{}
			if len(args) == 2 {
				payload["state"] = args[1]
			}

			attrs := map[string]any{}
			if attributes != "" {
				parsed, err := app.ParseObject(attributes)
				if err != nil {
					return err
				}
				for key, value := range parsed {
					attrs[key] = value
				}
			}

			if attributesJSON != "" {
				parsed, err := app.ParseObject(attributesJSON)
				if err != nil {
					return err
				}
				for key, value := range parsed {
					attrs[key] = value
				}
			}

			if len(attrs) > 0 {
				if merge {
					originalAny, err := app.RESTCall(ctx, "get", "states/"+entity, nil)
					if err == nil {
						if original, ok := originalAny.(map[string]any); ok {
							if currentAttrs, ok := original["attributes"].(map[string]any); ok {
								for key, value := range currentAttrs {
									if _, exists := attrs[key]; !exists {
										attrs[key] = value
									}
								}
							}
						}
					}
				}
				payload["attributes"] = attrs
			}

			result, err := app.RESTCall(ctx, "post", "states/"+entity, payload)
			if err != nil {
				return err
			}

			return printResult(cmd, result, "data")
		},
	}

	cmd.Flags().StringVarP(&attributes, "attributes", "a", "", "Attributes as key=value list")
	cmd.Flags().StringVar(&attributesJSON, "json", "", "Attributes as JSON map")
	cmd.Flags().BoolVar(&merge, "merge", false, "Merge with current attributes")

	return cmd
}

func newStateToggleCommand(action string) *cobra.Command {
	return &cobra.Command{
		Use:   action + " <entities...>",
		Short: strings.ReplaceAll(action, "_", " "),
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			result := make([]any, 0, len(args))
			for _, entity := range args {
				response, err := app.RESTCall(ctx, "post", "services/homeassistant/"+action, map[string]any{"entity_id": entity})
				if err != nil {
					return err
				}
				result = append(result, response)
			}
			// Flatten single entity result and use table output with default columns
			if len(args) == 1 && len(result) > 0 {
				if arr, ok := result[0].([]any); ok && len(arr) > 0 {
					result = arr
				}
			}
			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ENTITY", Path: "entity_id"},
					{Header: "DESCRIPTION", Path: "attributes.friendly_name"},
					{Header: "STATE", Path: "state"},
					{Header: "CHANGED", Path: "last_changed"},
				}
			}
			err := printResult(cmd, result, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newStateHistoryCommand() *cobra.Command {
	var since string
	var end string

	cmd := &cobra.Command{
		Use:   "history <entities...>",
		Short: "Show entity state history",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)

			sinceTime, err := app.ParseTimeOrDuration(since)
			if err != nil {
				return fmt.Errorf("invalid since time: %w", err)
			}

			endTime, err := app.ParseTimeOrDuration(end)
			if err != nil {
				return fmt.Errorf("invalid end time: %w", err)
			}

			sinceStr := sinceTime.Format(time.RFC3339)
			endStr := endTime.Format(time.RFC3339)

			query := fmt.Sprintf("history/period/%s?filter_entity_id=%s&end_time=%s", sinceStr, strings.Join(args, ","), endStr)
			result, err := app.RESTCall(ctx, "get", query, nil)
			if err != nil {
				return err
			}
			// Flatten history response and use table output with default columns
			if arr, ok := result.([]any); ok && len(arr) > 0 {
				flat := make([]any, 0)
				for _, item := range arr {
					if inner, ok := item.([]any); ok {
						flat = append(flat, inner...)
					} else {
						flat = append(flat, item)
					}
				}
				result = flat
			}
			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ENTITY", Path: "entity_id"},
					{Header: "DESCRIPTION", Path: "attributes.friendly_name"},
					{Header: "STATE", Path: "state"},
					{Header: "CHANGED", Path: "last_changed"},
				}
			}
			err = printResult(cmd, result, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}

	cmd.Flags().StringVar(&since, "since", "1d", "History since")
	cmd.Flags().StringVar(&end, "end", "now", "History end")

	return cmd
}
