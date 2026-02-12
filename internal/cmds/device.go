package cmds

import (
	"fmt"
	"regexp"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newDeviceCommand() *cobra.Command {
	device := &cobra.Command{Use: "device", Short: "Device registry operations (EXPERIMENTAL)"}

	device.AddCommand(newDeviceListCommand())
	device.AddCommand(newDeviceAssignCommand())
	device.AddCommand(newDeviceRenameCommand())

	return device
}

func newDeviceListCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "list [devicefilter]",
		Short: "List all devices",
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

			result, err := app.WSCall(ctx, "config/device_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := result.(map[string]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			devices, ok := resultMap["result"].([]any)
			if !ok {
				return printResult(cmd, result, "table")
			}

			filtered := make([]any, 0, len(devices))
			for _, item := range devices {
				obj, ok := item.(map[string]any)
				if !ok {
					continue
				}

				name, _ := obj["name"].(string)
				if areaID, ok := obj["area_id"].(string); ok && areaID != "" {
					obj["area_name"] = areas[areaID]
				}

				if re.MatchString(name) {
					filtered = append(filtered, obj)
				}
			}

			originalDefs := ctx.ColumnDefs
			if len(ctx.ColumnDefs) == 0 {
				ctx.ColumnDefs = []app.ColumnDef{
					{Header: "ID", Path: "id"},
					{Header: "NAME", Path: "name"},
					{Header: "MODEL", Path: "model"},
					{Header: "MANUFACTURER", Path: "manufacturer"},
					{Header: "AREA", Path: "area_name"},
				}
			}
			err = printResult(cmd, filtered, "table")
			ctx.ColumnDefs = originalDefs
			return err
		},
	}
}

func newDeviceAssignCommand() *cobra.Command {
	var match string

	cmd := &cobra.Command{
		Use:   "assign <area_id_or_name> [names...]",
		Short: "Assign devices to an area",
		Args:  cobra.MinimumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			areaIDOrName := args[0]
			names := args[1:]

			area, err := findArea(ctx, areaIDOrName)
			if err != nil {
				return fmt.Errorf("could not find area %s: %w", areaIDOrName, err)
			}

			devicesResult, err := app.WSCall(ctx, "config/device_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := devicesResult.(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			devices, ok := resultMap["result"].([]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			devicesToAssign := make([]map[string]any, 0)

			if match != "" {
				re, err := regexp.Compile(match)
				if err != nil {
					return fmt.Errorf("invalid match regex: %w", err)
				}

				for _, item := range devices {
					device, ok := item.(map[string]any)
					if !ok {
						continue
					}
					name, _ := device["name"].(string)
					if re.MatchString(name) {
						devicesToAssign = append(devicesToAssign, device)
					}
				}
			}

			for _, idOrName := range names {
				for _, item := range devices {
					device, ok := item.(map[string]any)
					if !ok {
						continue
					}

					id, _ := device["id"].(string)
					name, _ := device["name"].(string)

					if id == idOrName || name == idOrName {
						devicesToAssign = append(devicesToAssign, device)
						break
					}
				}
			}

			results := make([]any, 0, len(devicesToAssign))
			for _, device := range devicesToAssign {
				deviceID := device["id"].(string)
				result, err := app.WSCall(ctx, "config/device_registry/update", map[string]any{
					"device_id": deviceID,
					"area_id":   area["area_id"],
				})
				if err != nil {
					return fmt.Errorf("failed to assign device: %w", err)
				}
				results = append(results, result)
			}

			return printResult(cmd, results, "data")
		},
	}

	cmd.Flags().StringVar(&match, "match", "", "Expression used to find devices matching that name")

	return cmd
}

func newDeviceRenameCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "rename <device_id_or_name> <new_name>",
		Short: "Rename a device",
		Args:  cobra.ExactArgs(2),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			idOrName := args[0]
			newName := args[1]

			devicesResult, err := app.WSCall(ctx, "config/device_registry/list", nil)
			if err != nil {
				return err
			}

			resultMap, ok := devicesResult.(map[string]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			devices, ok := resultMap["result"].([]any)
			if !ok {
				return fmt.Errorf("unexpected response format")
			}

			var device map[string]any
			for _, item := range devices {
				d, ok := item.(map[string]any)
				if !ok {
					continue
				}

				id, _ := d["id"].(string)
				name, _ := d["name"].(string)

				if id == idOrName || name == idOrName {
					device = d
					break
				}
			}

			if device == nil {
				return fmt.Errorf("could not find device: %s", idOrName)
			}

			result, err := app.WSCall(ctx, "config/device_registry/update", map[string]any{
				"device_id":    device["id"],
				"name_by_user": newName,
			})
			if err != nil {
				return fmt.Errorf("failed to rename device: %w", err)
			}

			return printResult(cmd, result, "data")
		},
	}
}
