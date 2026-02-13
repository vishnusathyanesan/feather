-- Partial index on deleted_at for soft-delete queries (WHERE deleted_at IS NULL)
CREATE INDEX IF NOT EXISTS idx_messages_not_deleted
    ON messages(channel_id, created_at DESC) WHERE deleted_at IS NULL;

-- Index on file_attachments.user_id for listing user's files
CREATE INDEX IF NOT EXISTS idx_file_attachments_user
    ON file_attachments(user_id);
