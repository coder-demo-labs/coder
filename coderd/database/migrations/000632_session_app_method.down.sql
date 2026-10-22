-- Flatten the session counts back to app names, adding the counts of an app
-- reported under more than one method.
UPDATE workspace_agent_stats
SET session_counts = (
	SELECT jsonb_object_agg(totals.app_name, totals.sessions)
	FROM (
		SELECT
			sess.app_name,
			SUM(sess.sessions::bigint) AS sessions
		FROM
			jsonb_each(workspace_agent_stats.session_counts) AS m(method, apps),
			jsonb_each_text(m.apps) AS sess(app_name, sessions)
		GROUP BY sess.app_name
	) AS totals
)
WHERE session_counts <> '{}'::jsonb;

COMMENT ON COLUMN workspace_agent_stats.session_counts IS 'Positive session counts keyed by the canonical app name reported by the agent.';

-- Add the minutes of an app across methods, capped at the half hour a bucket
-- covers.
CREATE TEMPORARY TABLE session_apps_without_method ON COMMIT DROP AS
SELECT
	start_time,
	template_id,
	user_id,
	app_name,
	LEAST(SUM(usage_mins), 30)::smallint AS usage_mins
FROM
	template_usage_stats_session_apps
GROUP BY
	start_time, template_id, user_id, app_name;

TRUNCATE template_usage_stats_session_apps;

ALTER TABLE template_usage_stats_session_apps
	DROP CONSTRAINT template_usage_stats_session_apps_pkey,
	DROP CONSTRAINT template_usage_stats_session_apps_method_check,
	DROP COLUMN method,
	ADD CONSTRAINT template_usage_stats_session_apps_pkey
		PRIMARY KEY (start_time, user_id, template_id, app_name);

INSERT INTO template_usage_stats_session_apps (start_time, template_id, user_id, app_name, usage_mins)
SELECT start_time, template_id, user_id, app_name, usage_mins
FROM session_apps_without_method;
