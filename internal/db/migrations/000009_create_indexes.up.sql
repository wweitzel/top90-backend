CREATE INDEX IF NOT EXISTS goals_reddit_post_created_at ON goals(reddit_post_created_at);
CREATE INDEX IF NOT EXISTS goals_fixture_id_reddit_post_created_at ON goals(fixture_id, reddit_post_created_at);

CREATE INDEX IF NOT EXISTS fixtures_date ON fixtures(date);
CREATE INDEX IF NOT EXISTS fixtures_league_id ON fixtures(league_id);
CREATE INDEX IF NOT EXISTS fixtures_season ON fixtures(season);
CREATE INDEX IF NOT EXISTS fixtures_league_id_season ON fixtures(league_id, season);
