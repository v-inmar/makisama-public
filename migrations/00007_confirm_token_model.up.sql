CREATE TABLE IF NOT EXISTS confirm_token_model (
    id BIGSERIAL PRIMARY KEY,
    hashed TEXT UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '24 hour'),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS user_confirm_model (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT UNIQUE NOT NULL,
    confirm_id BIGINT UNIQUE NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

    -- Shall do this in the next iteration

    -- CONSTRAINT fk_user_confirm_user
    --     FOREIGN KEY (user_id)
    --     REFERENCES user_model(id)
    --     ON DELETE CASCADE,

    -- CONSTRAINT fk_user_confirm_confirm_token_model
    --     FOREIGN KEY (confirm_id)
    --     REFERENCES confirm_token_model(id)
    --     ON DELETE CASCADE
)

