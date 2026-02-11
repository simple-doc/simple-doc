CREATE TABLE users_history (
    id         SERIAL PRIMARY KEY,
    user_id    UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version    INT NOT NULL,
    firstname  TEXT NOT NULL,
    lastname   TEXT NOT NULL,
    company    TEXT NOT NULL DEFAULT '',
    email      TEXT NOT NULL,
    roles      TEXT NOT NULL DEFAULT '',
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE(user_id, version)
);

CREATE TABLE roles_history (
    id          SERIAL PRIMARY KEY,
    role_id     UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    version     INT NOT NULL,
    name        TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    changed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by  UUID REFERENCES users(id),
    UNIQUE(role_id, version)
);

ALTER TABLE users ADD COLUMN version INT NOT NULL DEFAULT 1;
ALTER TABLE roles ADD COLUMN version INT NOT NULL DEFAULT 1;
