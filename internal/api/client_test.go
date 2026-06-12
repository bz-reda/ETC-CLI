package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"paas-cli/internal/config"
)

func newTestClient(url string) *Client {
	return NewClient(&config.Config{APIHost: url, Token: "test-token"})
}

// TestEligibleBillingAccounts pins the project-create filter: only
// active accounts the caller owns or admins are eligible (mirrors the
// dashboard gate). Suspended/closed accounts and viewer-role accounts
// must be excluded so init never offers an account the API would reject.
func TestEligibleBillingAccounts(t *testing.T) {
	accounts := []BillingAccount{
		{ID: "1", Name: "personal", Status: "active", Role: "owner", IsPersonal: true},
		{ID: "2", Name: "team-admin", Status: "active", Role: "admin"},
		{ID: "3", Name: "team-viewer", Status: "active", Role: "viewer"}, // drop: viewer
		{ID: "4", Name: "suspended", Status: "suspended", Role: "owner"}, // drop: not active
		{ID: "5", Name: "closed", Status: "closed", Role: "owner"},       // drop: not active
	}
	got := EligibleBillingAccounts(accounts)
	if len(got) != 2 {
		t.Fatalf("got %d eligible; want 2 (active owner+admin). got=%+v", len(got), got)
	}
	if got[0].ID != "1" || got[1].ID != "2" {
		t.Errorf("eligible ids = %s,%s; want 1,2 (preserves order)", got[0].ID, got[1].ID)
	}
}

// TestListBillingAccounts_ParsesEnvelope confirms the client unwraps the
// {"accounts":[...]} envelope and sends the Bearer token.
func TestListBillingAccounts_ParsesEnvelope(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/api/v1/billing-accounts" || r.Method != http.MethodGet {
			t.Errorf("unexpected request %s %s", r.Method, r.URL.Path)
		}
		if got := r.Header.Get("Authorization"); got != "Bearer test-token" {
			t.Errorf("auth header = %q; want Bearer test-token", got)
		}
		io.WriteString(w, `{"accounts":[{"id":"a1","name":"Personal","status":"active","role":"owner","is_personal":true}]}`)
	}))
	defer ts.Close()

	accounts, err := newTestClient(ts.URL).ListBillingAccounts()
	if err != nil {
		t.Fatalf("ListBillingAccounts: %v", err)
	}
	if len(accounts) != 1 || accounts[0].ID != "a1" || !accounts[0].IsPersonal {
		t.Errorf("parsed = %+v; want one personal account a1", accounts)
	}
}

// TestCreateProject_SendsBillingAccountID is the core regression pin for
// the CLI billing blocker: CreateProject MUST include billing_account_id
// in the POST body when provided.
func TestCreateProject_SendsBillingAccountID(t *testing.T) {
	var gotBody map[string]string
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":"p1","name":"demo","slug":"demo","framework":"nextjs"}`)
	}))
	defer ts.Close()

	p, err := newTestClient(ts.URL).CreateProject("demo", "nextjs", "acct-123")
	if err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if p.ID != "p1" {
		t.Errorf("project id = %q; want p1", p.ID)
	}
	if gotBody["billing_account_id"] != "acct-123" {
		t.Errorf("body billing_account_id = %q; want acct-123 (body=%v)", gotBody["billing_account_id"], gotBody)
	}
	if gotBody["name"] != "demo" || gotBody["framework"] != "nextjs" {
		t.Errorf("body name/framework wrong: %v", gotBody)
	}
}

// TestCreateProject_OmitsBillingAccountWhenEmpty confirms the field is
// absent (not sent as "") when no account is supplied, so the server's
// own default/validation applies cleanly.
func TestCreateProject_OmitsBillingAccountWhenEmpty(t *testing.T) {
	var gotBody map[string]json.RawMessage
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		raw, _ := io.ReadAll(r.Body)
		_ = json.Unmarshal(raw, &gotBody)
		w.WriteHeader(http.StatusCreated)
		io.WriteString(w, `{"id":"p1"}`)
	}))
	defer ts.Close()

	if _, err := newTestClient(ts.URL).CreateProject("demo", "nextjs", ""); err != nil {
		t.Fatalf("CreateProject: %v", err)
	}
	if _, present := gotBody["billing_account_id"]; present {
		t.Errorf("billing_account_id should be omitted when empty; body=%v", gotBody)
	}
}

// TestCreateProject_DecodesAPIError confirms a non-201 surfaces the
// server's {"error":...} message (e.g. the BILLING_ACCOUNT_REQUIRED 400)
// rather than the old raw-JSON dump.
func TestCreateProject_DecodesAPIError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		io.WriteString(w, `{"error":"billing_account_id is required for plan=hobby","code":"BILLING_ACCOUNT_REQUIRED"}`)
	}))
	defer ts.Close()

	_, err := newTestClient(ts.URL).CreateProject("demo", "nextjs", "")
	if err == nil {
		t.Fatal("want error on 400")
	}
	if !strings.Contains(err.Error(), "billing_account_id is required") {
		t.Errorf("error = %q; want the decoded server message", err.Error())
	}
	if strings.Contains(err.Error(), "failed to create project: {") {
		t.Errorf("error still dumps the raw JSON body: %q", err.Error())
	}
}
