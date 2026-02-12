package cmds

import (
	"fmt"
	"net/url"
	"os/exec"
	"runtime"

	"github.com/marcboeker/hacli/internal/app"
	"github.com/spf13/cobra"
)

var mapServices = map[string]string{
	"openstreetmap": "https://www.openstreetmap.org/?mlat={lat}&mlon={lon}#map=17/{lat}/{lon}",
	"google":        "https://www.google.com/maps/search/?api=1&query={lat},{lon}",
	"bing":          "https://www.bing.com/maps?v=2&cp={lat}~{lon}&lvl=17&sp=point.{lat}_{lon}_{name}",
}

func newMapCommand() *cobra.Command {
	var service string

	cmd := &cobra.Command{
		Use:   "map [entity]",
		Short: "Show location on a map",
		Args:  cobra.MaximumNArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			ctx := appContext(cmd)

			var latitude, longitude float64
			var thing string

			if len(args) > 0 {
				entity := args[0]
				thing = entity

				result, err := app.RESTCall(ctx, "get", "states/"+entity, nil)
				if err != nil {
					return err
				}

				if state, ok := result.(map[string]any); ok {
					if attrs, ok := state["attributes"].(map[string]any); ok {
						if lat, ok := attrs["latitude"].(float64); ok {
							latitude = lat
						}
						if lon, ok := attrs["longitude"].(float64); ok {
							longitude = lon
						}
						if friendlyName, ok := attrs["friendly_name"].(string); ok {
							thing = friendlyName
						}
					}
				}
			} else {
				thing = "configuration"
				result, err := app.RESTCall(ctx, "get", "config", nil)
				if err != nil {
					return err
				}

				if config, ok := result.(map[string]any); ok {
					if lat, ok := config["latitude"].(float64); ok {
						latitude = lat
					}
					if lon, ok := config["longitude"].(float64); ok {
						longitude = lon
					}
					if locationName, ok := config["location_name"].(string); ok {
						thing = locationName
					}
				}
			}

			if latitude == 0 && longitude == 0 {
				return fmt.Errorf("no exact location info found in %s", thing)
			}

			urlPattern, ok := mapServices[service]
			if !ok {
				return fmt.Errorf("unknown service: %s", service)
			}

			mapURL := urlPattern
			mapURL = replaceAll(mapURL, "{lat}", fmt.Sprintf("%f", latitude))
			mapURL = replaceAll(mapURL, "{lon}", fmt.Sprintf("%f", longitude))
			mapURL = replaceAll(mapURL, "{name}", url.QueryEscape(thing))

			fmt.Fprintf(ctx.Stdout, "%s location is at %f, %f\n", thing, latitude, longitude)

			return openBrowser(mapURL)
		},
	}

	cmd.Flags().StringVar(&service, "service", "openstreetmap", "Map service to use (openstreetmap, google, bing)")

	return cmd
}

func replaceAll(s, old, new string) string {
	for {
		newStr := ""
		for i := 0; i < len(s); {
			if i+len(old) <= len(s) && s[i:i+len(old)] == old {
				newStr += new
				i += len(old)
			} else {
				newStr += string(s[i])
				i++
			}
		}
		if newStr == s {
			return s
		}
		s = newStr
	}
}

func openBrowser(url string) error {
	var cmd string
	var args []string

	switch runtime.GOOS {
	case "windows":
		cmd = "cmd"
		args = []string{"/c", "start"}
	case "darwin":
		cmd = "open"
	default:
		cmd = "xdg-open"
	}
	args = append(args, url)
	return exec.Command(cmd, args...).Start()
}
