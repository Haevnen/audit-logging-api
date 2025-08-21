-- 1. Define ENUM types first
CREATE TYPE async_task_status AS ENUM ('pending', 'running', 'succeeded', 'failed');
CREATE TYPE async_task_type AS ENUM ('log_cleanup', 'archive', 'export', 'reindex');

-- 2. Create table using ENUMs
CREATE TABLE async_tasks (
    task_id      UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    status       async_task_status NOT NULL,
    task_type    async_task_type   NOT NULL,
    payload      JSONB,
    created_at   TIMESTAMPTZ DEFAULT NOW(),
    updated_at   TIMESTAMPTZ DEFAULT NOW(),
    tenant_uid   TEXT,
    user_id      TEXT NOT NULL,
    error_msg    TEXT
);