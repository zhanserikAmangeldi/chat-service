CREATE TABLE participants (
                              conversation_id BIGINT REFERENCES conversations(id) ON DELETE CASCADE,
                              user_id BIGINT NOT NULL, -- Matches User Service ID type
                              joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
                              PRIMARY KEY (conversation_id, user_id)
);

CREATE INDEX idx_participants_user_id ON participants(user_id);