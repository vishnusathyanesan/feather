ALTER TABLE channels ALTER COLUMN name DROP NOT NULL;

DROP INDEX IF EXISTS idx_channels_name;
CREATE UNIQUE INDEX idx_channels_name ON channels(name) WHERE name IS NOT NULL AND type NOT IN ('dm', 'group_dm');
