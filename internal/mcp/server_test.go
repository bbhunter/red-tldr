package mcp

import (
	"testing"
)

func TestArgString(t *testing.T) {
	args := map[string]interface{}{
		"query":    "mimikatz",
		"platform": "windows",
	}

	if v := argString(args, "query", ""); v != "mimikatz" {
		t.Errorf("expected 'mimikatz', got '%s'", v)
	}
	if v := argString(args, "missing", "default"); v != "default" {
		t.Errorf("expected 'default', got '%s'", v)
	}
	if v := argString(args, "platform", ""); v != "windows" {
		t.Errorf("expected 'windows', got '%s'", v)
	}
}

func TestArgInt(t *testing.T) {
	args := map[string]interface{}{
		"limit": float64(20),
		"count": 5,
	}

	if v := argInt(args, "limit", 10); v != 20 {
		t.Errorf("expected 20 from float64, got %d", v)
	}
	if v := argInt(args, "count", 10); v != 5 {
		t.Errorf("expected 5 from int, got %d", v)
	}
	if v := argInt(args, "missing", 42); v != 42 {
		t.Errorf("expected 42 default, got %d", v)
	}
}

func TestArgString_NonStringValue(t *testing.T) {
	args := map[string]interface{}{
		"number": 123,
	}
	if v := argString(args, "number", "fallback"); v != "fallback" {
		t.Errorf("expected fallback for non-string value, got '%s'", v)
	}
}

func TestArgInt_StringValue(t *testing.T) {
	args := map[string]interface{}{
		"bad": "not_a_number",
	}
	if v := argInt(args, "bad", 99); v != 99 {
		t.Errorf("expected 99 fallback for string value, got %d", v)
	}
}
