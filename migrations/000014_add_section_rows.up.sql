CREATE TABLE section_rows (
    id          SERIAL PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    version     INT NOT NULL DEFAULT 1,
    changed_by  UUID REFERENCES users(id),
    deleted     BOOLEAN NOT NULL DEFAULT FALSE,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE section_rows_history (
    id           SERIAL PRIMARY KEY,
    row_id       INT NOT NULL REFERENCES section_rows(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    sort_order   INT NOT NULL DEFAULT 0,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by   UUID REFERENCES users(id),
    UNIQUE(row_id, version)
);

ALTER TABLE sections ADD COLUMN row_id INT REFERENCES section_rows(id) ON DELETE SET NULL;
ALTER TABLE sections_history ADD COLUMN row_id INT;
