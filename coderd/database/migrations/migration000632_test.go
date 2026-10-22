package migrations_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/coder/coder/v2/testutil"
)

// TestMigration000632SessionAppMethod covers the method derived for rows that
// predate it, the regrouping of raw session counts, and the way the down
// migration folds methods back together.
//
//nolint:tparallel,paralleltest // Subtests share one database with transaction-local fixtures.
func TestMigration000632SessionAppMethod(t *testing.T) {
	t.Parallel()

	sqlDB := testSQLDB(t)
	stepTo(t, sqlDB, 631)

	ctx := testutil.Context(t, testutil.WaitSuperLong)
	upSQL, err := os.ReadFile("000632_session_app_method.up.sql")
	require.NoError(t, err)
	downSQL, err := os.ReadFile("000632_session_app_method.down.sql")
	require.NoError(t, err)

	tx, err := sqlDB.BeginTx(ctx, nil)
	require.NoError(t, err)
	t.Cleanup(func() { _ = tx.Rollback() })

	_, err = tx.ExecContext(ctx, `
		INSERT INTO template_usage_stats (
			start_time, end_time, template_id, user_id, median_latency_ms,
			usage_mins, app_usage_mins
		) VALUES (
			date_trunc('hour', statement_timestamp()),
			date_trunc('hour', statement_timestamp()) + interval '30 minutes',
			'22222222-2222-2222-2222-222222222222'::uuid,
			'11111111-1111-1111-1111-111111111111'::uuid,
			NULL, 30, NULL
		)
	`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO template_usage_stats_session_apps (
			start_time, template_id, user_id, app_name, usage_mins
		) SELECT
			date_trunc('hour', statement_timestamp()),
			'22222222-2222-2222-2222-222222222222'::uuid,
			'11111111-1111-1111-1111-111111111111'::uuid,
			app_name, usage_mins
		FROM (VALUES
			('vscode', 10), ('ssh', 5), ('reconnecting_pty', 7), ('port_forwarding', 3)
		) AS apps(app_name, usage_mins)
	`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO workspace_agent_stats (
			id, created_at, user_id, agent_id, workspace_id, template_id,
			connections_by_proto, connection_count, rx_packets, rx_bytes,
			tx_packets, tx_bytes, connection_median_latency_ms, session_counts
		) VALUES
			(gen_random_uuid(), statement_timestamp(), gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), gen_random_uuid(),
				'{}', 0, 0, 0, 0, 0, 0,
				'{"vscode": 2, "reconnecting_pty": 1, "port_forwarding": 4}'),
			(gen_random_uuid(), statement_timestamp(), gen_random_uuid(), gen_random_uuid(), gen_random_uuid(), gen_random_uuid(),
				'{}', 0, 0, 0, 0, 0, 0, '{}')
	`)
	require.NoError(t, err)

	appMethods := func() map[string]string {
		rows, err := tx.QueryContext(ctx, `SELECT app_name, method::text FROM template_usage_stats_session_apps`)
		require.NoError(t, err)
		defer rows.Close()
		got := map[string]string{}
		for rows.Next() {
			var app, method string
			require.NoError(t, rows.Scan(&app, &method))
			got[app] = method
		}
		require.NoError(t, rows.Err())
		return got
	}
	storedCounts := func() []string {
		rows, err := tx.QueryContext(ctx, `SELECT session_counts::text FROM workspace_agent_stats ORDER BY session_counts::text`)
		require.NoError(t, err)
		defer rows.Close()
		var got []string
		for rows.Next() {
			var counts string
			require.NoError(t, rows.Scan(&counts))
			got = append(got, counts)
		}
		require.NoError(t, rows.Err())
		return got
	}

	_, err = tx.ExecContext(ctx, string(upSQL))
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"vscode":           "ssh",
		"ssh":              "ssh",
		"reconnecting_pty": "reconnecting_pty",
		"port_forwarding":  "port_forwarding",
	}, appMethods())
	require.JSONEq(t, `{}`, storedCounts()[0])
	require.JSONEq(t,
		`{"ssh": {"vscode": 2}, "reconnecting_pty": {"reconnecting_pty": 1}, "port_forwarding": {"port_forwarding": 4}}`,
		storedCounts()[1])

	// The method joins the primary key and is limited to the session methods.
	_, err = tx.ExecContext(ctx, `SAVEPOINT before_methods`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO template_usage_stats_session_apps (
			start_time, template_id, user_id, method, app_name, usage_mins
		) SELECT start_time, template_id, user_id, 'port_forwarding', 'vscode', 2
		FROM template_usage_stats_session_apps WHERE app_name = 'vscode'
	`)
	require.NoError(t, err)
	_, err = tx.ExecContext(ctx, `
		INSERT INTO template_usage_stats_session_apps (
			start_time, template_id, user_id, method, app_name, usage_mins
		) SELECT start_time, template_id, user_id, 'workspace_app', 'x', 1
		FROM template_usage_stats_session_apps WHERE app_name = 'vscode' AND method = 'ssh'
	`)
	require.ErrorContains(t, err, "template_usage_stats_session_apps_method_check")
	_, err = tx.ExecContext(ctx, `ROLLBACK TO SAVEPOINT before_methods`)
	require.NoError(t, err)

	// Add the same app under a second method: the down migration adds the
	// minutes, capped at the half hour a bucket covers.
	_, err = tx.ExecContext(ctx, `
		INSERT INTO template_usage_stats_session_apps (
			start_time, template_id, user_id, method, app_name, usage_mins
		) SELECT start_time, template_id, user_id, 'port_forwarding', 'vscode', 25
		FROM template_usage_stats_session_apps WHERE app_name = 'vscode'
	`)
	require.NoError(t, err)

	_, err = tx.ExecContext(ctx, string(downSQL))
	require.NoError(t, err)
	rows, err := tx.QueryContext(ctx, `SELECT app_name, usage_mins FROM template_usage_stats_session_apps ORDER BY app_name`)
	require.NoError(t, err)
	defer rows.Close()
	got := map[string]int64{}
	for rows.Next() {
		var app string
		var mins int64
		require.NoError(t, rows.Scan(&app, &mins))
		got[app] = mins
	}
	require.NoError(t, rows.Err())
	require.Equal(t, map[string]int64{"vscode": 30, "ssh": 5, "reconnecting_pty": 7, "port_forwarding": 3}, got)
	require.JSONEq(t, `{}`, storedCounts()[0])
	require.JSONEq(t, `{"vscode": 2, "reconnecting_pty": 1, "port_forwarding": 4}`, storedCounts()[1])
}
