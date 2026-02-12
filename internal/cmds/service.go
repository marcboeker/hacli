package cmds

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newServiceCommand() *cobra.Command {
	service := &cobra.Command{Use: "service", Short: "Service commands"}

	listCmd := &cobra.Command{
		Use:   "list [servicefilter]",
		Short: "List all services",
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
			data, err := app.RESTCall(ctx, "get", "services", nil)
			if err != nil {
				return err
			}

			items, ok := data.([]any)
			if !ok {
				return printResult(cmd, data, "table")
			}

			services := make([]any, 0)
			for _, domainItem := range items {
				domainObj, ok := domainItem.(map[string]any)
				if !ok {
					continue
				}
				domain, _ := domainObj["domain"].(string)
				body, _ := domainObj["services"].(map[string]any)
				for serviceName := range body {
					fullName := strings.Trim(domain+"."+serviceName, ".")
					if re.MatchString(fullName) {
						services = append(services, map[string]any{"service": fullName})
					}
				}
			}

			return printResult(cmd, services, "table")
		},
	}

	var arguments string
	callCmd := &cobra.Command{
		Use:   "call <service>",
		Short: "Call service",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			parts := strings.SplitN(args[0], ".", 2)
			if len(parts) != 2 {
				return fmt.Errorf("service must be in domain.service format")
			}

			payload, err := app.ParseObject(arguments)
			if err != nil {
				return err
			}

			ctx := appContext(cmd)
			result, err := app.RESTCall(ctx, "post", "services/"+parts[0]+"/"+parts[1], payload)
			if err != nil {
				return err
			}

			return printResult(cmd, result, "data")
		},
	}
	callCmd.Flags().StringVarP(&arguments, "arguments", "a", "", "Service call arguments as JSON or key=value pairs")

	service.AddCommand(listCmd, callCmd)
	return service
}
