package app

import (
	"bytes"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/tidwall/gjson"
	"gopkg.in/yaml.v3"
)

func PrintData(ctx *AppContext, data any, defaultOutput string) error {
	effective := ctx.EffectiveOutput(defaultOutput)

	switch effective {
	case "json":
		body, err := json.MarshalIndent(data, "", "  ")
		if err != nil {
			return fmt.Errorf("marshal json: %w", err)
		}
		_, err = fmt.Fprintln(ctx.Stdout, string(body))
		return err
	case "ndjson":
		items, ok := data.([]any)
		if !ok {
			body, err := json.Marshal(data)
			if err != nil {
				return fmt.Errorf("marshal ndjson: %w", err)
			}
			_, err = fmt.Fprintln(ctx.Stdout, string(body))
			return err
		}
		for _, item := range items {
			body, err := json.Marshal(item)
			if err != nil {
				return fmt.Errorf("marshal ndjson item: %w", err)
			}
			if _, err := fmt.Fprintln(ctx.Stdout, string(body)); err != nil {
				return err
			}
		}
		return nil
	case "yaml":
		body, err := yaml.Marshal(data)
		if err != nil {
			return fmt.Errorf("marshal yaml: %w", err)
		}
		_, err = fmt.Fprintln(ctx.Stdout, string(body))
		return err
	case "table":
		return printTable(ctx, data)
	default:
		return fmt.Errorf("unknown output format: %s", effective)
	}
}

func printTable(ctx *AppContext, data any) error {
	rows := normalizeRows(data)
	if ctx.SortBy != "" {
		sort.SliceStable(rows, func(left int, right int) bool {
			lval := rowPathValue(rows[left], ctx.SortBy)
			rval := rowPathValue(rows[right], ctx.SortBy)
			if lval == "" && rval != "" {
				return false
			}
			if lval != "" && rval == "" {
				return true
			}
			return lval < rval
		})
	}

	if len(rows) == 0 {
		_, err := fmt.Fprintln(ctx.Stdout)
		return err
	}

	var headers []string
	var paths []string

	if len(ctx.ColumnDefs) > 0 {
		// Use column definitions with display names and jsonpaths
		for _, def := range ctx.ColumnDefs {
			headers = append(headers, def.Header)
			paths = append(paths, def.Path)
		}
	} else if len(ctx.Columns) > 0 {
		// Simple column names (backward compatible)
		headers = ctx.Columns
		paths = ctx.Columns
	} else {
		// Auto-detect from data keys
		headers = sortedKeys(rows)
		paths = headers
	}

	buffer := &bytes.Buffer{}
	table := tablewriter.NewWriter(buffer)

	if ctx.TableFormat == "markdown" {
		// Markdown format
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("|")
		table.SetColumnSeparator("|")
		table.SetRowSeparator("")
		table.SetBorder(false)
		table.SetHeaderLine(true)
		table.SetAutoFormatHeaders(false)
	} else {
		// Plain text format
		table.SetAutoWrapText(false)
		table.SetHeaderAlignment(tablewriter.ALIGN_LEFT)
		table.SetAlignment(tablewriter.ALIGN_LEFT)
		table.SetCenterSeparator("")
		table.SetColumnSeparator("")
		table.SetRowSeparator("")
		table.SetBorder(false)
		table.SetHeaderLine(false)
	}

	if ctx.NoHeaders {
		table.SetHeader([]string{})
	} else {
		table.SetHeader(headers)
	}

	for _, row := range rows {
		values := make([]string, 0, len(paths))
		for _, path := range paths {
			values = append(values, rowPathValue(row, path))
		}
		table.Append(values)
	}

	table.Render()
	output := strings.TrimRight(buffer.String(), "\n")

	if ctx.TableFormat != "markdown" {
		// Trim leading/trailing whitespace from each line for plain text
		lines := strings.Split(output, "\n")
		for i, line := range lines {
			lines[i] = strings.TrimSpace(line)
		}
		output = strings.Join(lines, "\n")
	}

	_, err := fmt.Fprintln(ctx.Stdout, output)
	return err
}

func normalizeRows(data any) []map[string]any {
	if data == nil {
		return nil
	}

	switch value := data.(type) {
	case []any:
		rows := make([]map[string]any, 0, len(value))
		for _, item := range value {
			if row, ok := item.(map[string]any); ok {
				rows = append(rows, row)
			}
		}
		return rows
	case map[string]any:
		return []map[string]any{value}
	default:
		return []map[string]any{{"value": value}}
	}
}

func sortedKeys(rows []map[string]any) []string {
	keysSet := map[string]struct{}{}
	for _, row := range rows {
		for key := range row {
			keysSet[key] = struct{}{}
		}
	}
	keys := make([]string, 0, len(keysSet))
	for key := range keysSet {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	return keys
}

func rowPathValue(row map[string]any, path string) string {
	encoded, err := json.Marshal(row)
	if err != nil {
		return ""
	}
	return gjson.GetBytes(encoded, path).String()
}

func stringify(value any) string {
	switch typed := value.(type) {
	case nil:
		return ""
	case string:
		return typed
	default:
		encoded, err := json.Marshal(typed)
		if err != nil {
			return fmt.Sprint(typed)
		}
		return string(encoded)
	}
}
