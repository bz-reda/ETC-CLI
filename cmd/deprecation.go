package cmd

import (
	"fmt"
	"os"
	"strings"
	"time"

	"paas-cli/internal/config"
)

// deprecationInterval throttles repeat warnings. Once per week per notice
// strikes a balance between not being nagged and not missing the signal
// when the user hasn't touched the CLI in a while.
const deprecationInterval = 7 * 24 * time.Hour

// maybeWarnDeprecated prints a deprecation warning to stderr for oldPath
// (space-separated command path like "site add") unless the same notice was
// already shown in the last deprecationInterval. Intended to be called from
// the hidden alias command's Run/PreRun before it delegates to the real
// command's logic.
//
// The removalVersion gives users a concrete planning horizon — critical
// for anyone scripting against the CLI. Don't call this without a real
// planned removal version.
func maybeWarnDeprecated(oldPath, newPath, removalVersion string) {
	noticeID := strings.ReplaceAll(oldPath, " ", ".")

	cfg := config.Load()
	if cfg.CLI.DeprecationNotices == nil {
		cfg.CLI.DeprecationNotices = map[string]string{}
	}

	if last, ok := cfg.CLI.DeprecationNotices[noticeID]; ok {
		if t, err := time.Parse(time.RFC3339, last); err == nil {
			if time.Since(t) < deprecationInterval {
				return
			}
		}
	}

	fmt.Fprintf(os.Stderr,
		"[deprecation] '%s' is deprecated and will be removed in %s — use '%s' instead.\n",
		oldPath, removalVersion, newPath,
	)

	cfg.CLI.DeprecationNotices[noticeID] = time.Now().UTC().Format(time.RFC3339)
	// Ignore save errors — warning throttling is best-effort; failing to
	// persist just means the user sees the same notice next time.
	_ = cfg.Save()
}
