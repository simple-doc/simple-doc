ALTER TABLE images ADD COLUMN section_id TEXT REFERENCES sections(id);
