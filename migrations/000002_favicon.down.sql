ALTER TABLE site_settings_history DROP COLUMN IF EXISTS favicon_content_type;

ALTER TABLE site_settings DROP COLUMN IF EXISTS favicon_content_type;
ALTER TABLE site_settings DROP COLUMN IF EXISTS favicon_data;
