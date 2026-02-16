CREATE TYPE call_status AS ENUM ('ringing', 'in_progress', 'ended', 'missed', 'declined');
CREATE TYPE call_type AS ENUM ('audio', 'video');

CREATE TABLE calls (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id),
    initiator_id UUID NOT NULL REFERENCES users(id),
    call_type call_type NOT NULL DEFAULT 'audio',
    status call_status NOT NULL DEFAULT 'ringing',
    started_at TIMESTAMPTZ,
    ended_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE call_participants (
    call_id UUID NOT NULL REFERENCES calls(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    joined_at TIMESTAMPTZ,
    left_at TIMESTAMPTZ,
    PRIMARY KEY (call_id, user_id)
);

CREATE INDEX idx_calls_channel ON calls(channel_id);
CREATE INDEX idx_calls_status ON calls(status) WHERE status IN ('ringing', 'in_progress');
