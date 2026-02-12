package cmds

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

func newHaCommand() *cobra.Command {
	ha := &cobra.Command{Use: "ha", Short: "Home Assistant (former Hass.io) commands"}

	ha.AddCommand(newHaSupervisorCommand())
	ha.AddCommand(newHaSnapshotCommand())
	ha.AddCommand(newHaHostCommand())
	ha.AddCommand(newHaOSCommand())
	ha.AddCommand(newHaHardwareCommand())
	ha.AddCommand(newHaAddonsCommand())
	ha.AddCommand(newHaCoreCommand())
	ha.AddCommand(newHaDNSCommand())
	ha.AddCommand(newHaMulticastCommand())
	ha.AddCommand(newHaCLICommand())

	return ha
}

func hassioAPICall(ctx *app.AppContext, method string, httpMethod string) (any, error) {
	apiURL, err := app.BuildAPIURL(ctx.Server, "hassio/"+method)
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest(httpMethod, apiURL, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	for name, value := range ctx.AuthHeaders() {
		req.Header.Set(name, value)
	}
	req.Header.Set("content-type", "application/json")

	resp, err := ctx.HTTPClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	var result any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return result, nil
}

func newHaSupervisorCommand() *cobra.Command {
	supervisor := &cobra.Command{Use: "supervisor", Short: "Home Assistant supervisor commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"ping", "supervisor/ping", "GET"},
		{"info", "supervisor/info", "GET"},
		{"update", "supervisor/update", "POST"},
		{"options", "supervisor/options", "POST"},
		{"reload", "supervisor/reload", "POST"},
		{"logs", "supervisor/logs", "GET"},
		{"repair", "supervisor/repair", "POST"},
		{"stats", "supervisor/stats", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		supervisor.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " supervisor",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return supervisor
}

func newHaSnapshotCommand() *cobra.Command {
	snapshot := &cobra.Command{Use: "snapshot", Short: "Home Assistant snapshot commands"}

	snapshot.AddCommand(&cobra.Command{
		Use:   "reload",
		Short: "Reload snapshots",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := hassioAPICall(ctx, "snapshots/reload", "POST")
			if err != nil {
				return err
			}
			return printResult(cmd, result, "data")
		},
	})

	snapshot.AddCommand(&cobra.Command{
		Use:   "shutdown",
		Short: "Shutdown host",
		RunE: func(cmd *cobra.Command, _ []string) error {
			ctx := appContext(cmd)
			result, err := hassioAPICall(ctx, "host/shutdown", "POST")
			if err != nil {
				return err
			}
			return printResult(cmd, result, "data")
		},
	})

	return snapshot
}

func newHaHostCommand() *cobra.Command {
	host := &cobra.Command{Use: "host", Short: "Home Assistant host commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"reboot", "host/reboot", "POST"},
		{"reload", "host/reload", "POST"},
		{"shutdown", "host/shutdown", "POST"},
		{"info", "host/info", "GET"},
		{"options", "host/options", "POST"},
		{"services", "host/services", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		host.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " host",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return host
}

func newHaOSCommand() *cobra.Command {
	os := &cobra.Command{Use: "os", Short: "Home Assistant OS commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "os/info", "GET"},
		{"update", "os/update", "POST"},
	}

	for _, cmd := range commands {
		cmd := cmd
		os.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " OS",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return os
}

func newHaHardwareCommand() *cobra.Command {
	hardware := &cobra.Command{Use: "hardware", Short: "Home Assistant hardware info"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "hardware/info", "GET"},
		{"audio", "hardware/audio", "GET"},
		{"trigger", "hardware/trigger", "POST"},
	}

	for _, cmd := range commands {
		cmd := cmd
		hardware.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " hardware",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return hardware
}

func newHaAddonsCommand() *cobra.Command {
	addons := &cobra.Command{Use: "addons", Short: "Home Assistant addons commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"all", "addons", "GET"},
		{"reload", "addons/reload", "POST"},
	}

	for _, cmd := range commands {
		cmd := cmd
		addons.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " addons",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return addons
}

func newHaCoreCommand() *cobra.Command {
	core := &cobra.Command{Use: "core", Short: "Home Assistant core commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "core/info", "GET"},
		{"update", "core/update", "POST"},
		{"logs", "core/logs", "GET"},
		{"restart", "core/restart", "POST"},
		{"check", "core/check", "POST"},
		{"start", "core/start", "POST"},
		{"stop", "core/stop", "POST"},
		{"rebuild", "core/rebuild", "POST"},
		{"options", "core/options", "POST"},
		{"websocket", "core/websocket", "GET"},
		{"stats", "core/stats", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		core.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " core",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return core
}

func newHaDNSCommand() *cobra.Command {
	dns := &cobra.Command{Use: "dns", Short: "Home Assistant DNS commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "dns/info", "GET"},
		{"options", "dns/options", "POST"},
		{"restart", "dns/restart", "POST"},
		{"logs", "dns/logs", "GET"},
		{"stats", "dns/stats", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		dns.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " DNS",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return dns
}

func newHaMulticastCommand() *cobra.Command {
	multicast := &cobra.Command{Use: "multicast", Short: "Home Assistant Multicast commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "multicast/info", "GET"},
		{"update", "multicast/update", "POST"},
		{"restart", "multicast/restart", "POST"},
		{"logs", "multicast/logs", "GET"},
		{"stats", "multicast/stats", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		multicast.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " multicast",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return multicast
}

func newHaCLICommand() *cobra.Command {
	cli := &cobra.Command{Use: "ha-cli", Short: "Home Assistant ha-cli commands"}

	commands := []struct {
		name   string
		method string
		verb   string
	}{
		{"info", "cli/info", "GET"},
		{"update", "cli/update", "POST"},
		{"restart", "cli/restart", "POST"},
		{"logs", "cli/logs", "GET"},
		{"stats", "cli/stats", "GET"},
	}

	for _, cmd := range commands {
		cmd := cmd
		cli.AddCommand(&cobra.Command{
			Use:   cmd.name,
			Short: cmd.name + " ha-cli",
			RunE: func(c *cobra.Command, _ []string) error {
				ctx := appContext(c)
				result, err := hassioAPICall(ctx, cmd.method, cmd.verb)
				if err != nil {
					return err
				}
				return printResult(c, result, "data")
			},
		})
	}

	return cli
}
