package api

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// The logs API returns structured entries: {"entries":[{ts,msg,pod,...}],
// "truncated":bool}. The client used to decode a legacy {"logs":"..."}
// shape, so every response silently rendered as an empty string — during
// the 2026-07-02 Forge incident `ghayma logs` printed nothing while pods
// were crash-looping. These tests pin the new contract.

func TestGetAppLogs_ParsesEntries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/projects/p-123/logs" {
			t.Errorf("unexpected path %s", r.URL.Path)
		}
		if got := r.URL.Query().Get("lines"); got != "50" {
			t.Errorf("lines param = %q; want 50", got)
		}
		io.WriteString(w, `{"entries":[
			{"ts":"2026-07-02T10:51:54Z","msg":"Ready in 0ms","pod":"forge-abc","container":"forge","source":"stdout"},
			{"ts":"2026-07-02T10:52:01Z","msg":"GET / 200","pod":"forge-abc","container":"forge","source":"stdout"}
		],"truncated":false}`)
	}))
	defer ts.Close()

	logs, err := newTestClient(ts.URL).GetAppLogs("p-123", 50)
	if err != nil {
		t.Fatalf("GetAppLogs: %v", err)
	}
	for _, want := range []string{"Ready in 0ms", "GET / 200", "forge-abc"} {
		if !strings.Contains(logs, want) {
			t.Errorf("logs output should contain %q, got:\n%s", want, logs)
		}
	}
}

func TestGetAppLogs_EmptyEntries(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"entries":[],"truncated":false}`)
	}))
	defer ts.Close()

	logs, err := newTestClient(ts.URL).GetAppLogs("p-123", 100)
	if err != nil {
		t.Fatalf("GetAppLogs: %v", err)
	}
	if logs != "" {
		t.Errorf("expected empty string for zero entries, got %q", logs)
	}
}

// Non-200 responses must surface as errors, not decode into silence.
func TestGetAppLogs_ErrorStatus(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		io.WriteString(w, `{"error":"project not found"}`)
	}))
	defer ts.Close()

	_, err := newTestClient(ts.URL).GetAppLogs("p-nope", 100)
	if err == nil {
		t.Fatal("expected error for 404 response, got nil")
	}
	if !strings.Contains(err.Error(), "project not found") {
		t.Errorf("error should carry the server message, got: %v", err)
	}
}

func TestGetAppLogs_TruncatedNote(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, `{"entries":[{"ts":"2026-07-02T10:51:54Z","msg":"line","pod":"p","container":"c","source":"stdout"}],"truncated":true}`)
	}))
	defer ts.Close()

	logs, err := newTestClient(ts.URL).GetAppLogs("p-123", 100)
	if err != nil {
		t.Fatalf("GetAppLogs: %v", err)
	}
	if !strings.Contains(logs, "truncated") {
		t.Errorf("truncated responses should note it in the output, got:\n%s", logs)
	}
}
