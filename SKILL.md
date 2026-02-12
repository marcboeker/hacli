# Home Assistant CLI (hacli)

A command-line interface for Home Assistant written in Go.

## Overview

This CLI provides comprehensive management of Home Assistant instances via REST and WebSocket APIs. It supports entity state management, service calls, device/area registry operations, event monitoring, and Home Assistant OS/Supervisor management.

## Authentication

### Environment Variables

```bash
export HA_SERVER=http://homeassistant.local:8123/
export HA_TOKEN=your_long_lived_access_token
# OR
export HAIO_TOKEN=your_token  # For Home Assistant OS
```

### Command-Line Flags

```bash
hacli --server http://homeassistant.local:8123/ --token your_token <command>
```

### Legacy Password Auth

```bash
export HA_PASSWORD=your_password  # Not recommended
```

## Global Flags

| Flag             | Short | Description                                      |
| ---------------- | ----- | ------------------------------------------------ |
| `--server`       | `-s`  | Home Assistant server URL                        |
| `--token`        |       | Access token                                     |
| `--password`     |       | Password (legacy)                                |
| `--timeout`      |       | Request timeout in seconds (default: 5)          |
| `--output`       | `-o`  | Output format: json/yaml/table/auto/ndjson       |
| `--verbose`      | `-v`  | Enable verbose output (repeatable)               |
| `--insecure`     | `-x`  | Ignore SSL certificate warnings                  |
| `--cert`         |       | Custom certificate for SSL validation            |
| `--debug`        |       | Enable debug mode                                |
| `--columns`      |       | Comma-separated list of columns for table output |
| `--no-headers`   |       | Disable table headers                            |
| `--table-format` |       | Table format (default: plain)                    |
| `--sort-by`      |       | Sort table by JSON path                          |

## Commands

### Info

Display Home Assistant server information.

```bash
hacli info
hacli info -o yaml
```

### State Management

**List all entity states:**

```bash
hacli state list
hacli state list -o json
```

**Filter entities with regex:**

```bash
hacli state list "light.*"
hacli state list "sensor.temperature.*"
```

**Get specific entity state:**

```bash
hacli state get light.living_room
hacli state get sensor.temperature -o yaml
```

**Edit entity state:**

```bash
hacli state edit light.living_room on
hacli state edit sensor.custom_value 42
```

**Edit state with attributes:**

```bash
hacli state edit light.living_room on -a brightness=255,color_temp=300
hacli state edit entity.test "Hello" --json '{"icon": "mdi:home"}'
```

**Merge with existing attributes:**

```bash
hacli state edit light.living_room on -a brightness=128 --merge
```

**Toggle entities:**

```bash
hacli state toggle light.living_room
hacli state turn_on light.living_room light.kitchen
hacli state turn_off light.living_room
```

**Delete entity state:**

```bash
hacli state delete sensor.temporary
```

**View entity history:**

```bash
hacli state history light.living_room
hacli state history sensor.temperature --since 1h
hacli state history sensor.temperature --since 24h --end 12h
```

Time formats for `--since` and `--end`:

- `now` - Current time
- `1h`, `24h` - Hours ago
- `1d`, `7d` - Days ago
- `30m` - Minutes ago
- `2024-01-15T10:30:00` - ISO timestamp
- `2024-01-15` - Date only

### Services

**List all services:**

```bash
hacli service list
hacli service list "light.*"
```

**Call a service:**

```bash
hacli service call light.turn_on
hacli service call light.turn_on -a entity_id=light.living_room
hacli service call light.turn_on -a '{"entity_id": "light.living_room", "brightness": 255}'
```

**Call service with multiple entities:**

```bash
hacli service call homeassistant.turn_on -a entity_id=light.living_room,light.kitchen
```

### Entity Registry

**List all entities:**

```bash
hacli entity list
hacli entity list "sensor.*"
```

**Assign entity to area:**

```bash
hacli entity assign living_room light.living_room
hacli entity assign "Living Room" sensor.temperature
hacli entity assign kitchen --match ".*light.*"
```

**Rename entity:**

```bash
hacli entity rename light.old_name light.new_name
hacli entity rename light.living_room --name "Living Room Light"
```

### Device Registry

**List all devices:**

```bash
hacli device list
hacli device list "Philips.*"
```

**Assign device to area:**

```bash
hacli device assign bedroom device_id_or_name
hacli device assign kitchen --match ".*sensor.*"
```

**Rename device:**

```bash
hacli device rename "Old Device Name" "New Device Name"
hacli device rename device_id "New Name"
```

### Area Registry

**List all areas:**

```bash
hacli area list
hacli area list "Bed.*"
```

**Create areas:**

```bash
hacli area create Kitchen
hacli area create Bedroom "Living Room" Office
```

**Delete areas:**

```bash
hacli area delete Kitchen
hacli area delete bedroom "Living Room"
```

**Rename area:**

```bash
hacli area rename "Old Name" "New Name"
```

### Events

**Fire an event:**

```bash
hacli event fire my_custom_event
hacli event fire ios.notification_sent --json '{"message": "Hello"}'
```

**Watch events:**

```bash
hacli event watch                    # Watch all events
hacli event watch state_changed      # Watch specific event type
```

### Configuration

**Get full configuration:**

```bash
hacli config full
```

**Get loaded components:**

