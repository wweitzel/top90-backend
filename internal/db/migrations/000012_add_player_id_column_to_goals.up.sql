ALTER TABLE goals ADD COLUMN IF NOT EXISTS player_id INT REFERENCES players(id);
