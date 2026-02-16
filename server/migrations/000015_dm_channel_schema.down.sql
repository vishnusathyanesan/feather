DROP INDEX IF EXISTS idx_channels_name;
CREATE UNIQUE INDEX idx_channels_name ON channels(name);
ALTER TABLE channels ALTER COLUMN name SET NOT NULL;
