-- Add changed_by to track which user made each change

ALTER TABLE pages ADD COLUMN changed_by UUID REFERENCES users(id);
ALTER TABLE pages_history ADD COLUMN changed_by UUID REFERENCES users(id);

ALTER TABLE sections ADD COLUMN changed_by UUID REFERENCES users(id);
ALTER TABLE sections_history ADD COLUMN changed_by UUID REFERENCES users(id);

ALTER TABLE images ADD COLUMN changed_by UUID REFERENCES users(id);
ALTER TABLE images_history ADD COLUMN changed_by UUID REFERENCES users(id);

ALTER TABLE site_settings ADD COLUMN changed_by UUID REFERENCES users(id);
ALTER TABLE site_settings_history ADD COLUMN changed_by UUID REFERENCES users(id);
