ALTER TABLE site_settings_history DROP COLUMN changed_by;
ALTER TABLE site_settings DROP COLUMN changed_by;

ALTER TABLE images_history DROP COLUMN changed_by;
ALTER TABLE images DROP COLUMN changed_by;

ALTER TABLE sections_history DROP COLUMN changed_by;
ALTER TABLE sections DROP COLUMN changed_by;

ALTER TABLE pages_history DROP COLUMN changed_by;
ALTER TABLE pages DROP COLUMN changed_by;
