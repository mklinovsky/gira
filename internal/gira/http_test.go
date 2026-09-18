package gira

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDoJSONDecodesSuccess(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(`{"key":"ABC-1"}`))
	}))
	defer server.Close()

	var out struct {
		Key string `json:"key"`
	}
	if err := doJSON(server.Client(), request(t, http.MethodGet, server.URL, nil), &out, "TEST API"); err != nil {
		t.Fatalf("doJSON returned error: %v", err)
	}

	if out.Key != "ABC-1" {
		t.Errorf("key = %q, want %q", out.Key, "ABC-1")
	}
}

func TestDoJSONAcceptsNoContent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	var out map[string]any
	if err := doJSON(server.Client(), request(t, http.MethodPut, server.URL, nil), &out, "TEST API"); err != nil {
		t.Fatalf("doJSON returned error: %v", err)
	}
}

func TestDoJSONWrapsHTTPErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	err := doJSON(server.Client(), request(t, http.MethodGet, server.URL+"/nope", nil), nil, "TEST API")

	if err == nil {
		t.Fatal("doJSON succeeded, want error")
	}
	want := "TEST API: 404 Not Found " + server.URL + "/nope"
	if err.Error() != want {
		t.Errorf("error = %q, want %q", err.Error(), want)
	}
}

func TestDoJSONWrapsTransportErrors(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	client := server.Client()
	server.Close()

	err := doJSON(client, request(t, http.MethodGet, server.URL, nil), nil, "GitLab API")

	if err == nil {
		t.Fatal("doJSON succeeded, want error")
	}
	if !strings.HasPrefix(err.Error(), "GitLab API: ") {
		t.Errorf("error = %q, want the GitLab API prefix", err.Error())
	}
}

func TestDoRawKeepsBodyByteForByte(t *testing.T) {
	body := `{"zeta":1,"id":1234567890123456789,"alpha":"<tag>"}`
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Write([]byte(body))
	}))
	defer server.Close()

	raw, err := doRaw(server.Client(), request(t, http.MethodGet, server.URL, nil), "TEST API")
	if err != nil {
		t.Fatalf("doRaw returned error: %v", err)
	}

	if string(raw) != body {
		t.Errorf("body = %q, want %q", raw, body)
	}
}

func TestMarshalJSONDoesNotEscapeHTML(t *testing.T) {
	got, err := marshalJSON(map[string]string{"summary": "a < b & c > d"})
	if err != nil {
		t.Fatalf("marshalJSON returned error: %v", err)
	}

	want := `{"summary":"a < b & c > d"}`
	if string(got) != want {
		t.Errorf("marshalJSON = %q, want %q", got, want)
	}
}

func request(t *testing.T, method, url string, body []byte) *http.Request {
	t.Helper()

	var reader *strings.Reader
	if body != nil {
		reader = strings.NewReader(string(body))
	}

	var req *http.Request
	var err error
	if reader == nil {
		req, err = http.NewRequestWithContext(context.Background(), method, url, nil)
	} else {
		req, err = http.NewRequestWithContext(context.Background(), method, url, reader)
	}
	if err != nil {
		t.Fatalf("building request: %v", err)
	}

	return req
}
