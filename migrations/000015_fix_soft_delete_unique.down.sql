DROP INDEX pages_section_id_slug_active;
ALTER TABLE pages ADD CONSTRAINT pages_section_id_slug_key UNIQUE(section_id, slug);
