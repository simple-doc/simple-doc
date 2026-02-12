CREATE TABLE IF NOT EXISTS site_settings (
    id         INT PRIMARY KEY DEFAULT 1 CHECK (id = 1),
    site_title TEXT NOT NULL DEFAULT 'SolarFlux Documentation',
    badge      TEXT NOT NULL DEFAULT 'API Documentation',
    heading    TEXT NOT NULL DEFAULT 'SolarFlux API Docs',
    description TEXT NOT NULL DEFAULT 'Technical documentation for the SolarFlux space weather monitoring platform.',
    footer     TEXT NOT NULL DEFAULT 'SolarFlux Platform',
    version    INT NOT NULL DEFAULT 1,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now()
);

CREATE TABLE IF NOT EXISTS site_settings_history (
    id         SERIAL PRIMARY KEY,
    version    INT NOT NULL,
    site_title TEXT NOT NULL,
    badge      TEXT NOT NULL,
    heading    TEXT NOT NULL,
    description TEXT NOT NULL,
    footer     TEXT NOT NULL,
    changed_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
