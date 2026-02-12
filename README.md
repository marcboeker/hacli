# Home Assistant CLI

A command-line interface for Home Assistant written in Go.
Ispired by [home-assistant-ecosystem/home-assistant-cli](https://github.com/home-assistant-ecosystem/home-assistant-cli).

## Features

- **Entity Management**: List, get, edit, and delete entity states
- **Service Calls**: Invoke Home Assistant services
- **Device & Area Management**: Browse devices and areas
- **Configuration**: View and manage Home Assistant configuration
- **Events**: Listen to and interact with Home Assistant events
- **Templates**: Render Home Assistant templates
- **Multiple Output Formats**: JSON, YAML, table, and more

## Installation

### From Source

Requires Go 1.23 or later.

```bash
git clone https://github.com/marc/home-assistant-cli-go
cd home-assistant-cli-go
go build -o hacli ./cmd/hacli
```

### Using Make

```bash
make build
```

The binary will be available at `./bin/hacli`.

## Configuration

The CLI requires authentication to your Home Assistant server. Set these environment variables:

```bash
export HA_SERVER=http://your-home-assistant:8123/
export HA_TOKEN=your_long_lived_access_token
```

Or pass them as flags:

```bash
hacli --server http://your-home-assistant:8123/ --token your_token info
```

### Authentication Methods

- **Long-lived access token** (recommended): Generate in Home Assistant profile settings
- **Password**: Legacy authentication (not recommended)

## Usage

### Global Flags

- `-s, --server`: Home Assistant server URL
- `--token`: Access token
- `--password`: Password (legacy)
- `--timeout`: Request timeout
- `-o, --output`: Output format (json, yaml, table, auto, ndjson)
- `-v, --verbose`: Enable verbose output
- `-x, --insecure`: Ignore SSL certificate warnings
- `--debug`: Enable debug mode

### Commands

#### Info

Display Home Assistant server information:

```bash
hacli info
```

#### Entity State Management

List all entity states:

```bash
hacli state list
```

Filter entities:

```bash
hacli state list "light.*"
```

Get specific entity state:

```bash
hacli state get light.living_room
```

Edit entity state:

```bash
hacli state edit light.living_room on
```

Toggle entities:

```bash
hacli state toggle light.living_room light.kitchen
hacli state turn_on light.living_room
hacli state turn_off light.living_room
```

View entity history:

```bash
hacli state history light.living_room --since 1h
```

#### Services

List available services:

```bash
hacli service list
```

Call a service:

```bash
hacli service call light.turn_on light.living_room
```

#### Devices

List devices:

```bash
hacli device list
```

Get device details:

```bash
hacli device get <device_id>
```

#### Areas

List areas:

```bash
hacli area list
```

Get area details:

```bash
hacli area get <area_id>
```

#### Configuration

View configuration:

```bash
hacli config
```

Check config:

```bash
hacli config check
```

#### Events

List event listeners:

```bash
hacli event list
```

Fire an event:

```bash
hacli event fire my_custom_event
```

Watch events:

```bash
hacli event watch
```

#### Templates

Render a template:

```bash
hacli template "{{ states('light.living_room') }}"
```

#### Raw API Calls

Make raw API requests:

```bash
hacli raw get states
hacli raw post services/homeassistant/turn_on
```

#### System

View system information:

```bash
hacli system
```

## Output Formats

- `json`: Pretty-printed JSON
- `yaml`: YAML format
- `table`: Tabular format (default for most commands)
- `auto`: Automatically choose format based on output
- `ndjson`: Newline-delimited JSON

Example:

```bash
hacli state list -o json
hacli device list -o yaml
```

## Shell Completion

Generate shell completion scripts:

```bash
# Bash
hacli completion bash > /etc/bash_completion.d/hacli

# Zsh
hacli completion zsh > "${fpath[1]}/_hacli"

# Fish
hacli completion fish > ~/.config/fish/completions/hacli.fish
```

## Development

### Building

```bash
make build
```

### Testing

```bash
make test
```

### Formatting and Linting

```bash
make fmt
make vet
```

### Full Build Pipeline

```bash
make all  # fmt, vet, test, build
```

## License

MIT License - see [LICENSE.md](LICENSE.md)
