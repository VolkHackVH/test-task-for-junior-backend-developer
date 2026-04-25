ALTER TABLE tasks
    ADD COLUMN scheduled_for DATE NULL,
    ADD COLUMN source_task_id BIGINT NULL REFERENCES tasks(id) ON DELETE CASCADE;

CREATE UNIQUE INDEX IF NOT EXISTS ux_tasks_source_task_scheduled_for
    ON tasks (source_task_id, scheduled_for)
    WHERE source_task_id IS NOT NULL AND scheduled_for IS NOT NULL;
