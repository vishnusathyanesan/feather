CREATE TABLE messages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    channel_id UUID NOT NULL REFERENCES channels(id) ON DELETE CASCADE,
    user_id UUID NOT NULL REFERENCES users(id),
    parent_id UUID REFERENCES messages(id) ON DELETE SET NULL,
    content TEXT NOT NULL,
    is_alert BOOLEAN NOT NULL DEFAULT false,
    alert_severity VARCHAR(20) DEFAULT NULL,
    alert_metadata JSONB DEFAULT NULL,
    edited_at TIMESTAMPTZ,
    deleted_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_messages_channel_created ON messages(channel_id, created_at DESC);
CREATE INDEX idx_messages_parent ON messages(parent_id) WHERE parent_id IS NOT NULL;
CREATE INDEX idx_messages_user ON messages(user_id);
