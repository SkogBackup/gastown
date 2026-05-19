package daemon

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/steveyegge/gastown/internal/beads"
)

// bdReadOnlyEnv returns an environment slice for read-only bd/gt subprocess
// calls invoked by the daemon. It forces BD_DOLT_AUTO_COMMIT=off so that
// read-only operations (status checks, list, show) do not request a Dolt
// auto-commit on completion. Without this, every read-only call opens a
// fresh connection to attempt a no-op commit, producing thousands of
// failed-but-counted connections per hour on idle towns and spamming
// dolt.log. See gh#3596.
//
// Existing BD_DOLT_AUTO_COMMIT entries are filtered out before appending
// the authoritative "off" value, because glibc getenv() returns the first
// matching entry — a stale "on" earlier in the slice would otherwise win.
func bdReadOnlyEnv() []string {
	return forceBDReadOnly(beads.BuildRoutingBDEnv(os.Environ(), ""))
}

func bdReadOnlyRoutingEnv(townRoot string) []string {
	fallback := ""
	if townRoot != "" {
		fallback = filepath.Join(townRoot, ".beads")
	}
	return forceBDReadOnly(beads.BuildRoutingBDEnv(os.Environ(), fallback))
}

func bdMutationRoutingEnv(townRoot string) []string {
	fallback := ""
	if townRoot != "" {
		fallback = filepath.Join(townRoot, ".beads")
	}
	env := beads.BuildRoutingBDEnv(os.Environ(), fallback)
	return filterDaemonEnvKeys(env, "BD_READONLY")
}

func bdReadOnlyPinnedEnv(beadsDir string) []string {
	return forceBDReadOnly(beads.BuildPinnedBDEnv(os.Environ(), beadsDir))
}

func forceBDReadOnly(base []string) []string {
	filtered := filterDaemonEnvKeys(base, "BD_DOLT_AUTO_COMMIT", "BD_READONLY")
	return append(filtered, "BD_DOLT_AUTO_COMMIT=off", "BD_READONLY=true")
}

func filterDaemonEnvKeys(base []string, keys ...string) []string {
	filtered := make([]string, 0, len(base)+1)
	for _, e := range base {
		skip := false
		for _, key := range keys {
			if strings.HasPrefix(e, key+"=") {
				skip = true
				break
			}
		}
		if !skip {
			filtered = append(filtered, e)
		}
	}
	return filtered
}
