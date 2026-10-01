CREATE TABLE IF NOT EXISTS reset_token_model (
    id BIGSERIAL PRIMARY KEY,
    hashed TEXT UNIQUE NOT NULL,
    is_used BOOLEAN NOT NULL DEFAULT FALSE,
    user_id BIGINT NULL, -- can be null for deleted users
    ttl TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 year'), -- so it can be recycled (1 year maybe), collision is unlikely but might as well
    expires_at TIMESTAMPTZ NOT NULL DEFAULT (NOW() + INTERVAL '1 hour'),
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT fk_user_reset_token FOREIGN KEY (user_id) REFERENCES user_model (id) ON DELETE SET NULL
);