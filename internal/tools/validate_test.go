package tools

import (
	"encoding/json"
	"testing"
)

func TestValidateRequiredPresent(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}
	input := json.RawMessage(`{"name": "hello"}`)
	if err := ValidateInput(schema, input); err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestValidateRequiredMissing(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"name": map[string]any{"type": "string"},
		},
		"required": []any{"name"},
	}
	input := json.RawMessage(`{}`)
	err := ValidateInput(schema, input)
	if err == nil {
		t.Fatal("expected error for missing required field")
	}
	if ve, ok := err.(ValidationErrors); ok {
		if len(ve) != 1 || ve[0].Field != "name" {
			t.Errorf("unexpected validation error: %v", ve)
		}
	}
}

func TestValidateTypeString(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"msg": map[string]any{"type": "string"},
		},
	}
	err := ValidateInput(schema, json.RawMessage(`{"msg": 42}`))
	if err == nil {
		t.Fatal("expected type error")
	}
}

func TestValidateTypeNumber(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"count": map[string]any{"type": "number"},
		},
	}
	if err := ValidateInput(schema, json.RawMessage(`{"count": 3.14}`)); err != nil {
		t.Errorf("expected no error for number, got %v", err)
	}
	if err := ValidateInput(schema, json.RawMessage(`{"count": "three"}`)); err == nil {
		t.Error("expected error for string where number expected")
	}
}

func TestValidateTypeInteger(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"n": map[string]any{"type": "integer"},
		},
	}
	if err := ValidateInput(schema, json.RawMessage(`{"n": 5}`)); err != nil {
		t.Errorf("expected no error for integer, got %v", err)
	}
	if err := ValidateInput(schema, json.RawMessage(`{"n": 5.5}`)); err == nil {
		t.Error("expected error for float where integer expected")
	}
}

func TestValidateNilInput(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"opt": map[string]any{"type": "string"},
		},
	}
	if err := ValidateInput(schema, nil); err != nil {
		t.Errorf("expected no error for nil input, got %v", err)
	}
}

func TestValidateNoSchema(t *testing.T) {
	if err := ValidateInput(nil, json.RawMessage(`{"a": 1}`)); err != nil {
		t.Errorf("expected no error when schema is nil, got %v", err)
	}
}

func TestValidateInvalidJSON(t *testing.T) {
	schema := map[string]any{"type": "object"}
	err := ValidateInput(schema, json.RawMessage(`not json`))
	if err == nil {
		t.Fatal("expected error for invalid JSON")
	}
}

func TestValidateMultipleErrors(t *testing.T) {
	schema := map[string]any{
		"type": "object",
		"properties": map[string]any{
			"a": map[string]any{"type": "string"},
			"b": map[string]any{"type": "string"},
		},
		"required": []any{"a", "b"},
	}
	err := ValidateInput(schema, json.RawMessage(`{}`))
	if err == nil {
		t.Fatal("expected errors")
	}
	ve, ok := err.(ValidationErrors)
	if !ok {
		t.Fatalf("expected ValidationErrors, got %T", err)
	}
	if len(ve) != 2 {
		t.Errorf("expected 2 errors, got %d", len(ve))
	}
}
