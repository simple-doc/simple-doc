ALTER TABLE site_settings ADD COLUMN theme TEXT NOT NULL DEFAULT 'midnight';
ALTER TABLE site_settings ADD COLUMN accent_color TEXT NOT NULL DEFAULT 'blue';

ALTER TABLE site_settings_history ADD COLUMN theme TEXT NOT NULL DEFAULT 'midnight';
ALTER TABLE site_settings_history ADD COLUMN accent_color TEXT NOT NULL DEFAULT 'blue';
