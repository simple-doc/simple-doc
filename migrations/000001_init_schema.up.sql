CREATE TABLE sections (
    id          TEXT PRIMARY KEY,
    title       TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE pages (
    id          SERIAL PRIMARY KEY,
    section_id  TEXT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    slug        TEXT NOT NULL,
    title       TEXT NOT NULL,
    content_md  TEXT NOT NULL,
    sort_order  INT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(section_id, slug)
);

CREATE INDEX idx_pages_section_id ON pages(section_id);

CREATE TABLE images (
    id           SERIAL PRIMARY KEY,
    filename     TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL DEFAULT 'image/svg+xml',
    data         BYTEA NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT now()
);
