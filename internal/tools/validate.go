package tools

import (
	"encoding/json"
	"fmt"
	"strings"
)

type ValidationError struct {
	Field   string
	Message string
}

func (e ValidationError) Error() string {
	if e.Field != "" {
		return fmt.Sprintf("%s: %s", e.Field, e.Message)
	}
	return e.Message
}

type ValidationErrors []ValidationError

func (ve ValidationErrors) Error() string {
	msgs := make([]string, len(ve))
	for i, e := range ve {
		msgs[i] = e.Error()
	}
	return strings.Join(msgs, "; ")
}

func ValidateInput(schema any, input json.RawMessage) error {
	schemaMap, ok := schema.(map[string]any)
	if !ok {
		return nil
	}

	var data map[string]any
	if len(input) == 0 || string(input) == "null" {
		data = map[string]any{}
	} else {
		if err := json.Unmarshal(input, &data); err != nil {
			return ValidationError{Message: "arguments must be a JSON object"}
		}
	}

	var errs ValidationErrors

	for _, name := range extractRequired(schemaMap) {
		if _, exists := data[name]; !exists {
			errs = append(errs, ValidationError{Field: name, Message: "required"})
		}
	}

	if properties, ok := schemaMap["properties"].(map[string]any); ok {
		for key, val := range data {
			propSchema, exists := properties[key]
			if !exists {
				continue
			}
			propMap, ok := propSchema.(map[string]any)
			if !ok {
				continue
			}
			expectedType, _ := propMap["type"].(string)
			if expectedType == "" {
				continue
			}
			if err := checkType(key, val, expectedType); err != nil {
				errs = append(errs, *err)
			}
		}
	}

	if len(errs) > 0 {
		return errs
	}
	return nil
}

func extractRequired(schema map[string]any) []string {
	switch r := schema["required"].(type) {
	case []any:
		out := make([]string, 0, len(r))
		for _, v := range r {
			if s, ok := v.(string); ok && s != "" {
				out = append(out, s)
			}
		}
		return out
	case []string:
		return r
	}
	return nil
}

func checkType(field string, value any, expectedType string) *ValidationError {
	var ok bool
	switch expectedType {
	case "string":
		_, ok = value.(string)
	case "number":
		_, ok = value.(float64)
	case "integer":
		f, isFloat := value.(float64)
		ok = isFloat && f == float64(int64(f))
	case "boolean":
		_, ok = value.(bool)
	case "object":
		_, ok = value.(map[string]any)
	case "array":
		_, ok = value.([]any)
	default:
		return nil
	}
	if !ok {
		return &ValidationError{Field: field, Message: fmt.Sprintf("expected %s", expectedType)}
	}
	return nil
}
