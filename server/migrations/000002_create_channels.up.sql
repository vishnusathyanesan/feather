CREATE TYPE channel_type AS ENUM ('public', 'private', 'system');

CREATE TABLE channels (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL,
    topic VARCHAR(500) DEFAULT '',
    description TEXT DEFAULT '',
    type channel_type NOT NULL DEFAULT 'public',
    is_readonly BOOLEAN NOT NULL DEFAULT false,
    creator_id UUID REFERENCES users(id),
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE UNIQUE INDEX idx_channels_name ON channels(name);
