package gira

import (
	"encoding/json"
	"testing"

	"github.com/google/go-cmp/cmp"
)

func TestParseCustomField(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    map[string]string
		wantErr string
	}{
		{name: "empty input", input: "", want: nil},
		{name: "simple string value", input: "key=value", want: map[string]string{"key": `"value"`}},
		{name: "json object value", input: `key={"id":"123"}`, want: map[string]string{"key": `{"id":"123"}`}},
		{name: "equals signs in the value", input: "key=value1=value2", want: map[string]string{"key": `"value1=value2"`}},
		{name: "json array value", input: `labels=["bug","urgent"]`, want: map[string]string{"labels": `["bug","urgent"]`}},
		{name: "number value keeps its literal", input: "customfield_1=12345678901234567890", want: map[string]string{"customfield_1": "12345678901234567890"}},
		{name: "empty value falls back to a string", input: "key=", want: map[string]string{"key": `""`}},
		{name: "key only", input: "keyonly", wantErr: "Invalid custom field format. Expected key=value."},
		{name: "missing key", input: "=value", wantErr: "Invalid custom field format. Expected key=value."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseCustomField(tt.input)

			if tt.wantErr != "" {
				if err == nil {
					t.Fatalf("ParseCustomField(%q) = %v, want error", tt.input, got)
				}
				if err.Error() != tt.wantErr {
					t.Errorf("error = %q, want %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseCustomField(%q) returned error: %v", tt.input, err)
			}
			if diff := cmp.Diff(tt.want, rawStrings(got)); diff != "" {
				t.Errorf("ParseCustomField(%q) mismatch (-want +got):\n%s", tt.input, diff)
			}
		})
	}
}

func rawStrings(fields map[string]json.RawMessage) map[string]string {
	if fields == nil {
		return nil
	}

	out := make(map[string]string, len(fields))
	for key, value := range fields {
		out[key] = string(value)
	}

	return out
}
