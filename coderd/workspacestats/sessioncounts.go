package workspacestats

import (
	"cmp"
	"maps"
	"slices"

	agentproto "github.com/coder/coder/v2/agent/proto"
	"github.com/coder/coder/v2/codersdk"
)

// deprecatedSessionCounts returns the fixed session_count_* fields keyed by
// normalized app name. Agents predating API v2.11 set these instead of
// session_counts, never both.
func deprecatedSessionCounts(st *agentproto.Stats) map[string]int64 {
	return map[string]int64{
		string(codersdk.AppFamilyVSCode):          st.SessionCountVscode,
		string(codersdk.AppFamilyJetBrains):       st.SessionCountJetbrains,
		string(codersdk.AppFamilyReconnectingPTY): st.SessionCountReconnectingPty,
		string(codersdk.AppFamilySSH):             st.SessionCountSsh,
	}
}

// normalizedSessionCounts returns per-app session counts, names normalized and
// non-positive counts dropped.
func normalizedSessionCounts(st *agentproto.Stats) map[string]int64 {
	reported := st.GetSessionCounts()
	if len(reported) == 0 {
		reported = deprecatedSessionCounts(st)
	}
	counts := make(map[string]int64, len(reported))
	for app, count := range reported {
		if count <= 0 {
			continue
		}
		name := codersdk.NormalizeAppName(app)
		// Reserved for capSessionCounts.
		if name == codersdk.AppNameOverflow {
			name = string(codersdk.AppFamilyUnknown)
		}
		counts[name] += count
	}
	return counts
}

// maxSessionCountEntries bounds distinct app names per stats report. Overflow
// aggregates under codersdk.AppNameOverflow.
const maxSessionCountEntries = 64

// capSessionCounts keeps the busiest maxSessionCountEntries normalized names,
// preferring known apps, and sums the rest into codersdk.AppNameOverflow, so the
// result can hold one name past the cap. The second return counts the names
// summed away.
func capSessionCounts(counts map[string]int64) (map[string]int64, int) {
	if len(counts) <= maxSessionCountEntries {
		return counts, 0
	}
	// Names are already normalized, so look up families directly rather
	// than normalizing again on every sort comparison.
	families := codersdk.SessionCountAppFamilies()
	rank := func(name string) int {
		if family, ok := families[name]; ok && family != codersdk.AppFamilyUnknown {
			return 0
		}
		return 1
	}
	ranked := slices.SortedFunc(maps.Keys(counts), func(a, b string) int {
		return cmp.Or(
			cmp.Compare(rank(a), rank(b)),     // known apps first
			cmp.Compare(counts[b], counts[a]), // then busiest
			cmp.Compare(a, b),                 // then name, so the cut is stable
		)
	})
	capped := make(map[string]int64, maxSessionCountEntries+1)
	for _, name := range ranked[:maxSessionCountEntries] {
		capped[name] = counts[name]
	}
	overflow := ranked[maxSessionCountEntries:]
	for _, name := range overflow {
		capped[codersdk.AppNameOverflow] += counts[name]
	}
	return capped, len(overflow)
}

// sessionAppMethod derives how a session app connected from its normalized
// name. Every current reporter uses a name that matches its method, and
// anything else came in over SSH.
func sessionAppMethod(appName string) codersdk.ConnectionLogMethod {
	switch codersdk.AppFamilyName(appName) {
	case codersdk.AppFamilyReconnectingPTY:
		return codersdk.ConnectionLogMethodReconnectingPTY
	case codersdk.AppFamilyPortForwarding:
		return codersdk.ConnectionLogMethodPortForwarding
	default:
		return codersdk.ConnectionLogMethodSSH
	}
}

// methodSessionCounts returns the per-app session counts grouped by connection
// method, the shape stored in workspace_agent_stats.session_counts. Each
// method is capped on its own, so one noisy method cannot push another
// method's apps into overflow. The second return counts the names summed
// away.
func methodSessionCounts(counts map[string]int64) (map[string]map[string]int64, int) {
	byMethod := make(map[string]map[string]int64)
	for app, count := range counts {
		method := string(sessionAppMethod(app))
		if byMethod[method] == nil {
			byMethod[method] = make(map[string]int64)
		}
		byMethod[method][app] = count
	}
	var overflow int
	for method, apps := range byMethod {
		capped, n := capSessionCounts(apps)
		byMethod[method] = capped
		overflow += n
	}
	return byMethod, overflow
}

// HasSessionCounts reports whether the stats contain any active session.
func HasSessionCounts(st *agentproto.Stats) bool {
	return len(normalizedSessionCounts(st)) > 0
}

// ClearSessionCounts zeroes every session count on the given stats.
func ClearSessionCounts(st *agentproto.Stats) {
	st.SessionCounts = nil
	st.SessionCountVscode = 0
	st.SessionCountJetbrains = 0
	st.SessionCountReconnectingPty = 0
	st.SessionCountSsh = 0
}
