ALTER TABLE mentions DROP CONSTRAINT IF EXISTS mentions_mentioned_group_id_fkey;
ALTER TABLE mentions ADD CONSTRAINT mentions_mentioned_group_id_fkey
    FOREIGN KEY (mentioned_group_id) REFERENCES user_groups(id);
