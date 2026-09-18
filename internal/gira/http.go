package gira

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

func doRaw(client *http.Client, req *http.Request, errorPrefix string) ([]byte, error) {
	response, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}
	defer response.Body.Close()

	if response.StatusCode < 200 || response.StatusCode > 299 {
		return nil, fmt.Errorf(
			"%s: %d %s %s",
			errorPrefix,
			response.StatusCode,
			http.StatusText(response.StatusCode),
			req.URL.String(),
		)
	}

	if response.StatusCode == http.StatusNoContent {
		return nil, nil
	}

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return body, nil
}

func doJSON(client *http.Client, req *http.Request, out any, errorPrefix string) error {
	body, err := doRaw(client, req, errorPrefix)
	if err != nil {
		return err
	}

	if out == nil || len(bytes.TrimSpace(body)) == 0 {
		return nil
	}

	if err := json.Unmarshal(body, out); err != nil {
		return fmt.Errorf("%s: %w", errorPrefix, err)
	}

	return nil
}

func marshalJSON(value any) ([]byte, error) {
	encoded, err := encodeJSON(value, "")
	if err != nil {
		return nil, err
	}

	return bytes.TrimRight(encoded, "\n"), nil
}
