BEGIN;

-- ============================================================
-- Reverse: copy name back to id, drop name column
-- ============================================================

-- Build mapping: current uuid id -> name (which was the old id)
CREATE TEMP TABLE section_id_map (
    new_id TEXT PRIMARY KEY,
    old_id TEXT NOT NULL
);
INSERT INTO section_id_map (new_id, old_id)
    SELECT id, name FROM sections;

-- Drop FK constraints
ALTER TABLE pages DROP CONSTRAINT pages_section_id_fkey;
ALTER TABLE sections_history DROP CONSTRAINT sections_history_section_id_fkey;
ALTER TABLE sections_history DROP CONSTRAINT sections_history_section_id_version_key;
ALTER TABLE images DROP CONSTRAINT images_section_id_fkey;
DROP INDEX IF EXISTS pages_section_id_slug_active;

-- Revert FK columns
UPDATE pages p
   SET section_id = m.old_id
  FROM section_id_map m
 WHERE p.section_id = m.new_id;

UPDATE images i
   SET section_id = m.old_id
  FROM section_id_map m
 WHERE i.section_id = m.new_id;

UPDATE sections_history sh
   SET section_id = m.old_id
  FROM section_id_map m
 WHERE sh.section_id = m.new_id;

UPDATE pages_history ph
   SET section_id = m.old_id
  FROM section_id_map m
 WHERE ph.section_id = m.new_id;

-- Revert sections.id
UPDATE sections s
   SET id = m.old_id
  FROM section_id_map m
 WHERE s.id = m.new_id;

-- Remove default and name column
ALTER TABLE sections ALTER COLUMN id DROP DEFAULT;
DROP INDEX IF EXISTS sections_name_active;
ALTER TABLE sections DROP COLUMN name;

-- Re-add constraints
CREATE UNIQUE INDEX pages_section_id_slug_active ON pages(section_id, slug) WHERE deleted = false;

ALTER TABLE pages ADD CONSTRAINT pages_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE;

ALTER TABLE sections_history ADD CONSTRAINT sections_history_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE;
ALTER TABLE sections_history ADD CONSTRAINT sections_history_section_id_version_key
    UNIQUE (section_id, version);

ALTER TABLE images ADD CONSTRAINT images_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id);

DROP TABLE section_id_map;

COMMIT;
