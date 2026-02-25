BEGIN;

-- ============================================================
-- 1. Add `name` column, populate from current text id
-- ============================================================
ALTER TABLE sections ADD COLUMN name TEXT;
UPDATE sections SET name = id;
ALTER TABLE sections ALTER COLUMN name SET NOT NULL;

-- ============================================================
-- 2. Build old_id -> new_uuid mapping
-- ============================================================
CREATE TEMP TABLE section_id_map (
    old_id TEXT PRIMARY KEY,
    new_id TEXT NOT NULL DEFAULT gen_random_uuid()::text
);
INSERT INTO section_id_map (old_id)
    SELECT id FROM sections;

-- ============================================================
-- 3. Drop FK constraints referencing sections(id)
-- ============================================================
ALTER TABLE pages DROP CONSTRAINT pages_section_id_fkey;
ALTER TABLE sections_history DROP CONSTRAINT sections_history_section_id_fkey;
ALTER TABLE sections_history DROP CONSTRAINT sections_history_section_id_version_key;
ALTER TABLE images DROP CONSTRAINT images_section_id_fkey;
-- pages_history.section_id has no FK, just data

-- Drop partial unique index on pages(section_id, slug)
DROP INDEX IF EXISTS pages_section_id_slug_active;

-- ============================================================
-- 4. Update FK references to use new UUIDs
-- ============================================================
UPDATE pages p
   SET section_id = m.new_id
  FROM section_id_map m
 WHERE p.section_id = m.old_id;

UPDATE images i
   SET section_id = m.new_id
  FROM section_id_map m
 WHERE i.section_id = m.old_id;

UPDATE sections_history sh
   SET section_id = m.new_id
  FROM section_id_map m
 WHERE sh.section_id = m.old_id;

UPDATE pages_history ph
   SET section_id = m.new_id
  FROM section_id_map m
 WHERE ph.section_id = m.old_id;

-- ============================================================
-- 5. Update sections.id to new UUIDs
-- ============================================================
UPDATE sections s
   SET id = m.new_id
  FROM section_id_map m
 WHERE s.id = m.old_id;

-- ============================================================
-- 6. Set default for new rows
-- ============================================================
ALTER TABLE sections ALTER COLUMN id SET DEFAULT gen_random_uuid()::text;

-- ============================================================
-- 7. Unique index on name (active sections only)
-- ============================================================
CREATE UNIQUE INDEX sections_name_active ON sections(name) WHERE deleted = false;

-- Re-create partial unique index on pages
CREATE UNIQUE INDEX pages_section_id_slug_active ON pages(section_id, slug) WHERE deleted = false;

-- ============================================================
-- 8. Re-add FK constraints
-- ============================================================
ALTER TABLE pages ADD CONSTRAINT pages_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE;

ALTER TABLE sections_history ADD CONSTRAINT sections_history_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id) ON DELETE CASCADE;
ALTER TABLE sections_history ADD CONSTRAINT sections_history_section_id_version_key
    UNIQUE (section_id, version);

ALTER TABLE images ADD CONSTRAINT images_section_id_fkey
    FOREIGN KEY (section_id) REFERENCES sections(id);

-- Clean up temp table
DROP TABLE section_id_map;

COMMIT;
