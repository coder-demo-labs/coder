-- Session apps record how the client connected, so two apps that share a name
-- but not a method stay apart. Existing rows predate the method, so derive it
-- from the app name the way the ingest path does.
ALTER TABLE template_usage_stats_session_apps
	ADD COLUMN method connection_log_method DEFAULT 'ssh' NOT NULL;

UPDATE template_usage_stats_session_apps
SET method = app_name::connection_log_method
WHERE app_name IN ('reconnecting_pty', 'port_forwarding');

ALTER TABLE template_usage_stats_session_apps
	ALTER COLUMN method DROP DEFAULT,
	ADD CONSTRAINT template_usage_stats_session_apps_method_check
		CHECK (method IN ('ssh', 'reconnecting_pty', 'port_forwarding'));

ALTER TABLE template_usage_stats_session_apps
	DROP CONSTRAINT template_usage_stats_session_apps_pkey,
	ADD CONSTRAINT template_usage_stats_session_apps_pkey
		PRIMARY KEY (start_time, user_id, template_id, method, app_name);

COMMENT ON COLUMN template_usage_stats_session_apps.method IS 'How the client connected. Together with app_name it identifies a session app.';

-- Regroup the session counts the rollup and operational statistics still read
-- under their method, deriving it from the app name.
UPDATE workspace_agent_stats
SET session_counts = (
	SELECT jsonb_object_agg(grouped.method, grouped.apps)
	FROM (
		SELECT
			CASE
				WHEN sess.app_name IN ('reconnecting_pty', 'port_forwarding') THEN sess.app_name
				ELSE 'ssh'
			END AS method,
			jsonb_object_agg(sess.app_name, sess.sessions) AS apps
		FROM jsonb_each(workspace_agent_stats.session_counts) AS sess(app_name, sessions)
		GROUP BY 1
	) AS grouped
)
WHERE session_counts <> '{}'::jsonb;

COMMENT ON COLUMN workspace_agent_stats.session_counts IS 'Positive session counts keyed by connection method, then by the canonical app name reported by the agent.';