```bash
hacli config components
```

**Get whitelisted directories:**

```bash
hacli config whitelist_dirs
```

**Get release version:**

```bash
hacli config release
```

### System

**Get error logs:**

```bash
hacli system log
```

**Get system health:**

```bash
hacli system health
```

### Templates

**Render template on server:**

```bash
hacli template my_template.j2
cat template.j2 | hacli template /dev/stdin
echo "{{ states('light.living_room') }}" > template.j2 && hacli template template.j2
```

**Render with variables:**

```bash
hacli template template.j2 data.json
```

### Raw API

**Raw GET request:**

```bash
hacli raw get states
hacli raw get config
hacli raw get "history/period/2024-01-01"
```

**Raw POST request:**

```bash
hacli raw post services/light/turn_on --json '{"entity_id": "light.living_room"}'
```

**Raw WebSocket call:**

```bash
hacli raw ws config/area_registry/list
hacli raw ws config/entity_registry/update --json '{"entity_id": "light.test", "name": "Test"}'
```

### Home Assistant OS/Supervisor (hass.io)

**Supervisor commands:**

```bash
hacli ha supervisor ping
hacli ha supervisor info
hacli ha supervisor update
hacli ha supervisor logs
hacli ha supervisor stats
```

**Host commands:**

```bash
hacli ha host reboot
hacli ha host shutdown
hacli ha host info
```

**OS commands:**

```bash
hacli ha os info
hacli ha os update
```

**Core commands:**

```bash
hacli ha core restart
hacli ha core stop
hacli ha core start
hacli ha core logs
```

**Hardware commands:**

```bash
hacli ha hardware info
hacli ha hardware audio
```

**Addons:**

```bash
hacli ha addons all
hacli ha addons reload
```

### Map

**Show Home Assistant location:**

```bash
hacli map
hacli map --service google
hacli map --service bing
```

**Show entity location:**

```bash
hacli map device_tracker.phone
hacli map person.john --service google
```

## Output Formats

### JSON

```bash
hacli state list -o json
hacli service list -o json | jq '.[] | select(.service | contains("light"))'
```

### YAML

```bash
hacli config full -o yaml
hacli entity list -o yaml
```

### Table (default)

```bash
hacli state list                    # Default columns: ENTITY, DESCRIPTION, STATE, CHANGED
hacli state list --columns entity_id,state,last_changed
hacli state list --sort-by entity_id
hacli state list --no-headers
```

### NDJSON (newline-delimited JSON)

```bash
hacli state list -o ndjson | jq -c '{entity: .entity_id, state: .state}'
```

### Auto

Automatically chooses format:

- `yaml` for data output
- `table` for lists (default)

## Shell Completion

**Bash:**

```bash
hacli completion bash > /etc/bash_completion.d/hacli
# OR
source <(hacli completion bash)
```

**Zsh:**

```bash
hacli completion zsh > "${fpath[1]}/_hacli"
# OR
source <(hacli completion zsh)
```

## Common Workflows

### Find and Control Lights

```bash
# List all lights
hacli state list "light.*"

# Turn on specific lights
hacli state turn_on light.living_room light.kitchen

# Set brightness and color
hacli service call light.turn_on -a entity_id=light.living_room,brightness=255,color_temp=300

# Toggle all lights in an area
hacli entity list | grep living_room
hacli state toggle light.living_room_light_1 light.living_room_light_2
```

### Organize Devices

```bash
# Create areas
hacli area create Kitchen Bedroom "Living Room"

# Assign devices to areas
hacli device list
hacli device assign Kitchen "Kitchen Light Switch"
hacli device assign Bedroom --match ".*bedroom.*"

# Rename devices for clarity
hacli device rename "zwave_js_123" "Kitchen Light Switch"
```

### Monitor Events

```bash
# Watch for state changes
hacli event watch state_changed

# Watch for specific entity updates
hacli event watch state_changed | jq 'select(.data.entity_id == "sensor.temperature")'

# Fire custom event
hacli event fire my_automation_triggered --json '{"source": "cli"}'
```

### Debug and Troubleshoot

```bash
# Check server info
hacli info

# View error logs
hacli system log

# Check system health
hacli system health

# Test raw API
hacli raw get states | jq '.[] | select(.state == "unavailable")'
```

### Backup and Restore (OS/Supervisor)

```bash
# Create snapshot
hacli ha snapshot reload

# View supervisor info
hacli ha supervisor info

# Restart Home Assistant Core
hacli ha core restart
```

## Development

### Building

```bash
make build          # Build binary to ./bin/hacli
go build -o hacli ./cmd/hacli
```

### Testing

```bash
make test
```

### Formatting and Linting

```bash
make fmt            # go fmt
make vet            # go vet
make all            # fmt, vet, test, build
```

### Running Locally

```bash
HA_SERVER=http://192.168.1.100:8123/ HA_TOKEN=your_token go run cmd/hacli/main.go state list
```

## Tips

- Use regex filters extensively: `hacli state list "sensor.temperature_.*"`
- Combine with jq for powerful filtering: `hacli state list -o json | jq ...`
- Use `--sort-by` for consistent table ordering
- WebSocket operations (entity/device/area registries) require token authentication
- History queries can be slow for many entities; filter by entity ID first
- Use `hacli info` to verify connectivity before running complex commands
