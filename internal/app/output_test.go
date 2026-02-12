package app

import (
	"bytes"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNormalizeRows(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected []map[string]any
	}{
		{
			name:     "nil",
			input:    nil,
			expected: nil,
		},
		{
			name:     "empty_slice",
			input:    []any{},
			expected: []map[string]any{},
		},
		{
			name: "slice_of_maps",
			input: []any{
				map[string]any{"id": "1", "name": "test1"},
				map[string]any{"id": "2", "name": "test2"},
			},
			expected: []map[string]any{
				{"id": "1", "name": "test1"},
				{"id": "2", "name": "test2"},
			},
		},
		{
			name:     "single_map",
			input:    map[string]any{"id": "1", "name": "test"},
			expected: []map[string]any{{"id": "1", "name": "test"}},
		},
		{
			name:     "primitive",
			input:    "hello",
			expected: []map[string]any{{"value": "hello"}},
		},
		{
			name:     "int",
			input:    42,
			expected: []map[string]any{{"value": 42}},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := normalizeRows(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestSortedKeys(t *testing.T) {
	rows := []map[string]any{
		{"z": 1, "a": 2, "m": 3},
		{"b": 4, "a": 5},
	}

	result := sortedKeys(rows)
	assert.Equal(t, []string{"a", "b", "m", "z"}, result)
}

func TestSortedKeys_Empty(t *testing.T) {
	rows := []map[string]any{}
	result := sortedKeys(rows)
	assert.Empty(t, result)
}

func TestRowPathValue(t *testing.T) {
	row := map[string]any{
		"entity_id": "light.living_room",
		"state":     "on",
		"attributes": map[string]any{
			"brightness": 255,
		},
	}

	tests := []struct {
		name     string
		path     string
		expected string
	}{
		{"top_level", "entity_id", "light.living_room"},
		{"nested", "attributes.brightness", "255"},
		{"missing", "nonexistent", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := rowPathValue(row, tt.path)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestStringify(t *testing.T) {
	tests := []struct {
		name     string
		input    any
		expected string
	}{
		{"nil", nil, ""},
		{"string", "hello", "hello"},
		{"int", 42, "42"},
		{"float", 3.14, "3.14"},
		{"bool", true, "true"},
		{"slice", []string{"a", "b"}, `["a","b"]`},
		{"map", map[string]int{"x": 1}, `{"x":1}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stringify(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestPrintData_JSON(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "json",
		Stdout: &buf,
	}

	data := map[string]any{
		"name":  "test",
		"value": 42,
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `"name": "test"`)
	assert.Contains(t, output, `"value": 42`)
}

func TestPrintData_YAML(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "yaml",
		Stdout: &buf,
	}

	data := map[string]any{
		"name":  "test",
		"value": 42,
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "name: test")
	assert.Contains(t, output, "value: 42")
}

func TestPrintData_NDJSON_Single(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "ndjson",
		Stdout: &buf,
	}

	data := map[string]any{
		"name": "test",
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, `{"name":"test"}`)
}

func TestPrintData_NDJSON_Array(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "ndjson",
		Stdout: &buf,
	}

	data := []any{
		map[string]any{"id": "1"},
		map[string]any{"id": "2"},
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	lines := bytes.Split(bytes.TrimSpace(buf.Bytes()), []byte("\n"))
	assert.Len(t, lines, 2)
	assert.Contains(t, output, `{"id":"1"}`)
	assert.Contains(t, output, `{"id":"2"}`)
}

func TestPrintData_UnknownFormat(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "unknown",
		Stdout: &buf,
	}

	err := PrintData(appCtx, "data", "")
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown output format")
}

func TestPrintData_Table(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output:  "table",
		Stdout:  &buf,
		Columns: []string{"entity_id", "state"},
	}

	data := []any{
		map[string]any{
			"entity_id": "light.living_room",
			"state":     "on",
		},
		map[string]any{
			"entity_id": "switch.kitchen",
			"state":     "off",
		},
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	// tablewriter formats headers with spaces instead of underscores
	assert.Contains(t, output, "ENTITY ID")
	assert.Contains(t, output, "STATE")
	assert.Contains(t, output, "light.living_room")
	assert.Contains(t, output, "switch.kitchen")
}

func TestPrintData_Table_NoHeaders(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output:    "table",
		Stdout:    &buf,
		Columns:   []string{"entity_id"},
		NoHeaders: true,
	}

	data := []any{
		map[string]any{"entity_id": "light.test"},
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	// Should not contain header
	assert.NotContains(t, output, "ENTITY_ID")
	assert.Contains(t, output, "light.test")
}

func TestPrintData_Table_SortBy(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output:  "table",
		Stdout:  &buf,
		Columns: []string{"name"},
		SortBy:  "name",
	}

	data := []any{
		map[string]any{"name": "zebra"},
		map[string]any{"name": "alpha"},
		map[string]any{"name": "beta"},
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	// Check that alpha comes before beta and zebra
	alphaIdx := bytes.Index(buf.Bytes(), []byte("alpha"))
	betaIdx := bytes.Index(buf.Bytes(), []byte("beta"))
	zebraIdx := bytes.Index(buf.Bytes(), []byte("zebra"))

	assert.True(t, alphaIdx < betaIdx)
	assert.True(t, betaIdx < zebraIdx)
}

func TestPrintData_Table_EmptyData(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output:  "table",
		Stdout:  &buf,
		Columns: []string{"name"},
	}

	data := []any{}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	// Should just output a newline
	assert.Equal(t, "\n", buf.String())
}

func TestPrintData_Table_WithColumnDefs(t *testing.T) {
	var buf bytes.Buffer
	appCtx := &AppContext{
		Output: "table",
		Stdout: &buf,
		ColumnDefs: []ColumnDef{
			{Header: "ID", Path: "entity_id"},
			{Header: "Status", Path: "state"},
		},
	}

	data := []any{
		map[string]any{
			"entity_id": "light.living_room",
			"state":     "on",
		},
	}

	err := PrintData(appCtx, data, "")
	require.NoError(t, err)

	output := buf.String()
	assert.Contains(t, output, "ID")
	assert.Contains(t, output, "STATUS")
}
