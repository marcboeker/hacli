package app

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func ParseTimeOrDuration(input string) (time.Time, error) {
	input = strings.TrimSpace(strings.ToLower(input))

	if input == "" || input == "now" {
		return time.Now(), nil
	}

	re := regexp.MustCompile(`^(\d+)([hdm])$`)
	matches := re.FindStringSubmatch(input)
	if matches != nil {
		amount, err := strconv.Atoi(matches[1])
		if err != nil {
			return time.Time{}, fmt.Errorf("invalid duration: %s", input)
		}

		unit := matches[2]
		now := time.Now()
		switch unit {
		case "h":
			return now.Add(-time.Duration(amount) * time.Hour), nil
		case "d":
			return now.Add(-time.Duration(amount) * 24 * time.Hour), nil
		case "m":
			return now.Add(-time.Duration(amount) * time.Minute), nil
		}
	}

	t, err := time.Parse(time.RFC3339, input)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("2006-01-02T15:04:05", input)
	if err == nil {
		return t, nil
	}

	t, err = time.Parse("2006-01-02", input)
	if err == nil {
		return t, nil
	}

	return time.Time{}, fmt.Errorf("invalid time or duration: %s", input)
}

func ParseObject(raw string) (map[string]any, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]any{}, nil
	}

	var jsonObj map[string]any
	if err := json.Unmarshal([]byte(raw), &jsonObj); err == nil {
		return jsonObj, nil
	}

	values := map[string]any{}
	pairs := strings.Split(raw, ",")
	for _, pair := range pairs {
		items := strings.SplitN(pair, "=", 2)
		if len(items) != 2 {
			return nil, fmt.Errorf("invalid key/value pair: %q", pair)
		}
		key := strings.TrimSpace(items[0])
		value := strings.TrimSpace(items[1])
		if key == "" {
			return nil, fmt.Errorf("invalid empty key in arguments")
		}
		values[key] = value
	}

	return values, nil
}
