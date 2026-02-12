package cmds

import (
	"encoding/json"
	"fmt"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
	"golang.org/x/net/websocket"
)

func newEventCommand() *cobra.Command {
	event := &cobra.Command{Use: "event", Short: "Event interaction"}

	event.AddCommand(newEventFireCommand())
	event.AddCommand(newEventWatchCommand())

	return event
}

func newEventFireCommand() *cobra.Command {
	var eventJSON string

	cmd := &cobra.Command{
		Use:   "fire <event>",
		Short: "Fire an event in Home Assistant",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)
			eventType := args[0]

			var payload map[string]any
			if eventJSON != "" {
				var err error
				payload, err = app.ParseObject(eventJSON)
				if err != nil {
					return fmt.Errorf("invalid JSON: %w", err)
				}
			}

			result, err := app.RESTCall(ctx, "post", "events/"+eventType, payload)
			if err != nil {
				return err
			}

			return printResult(cmd, result, "data")
		},
	}

	cmd.Flags().StringVar(&eventJSON, "json", "", "Raw JSON data for the event")

	return cmd
}

func newEventWatchCommand() *cobra.Command {
	return &cobra.Command{
		Use:   "watch [event_type]",
		Short: "Subscribe and print events",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)

			var eventType string
			if len(args) > 0 {
				eventType = args[0]
			}

			return watchEvents(ctx, eventType, cmd)
		},
	}
}

func watchEvents(ctx *app.AppContext, eventType string, cmd *cobra.Command) error {
	serverURL, err := app.ParseServerURL(ctx.Server)
	if err != nil {
		return err
	}

	if serverURL.Scheme == "https" {
		serverURL.Scheme = "wss"
	} else {
		serverURL.Scheme = "ws"
	}
	serverURL.Path = "/api/websocket"

	origin := "http://localhost/"
	conn, err := websocket.Dial(serverURL.String(), "", origin)
	if err != nil {
		return fmt.Errorf("dial websocket: %w", err)
	}
	defer conn.Close()

	var greeting map[string]any
	if err := websocket.JSON.Receive(conn, &greeting); err != nil {
		return fmt.Errorf("receive ws greeting: %w", err)
	}

	auth := map[string]any{"type": "auth"}
	if ctx.Token != "" {
		auth["access_token"] = ctx.Token
	}
	if err := websocket.JSON.Send(conn, auth); err != nil {
		return fmt.Errorf("send ws auth: %w", err)
	}

	var authResp map[string]any
	if err := websocket.JSON.Receive(conn, &authResp); err != nil {
		return fmt.Errorf("receive ws auth response: %w", err)
	}
	if authResp["type"] != "auth_ok" {
		return fmt.Errorf("ws auth failed")
	}

	frame := map[string]any{
		"id":   1,
		"type": "subscribe_events",
	}
	if eventType != "" {
		frame["event_type"] = eventType
	}

	if err := websocket.JSON.Send(conn, frame); err != nil {
		return fmt.Errorf("send subscribe request: %w", err)
	}

	var result map[string]any
	if err := websocket.JSON.Receive(conn, &result); err != nil {
		return fmt.Errorf("receive subscribe response: %w", err)
	}

	if result["success"] != true {
		return fmt.Errorf("subscribe failed: %v", result)
	}

	fmt.Fprintf(ctx.Stdout, "Watching for events...\n")

	for {
		var msg map[string]any
		if err := websocket.JSON.Receive(conn, &msg); err != nil {
			return fmt.Errorf("receive event: %w", err)
		}

		if msg["type"] == "event" {
			event := msg["event"].(map[string]any)
			data, _ := json.MarshalIndent(event, "", "  ")
			fmt.Fprintf(ctx.Stdout, "%s\n", string(data))
		} else if msg["type"] == "auth_invalid" {
			return fmt.Errorf("auth invalid: %s", msg["message"])
		}
	}
}
