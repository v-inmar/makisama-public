CREATE TABLE IF NOT EXISTS report_board_model (
    id BIGSERIAL PRIMARY KEY,
    report_id TEXT UNIQUE NOT NULL,
    reason TEXT NOT NULL,
    comment TEXT DEFAULT NULL,
    email TEXT DEFAULT NULL,
    board_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

    -- this is because boards can be deleted and this report must stay
    -- CONSTRAINT fk_report_board FOREIGN KEY (board_id) REFERENCES board_model (id)
);


CREATE TABLE IF NOT EXISTS report_feedback_model (
    id BIGSERIAL PRIMARY KEY,
    report_id TEXT UNIQUE NOT NULL,
    reason TEXT NOT NULL,
    comment TEXT DEFAULT NULL,
    email TEXT DEFAULT NULL,
    feedback_id BIGINT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP

    -- this is because feedback can be deleted(soft delete) and this report must stay
    -- CONSTRAINT fk_report_feedback FOREIGN KEY (feedback_id) REFERENCES feedback_model (id)
);

