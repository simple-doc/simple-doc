-- pages: replace full unique constraint with partial (only active rows)
ALTER TABLE pages DROP CONSTRAINT pages_section_id_slug_key;
CREATE UNIQUE INDEX pages_section_id_slug_active ON pages(section_id, slug) WHERE deleted = false;
