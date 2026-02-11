ALTER TABLE sections ADD COLUMN version INT NOT NULL DEFAULT 1;
ALTER TABLE pages ADD COLUMN version INT NOT NULL DEFAULT 1;

CREATE TABLE sections_history (
    id           SERIAL PRIMARY KEY,
    section_id   TEXT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    title        TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    sort_order   INT NOT NULL DEFAULT 0,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(section_id, version)
);

CREATE TABLE pages_history (
    id           SERIAL PRIMARY KEY,
    page_id      INT NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    section_id   TEXT NOT NULL,
    slug         TEXT NOT NULL,
    title        TEXT NOT NULL,
    content_md   TEXT NOT NULL,
    sort_order   INT NOT NULL DEFAULT 0,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(page_id, version)
);
