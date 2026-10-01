CREATE TABLE IF NOT EXISTS board_model (
    id BIGSERIAL PRIMARY KEY,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    title TEXT NOT NULL,
    website TEXT DEFAULT NULL,
    description TEXT DEFAULT NULL,
    slug TEXT UNIQUE NOT NULL,
    owner_id BIGINT NOT NULL,
    CONSTRAINT fk_user FOREIGN KEY (owner_id) REFERENCES user_model (id)
);


CREATE INDEX IF NOT EXISTS idx_board_owner_id ON board_model(owner_id);
