package models

import (
	"database/sql/driver"
	"fmt"
	"strings"
)

type StringArray []string

func (a StringArray) Value() (driver.Value, error) {
	if a == nil {
		return "{}", nil
	}

	escaped := make([]string, len(a))
	for i, v := range a {
		v = strings.ReplaceAll(v, `\`, `\\`)
		v = strings.ReplaceAll(v, `"`, `\"`)
		escaped[i] = `"` + v + `"`
	}

	return "{" + strings.Join(escaped, ",") + "}", nil
}

func (a *StringArray) Scan(src any) error {
	switch v := src.(type) {
	case nil:
		*a = nil
		return nil
	case string:
		return a.parse(v)
	case []byte:
		return a.parse(string(v))
	default:
		return fmt.Errorf("cannot scan %T into StringArray", src)
	}
}

func (a *StringArray) parse(raw string) error {
	raw = strings.TrimSpace(raw)
	if !strings.HasPrefix(raw, "{") || !strings.HasSuffix(raw, "}") {
		return fmt.Errorf("invalid array literal: %q", raw)
	}

	raw = raw[1 : len(raw)-1]
	if raw == "" {
		*a = StringArray{}
		return nil
	}

	var (
		result   []string
		current  strings.Builder
		inQuotes bool
	)

	for i := 0; i < len(raw); i++ {
		switch ch := raw[i]; {
		case ch == '\\' && i+1 < len(raw):
			i++
			current.WriteByte(raw[i])
		case ch == '"':
			inQuotes = !inQuotes
		case ch == ',' && !inQuotes:
			result = append(result, current.String())
			current.Reset()
		default:
			current.WriteByte(ch)
		}
	}
	result = append(result, current.String())

	*a = result
	return nil
}
