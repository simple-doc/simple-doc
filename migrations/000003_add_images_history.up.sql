ALTER TABLE images ADD COLUMN version INT NOT NULL DEFAULT 1;
ALTER TABLE images ADD COLUMN updated_at TIMESTAMPTZ NOT NULL DEFAULT now();

CREATE TABLE images_history (
    id           SERIAL PRIMARY KEY,
    image_id     INT NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    version      INT NOT NULL,
    filename     TEXT NOT NULL,
    content_type TEXT NOT NULL,
    data         BYTEA NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL,
    changed_at   TIMESTAMPTZ NOT NULL DEFAULT now(),
    UNIQUE(image_id, version)
);
