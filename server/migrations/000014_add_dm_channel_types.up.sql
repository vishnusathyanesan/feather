ALTER TYPE channel_type ADD VALUE IF NOT EXISTS 'dm';
ALTER TYPE channel_type ADD VALUE IF NOT EXISTS 'group_dm';

ALTER TABLE channels ALTER COLUMN name DROP NOT NULL;

DROP INDEX IF EXISTS idx_channels_name;
CREATE UNIQUE INDEX idx_channels_name ON channels(name) WHERE name IS NOT NULL AND type NOT IN ('dm', 'group_dm');
