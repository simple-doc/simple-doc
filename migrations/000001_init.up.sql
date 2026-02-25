-- =============================================================================
-- Consolidated initial schema
-- =============================================================================

-- Users & Auth
-- =============================================================================

CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    company TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL UNIQUE,
    password TEXT NOT NULL,
    last_login TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE roles (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name TEXT NOT NULL UNIQUE,
    description TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1
);

CREATE TABLE user_roles (
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    PRIMARY KEY (user_id, role_id)
);

CREATE INDEX idx_user_roles_user_id ON user_roles(user_id);
CREATE INDEX idx_user_roles_role_id ON user_roles(role_id);

CREATE TABLE sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    preview_roles TEXT
);

CREATE INDEX idx_sessions_token ON sessions(token);
CREATE INDEX idx_sessions_expires_at ON sessions(expires_at);

CREATE TABLE password_reset_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX idx_password_reset_tokens_user_id ON password_reset_tokens(user_id);
CREATE INDEX idx_password_reset_tokens_token ON password_reset_tokens(token);

CREATE TABLE login_log (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address TEXT NOT NULL,
    user_agent TEXT NOT NULL DEFAULT '',
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE INDEX login_log_user_id_idx ON login_log(user_id);
CREATE INDEX login_log_created_at_idx ON login_log(created_at DESC);

-- Content
-- =============================================================================

CREATE TABLE section_rows (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    version INT NOT NULL DEFAULT 1,
    changed_by UUID REFERENCES users(id),
    deleted BOOLEAN NOT NULL DEFAULT false,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE sections (
    id TEXT PRIMARY KEY DEFAULT gen_random_uuid()::text,
    name TEXT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1,
    icon TEXT NOT NULL DEFAULT 'document',
    changed_by UUID REFERENCES users(id),
    required_role TEXT,
    deleted BOOLEAN NOT NULL DEFAULT false,
    row_id UUID REFERENCES section_rows(id) ON DELETE SET NULL
);

CREATE UNIQUE INDEX sections_name_active ON sections(name) WHERE deleted = false;

CREATE TABLE pages (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id TEXT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    content_md TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1,
    changed_by UUID REFERENCES users(id),
    deleted BOOLEAN NOT NULL DEFAULT false,
    parent_slug TEXT
);

CREATE INDEX idx_pages_section_id ON pages(section_id);
CREATE UNIQUE INDEX pages_section_id_slug_active ON pages(section_id, slug) WHERE deleted = false;

CREATE TABLE images (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    filename TEXT NOT NULL UNIQUE,
    content_type TEXT NOT NULL DEFAULT 'image/svg+xml',
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    section_id TEXT,
    changed_by UUID REFERENCES users(id)
);

-- Site Settings
-- =============================================================================

CREATE TABLE site_settings (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    site_title TEXT NOT NULL DEFAULT 'SolarFlux Documentation',
    badge TEXT NOT NULL DEFAULT 'API Documentation',
    heading TEXT NOT NULL DEFAULT 'SolarFlux API Docs',
    description TEXT NOT NULL DEFAULT 'Technical documentation for the SolarFlux space weather monitoring platform.',
    footer TEXT NOT NULL DEFAULT 'SolarFlux Platform',
    version INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    theme TEXT NOT NULL DEFAULT 'midnight',
    accent_color TEXT NOT NULL DEFAULT 'blue',
    changed_by UUID REFERENCES users(id),
    singleton BOOLEAN NOT NULL DEFAULT true,
    CONSTRAINT site_settings_singleton_check CHECK (singleton = true)
);

CREATE UNIQUE INDEX site_settings_singleton_unique ON site_settings(singleton);

-- History Tables
-- =============================================================================

CREATE TABLE sections_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    section_id TEXT NOT NULL REFERENCES sections(id) ON DELETE CASCADE,
    version INT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    icon TEXT NOT NULL DEFAULT 'document',
    changed_by UUID REFERENCES users(id),
    required_role TEXT,
    row_id UUID,
    UNIQUE (section_id, version)
);

CREATE TABLE pages_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    page_id UUID NOT NULL REFERENCES pages(id) ON DELETE CASCADE,
    version INT NOT NULL,
    section_id TEXT NOT NULL,
    slug TEXT NOT NULL,
    title TEXT NOT NULL,
    content_md TEXT NOT NULL,
    sort_order INT NOT NULL DEFAULT 0,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE (page_id, version)
);

CREATE TABLE images_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    image_id UUID NOT NULL REFERENCES images(id) ON DELETE CASCADE,
    version INT NOT NULL,
    filename TEXT NOT NULL,
    content_type TEXT NOT NULL,
    data BYTEA NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE (image_id, version)
);

CREATE TABLE site_settings_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    version INT NOT NULL,
    site_title TEXT NOT NULL,
    badge TEXT NOT NULL,
    heading TEXT NOT NULL,
    description TEXT NOT NULL,
    footer TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    theme TEXT NOT NULL DEFAULT 'midnight',
    accent_color TEXT NOT NULL DEFAULT 'blue',
    changed_by UUID REFERENCES users(id)
);

CREATE TABLE users_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    version INT NOT NULL,
    firstname TEXT NOT NULL,
    lastname TEXT NOT NULL,
    company TEXT NOT NULL DEFAULT '',
    email TEXT NOT NULL,
    roles TEXT NOT NULL DEFAULT '',
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE (user_id, version)
);

CREATE TABLE roles_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    role_id UUID NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    version INT NOT NULL,
    name TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE (role_id, version)
);

CREATE TABLE section_rows_history (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    row_id UUID NOT NULL REFERENCES section_rows(id) ON DELETE CASCADE,
    version INT NOT NULL,
    title TEXT NOT NULL,
    description TEXT NOT NULL DEFAULT '',
    sort_order INT NOT NULL DEFAULT 0,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    changed_by UUID REFERENCES users(id),
    UNIQUE (row_id, version)
);
