package gira

import (
	"encoding/json"
	"errors"
	"strings"
)

func ParseCustomField(customField string) (map[string]json.RawMessage, error) {
	if customField == "" {
		return nil, nil
	}

	key, value, found := strings.Cut(customField, "=")
	if key == "" || !found {
		return nil, errors.New("Invalid custom field format. Expected key=value.")
	}

	var probe any
	if json.Unmarshal([]byte(value), &probe) == nil {
		return map[string]json.RawMessage{key: json.RawMessage(value)}, nil
	}

	quoted, err := json.Marshal(value)
	if err != nil {
		return nil, err
	}

	return map[string]json.RawMessage{key: quoted}, nil
}
