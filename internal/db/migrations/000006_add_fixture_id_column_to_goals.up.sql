ALTER TABLE goals ADD COLUMN fixture_id INT REFERENCES fixtures(id);