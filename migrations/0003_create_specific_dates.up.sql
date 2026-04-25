CREATE TABLE IF NOT EXISTS specific_dates (
    id BIGSERIAL PRIMARY KEY,
    recurrence_id BIGINT NOT NULL REFERENCES task_recurrences(id) ON DELETE CASCADE,
    scheduled_date DATE NOT NULL,
    UNIQUE (recurrence_id, scheduled_date)
);
