CREATE TABLE IF NOT EXISTS jobs_queue
(
    -- Job specification.
    id              TEXT      NOT NULL PRIMARY KEY,
    type            TEXT      NOT NULL,
    run_at          TIMESTAMP NOT NULL,
    payload         bytea     NOT NULL,
    created_at      TIMESTAMP NOT NULL DEFAULT current_timestamp,
    updated_at      TIMESTAMP NOT NULL DEFAULT current_timestamp,

    -- Result generated by execution.
    attempts_done   INTEGER   NOT NULL DEFAULT 0,
    last_attempt_at TIMESTAMP,
    last_error      TEXT
);


CREATE INDEX IF NOT EXISTS idx_jobs_queue_type ON jobs_queue (type);
CREATE INDEX IF NOT EXISTS idx_jobs_queue_run_at ON jobs_queue (run_at);
