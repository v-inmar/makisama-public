CREATE TABLE IF NOT EXISTS feedback_model (
    id BIGSERIAL PRIMARY KEY,
    title TEXT NOT NULL,
    description TEXT DEFAULT NULL, 
    slug TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ DEFAULT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    board_id BIGINT NOT NULL,
    owner_id BIGINT DEFAULT NULL,
    email TEXT DEFAULT NULL,
    scope TEXT DEFAULT NULL,
    resolution TEXT DEFAULT NULL,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL
    

    -- removed because when board is deleted, feedbacks needs to stay, feedback owners can be different from board owner
    -- board owner cannot delete feedback by deleting the board
    -- board owner must delete a feedback by reviewing them manually
    -- CONSTRAINT fk_board FOREIGN KEY (board_id) REFERENCES board_model (id),
    -- CONSTRAINT fk_owner FOREIGN KEY (owner_id) REFERENCES user_model (id)
);

CREATE INDEX IF NOT EXISTS idx_feedback_expires_at ON feedback_model (expires_at);