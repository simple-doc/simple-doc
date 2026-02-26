ALTER TABLE site_settings ADD COLUMN favicon_data BYTEA;
ALTER TABLE site_settings ADD COLUMN favicon_content_type TEXT;

ALTER TABLE site_settings_history ADD COLUMN favicon_content_type TEXT;
