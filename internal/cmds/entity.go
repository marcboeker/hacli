package cmds

import (
	"fmt"
	"regexp"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newEntityCommand() *cobra.Command {
	entity := &cobra.Command{Use: "entity", Short: "Entity operations"}

	entity.AddCommand(newEntityListCommand())
	entity.AddCommand(newEntityAssignCommand())
	entity.AddCommand(newEntityRenameCommand())

	return entity
}

func newEntityListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [entityfilter]",
		Short: "List all entities",
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

			areasResult, err := app.WSCall(ctx, "config/area_registry/list", nil)
			if err != nil {
				return err
			}

			areas := make(map[string]string)
			if resultMap, ok := areasResult.(map[string]any); ok {
				if areaList, ok := resultMap["result"].([]any); ok {
					for _, item := range areaList {
						if area, ok := item.(map[string]any); ok {
							areaID, _ := area["area_id"].(string)
							name, _ := area["name"].(string)
							areas[areaID] = name
						}
					}
				}
			}

			result, err := app.WSCall(ctx, "config/entity_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			entities, ok := resultMap["result"].([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			filtered := make([]any, 0, len(entities))
			for _, item := range entities {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}

				entityID, _ := obj["entity_id"].(string)
				if areaID, ok := obj["area_id"].(string); ok && areaID != "" {
					obj["area_name"] = areas[areaID]
				}

				if re.MatchString(entityID) {
					filtered = append(filtered, obj)
				}
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ENTITY_ID", Path: "entity_id"},
					{Header: "NAME", Path: "name"},
					{Header: "DEVICE_ID", Path: "device_id"},
					{Header: "PLATFORM", Path: "platform"},
					{Header: "AREA", Path: "area_name"},
					{Header: "CONFIG_ENTRY_ID", Path: "config_entry_id"},
					{Header: "DISABLED_BY", Path: "disabled_by"},
				}
			}
			err = printResult(cmd, filtered, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newEntityAssignCommand() *cobra.Command {
	var match string

	cmd := &cobra.Command{
		Use:   "assign <area_id_or_name> [names...]",
		Short: "Assign entities to an area",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			areaIDOrName := args[0]
			names := args[1:]

			area, err := findArea(ctx, areaIDOrName)
			if err != nil {
				return fmt.Errorf("could not find area %s: %w", areaIDOrName, err)
			}

			entitiesResult, err := app.WSCall(ctx, "config/entity_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := entitiesResult.(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			entities, ok := resultMap["result"].([]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			entitiesToAssign := make([]map[string]any, 0)

			if match != "" {
				re, err := regexp.Compile(match)
				if err != nil {
					return fmt.Errorf("invalid match regex: %w", err)
				}

				for _, item := range entities {
					entity, ok := item.(map[string]any)
					if !ok {
						continue
					}
					name, _ := entity["name"].(string)
					if re.MatchString(name) {
						entitiesToAssign = append(entitiesToAssign, entity)
					}
				}
			}

			for _, idOrName := range names {
				for _, item := range entities {
					entity, ok := item.(map[string]any)
					if !ok {
						continue
					}

					entityID, _ := entity["entity_id"].(string)
					name, _ := entity["name"].(string)

					if entityID == idOrName || name == idOrName {
						entitiesToAssign = append(entitiesToAssign, entity)
						break
					}
				}
			}

			results := make([]any, 0, len(entitiesToAssign))
			for _, entity := range entitiesToAssign {
				entityID := entity["entity_id"].(string)
				result, err := app.WSCall(ctx, "config/entity_registry/update", map[string]any{
					"entity_id": entityID,
					"area_id":   area["area_id"],
				})
				if err != nil {
					return fmt.Errorf("failed to assign entity: %w", err)
				}
				results = append(results, result)
			}

			return printResult(cmd, results, "data")
		},
	}

	cmd.Flags().StringVar(&match, "match", "", "Expression used to find entities matching that name")

	return cmd
}

func newEntityRenameCommand() *cobra.Command {
	var newName string

	cmd := &cobra.Command{
		Use:   "rename <oldid> [newid]",
		Short: "Rename an entity",
		Args:  cobra.RangeArgs(1, 2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			oldID := args[0]

			var newID string
			if len(args) > 1 {
				newID = args[1]
			}

			if newID == "" && newName == "" {
				return fmt.Errorf("need to at least specify either a new id or new name")
			}

			payload := map[string]any{
				"entity_id": oldID,
			}

			if newID != "" {
				payload["new_entity_id"] = newID
			}
			if newName != "" {
				payload["name"] = newName
			}

			result, err := app.WSCall(ctx, "config/entity_registry/update", payload)
			if err != nil {
				return fmt.Errorf("failed to rename entity: %w", err)
			}

			return printResult(cmd, result, "data")
		},
	}

	cmd.Flags().StringVar(&newName, "name", "", "New name for the entity")

	return cmd
}
