CREATE TABLE IF NOT EXISTS task_recurrences (
    id BIGSERIAL PRIMARY KEY,
    task_id BIGINT NOT NULL UNIQUE REFERENCES tasks(id) ON DELETE CASCADE,
    recurrence_type TEXT NOT NULL,
    interval_days INT NULL,
    day_of_month INT NULL,
    start_date DATE NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_task_recurrences_type_is_active
    ON task_recurrences(recurrence_type, is_active);
