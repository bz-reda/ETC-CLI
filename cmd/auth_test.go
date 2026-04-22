package cmd

import (
	"strings"
	"testing"
)

func TestLooksLikeSecret(t *testing.T) {
	secret := []string{
		"google_client_secret",
		"github_client_secret",
		"password",
		"api_token",
		"private_key",
		"GOOGLE_CLIENT_SECRET", // case insensitive
		"refresh_token",
	}
	public := []string{
		"google_client_id",
		"github_client_id",
		"jwt_expiry_seconds",
		"refresh_expiry_seconds",
		"email_verification_required",
		"allowed_origins",
		"name",
	}
	for _, k := range secret {
		if !looksLikeSecret(k) {
			t.Errorf("expected %q to be flagged as secret", k)
		}
	}
	for _, k := range public {
		if looksLikeSecret(k) {
			t.Errorf("expected %q NOT to be flagged as secret", k)
		}
	}
}

// TestFormatUpdatesList_NeverLeaksSecretValues is the load-bearing test for
// the PR A security fix. `auth config` previously echoed the full `updates`
// map back to the terminal after a successful update, which printed
// google_client_secret / github_client_secret in plaintext.
//
// Fixture values are deliberately un-prefixed (no GOCSPX-, ghp_,
// .apps.googleusercontent.com, Iv1.) so GitHub's push-protection secret
// scanner doesn't match them as real credentials.
func TestFormatUpdatesList_NeverLeaksSecretValues(t *testing.T) {
	googleSecret := "FAKE-REDACT-ME-GOOGLE-fixture-value"
	githubSecret := "FAKE-REDACT-ME-GITHUB-fixture-value"

	updates := map[string]interface{}{
		"google_client_id":     "FAKE-PUBLIC-GOOGLE-CLIENT-ID",
		"google_client_secret": googleSecret,
		"github_client_id":     "FAKE-PUBLIC-GITHUB-CLIENT-ID",
		"github_client_secret": githubSecret,
		"jwt_expiry_seconds":   3600,
		"allowed_origins":      []string{"https://a.example.com"},
	}

	out := strings.Join(formatUpdatesList(updates), "\n")

	for _, leaked := range []string{googleSecret, githubSecret} {
		if strings.Contains(out, leaked) {
			t.Errorf("secret value leaked into output:\n%s", out)
		}
	}

	// Redaction marker must be present for both secret keys.
	for _, key := range []string{"google_client_secret", "github_client_secret"} {
		wantLine := key + " = <redacted>"
		if !strings.Contains(out, wantLine) {
			t.Errorf("expected redacted marker %q in output:\n%s", wantLine, out)
		}
	}

	// Non-secret values must still be visible.
	for _, want := range []string{
		"google_client_id = FAKE-PUBLIC-GOOGLE-CLIENT-ID",
		"github_client_id = FAKE-PUBLIC-GITHUB-CLIENT-ID",
		"jwt_expiry_seconds = 3600",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("expected %q in output, got:\n%s", want, out)
		}
	}
}
