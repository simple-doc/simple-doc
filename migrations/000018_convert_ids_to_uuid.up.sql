BEGIN;

-- ============================================================
-- 1. pages: id SERIAL -> UUID
-- ============================================================

-- Drop FK + unique constraints in pages_history that reference pages(id)
ALTER TABLE pages_history DROP CONSTRAINT pages_history_page_id_fkey;
ALTER TABLE pages_history DROP CONSTRAINT pages_history_page_id_version_key;

-- Convert pages.id from int to UUID
ALTER TABLE pages ALTER COLUMN id DROP DEFAULT;
ALTER TABLE pages ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE pages ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS pages_id_seq;

-- Convert pages_history.page_id to UUID
ALTER TABLE pages_history ALTER COLUMN page_id SET DATA TYPE UUID USING gen_random_uuid();

-- Convert pages_history.id from int to UUID
ALTER TABLE pages_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE pages_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE pages_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS pages_history_id_seq;

-- Re-add constraints
ALTER TABLE pages_history ADD CONSTRAINT pages_history_page_id_fkey
    FOREIGN KEY (page_id) REFERENCES pages(id) ON DELETE CASCADE;
ALTER TABLE pages_history ADD CONSTRAINT pages_history_page_id_version_key
    UNIQUE (page_id, version);

-- ============================================================
-- 2. images: id SERIAL -> UUID
-- ============================================================

ALTER TABLE images_history DROP CONSTRAINT images_history_image_id_fkey;
ALTER TABLE images_history DROP CONSTRAINT images_history_image_id_version_key;

ALTER TABLE images ALTER COLUMN id DROP DEFAULT;
ALTER TABLE images ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE images ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS images_id_seq;

ALTER TABLE images_history ALTER COLUMN image_id SET DATA TYPE UUID USING gen_random_uuid();

ALTER TABLE images_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE images_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE images_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS images_history_id_seq;

ALTER TABLE images_history ADD CONSTRAINT images_history_image_id_fkey
    FOREIGN KEY (image_id) REFERENCES images(id) ON DELETE CASCADE;
ALTER TABLE images_history ADD CONSTRAINT images_history_image_id_version_key
    UNIQUE (image_id, version);

-- ============================================================
-- 3. section_rows: id SERIAL -> UUID
-- ============================================================

-- Drop FK constraints that reference section_rows(id)
ALTER TABLE sections DROP CONSTRAINT sections_row_id_fkey;
ALTER TABLE section_rows_history DROP CONSTRAINT section_rows_history_row_id_fkey;
ALTER TABLE section_rows_history DROP CONSTRAINT section_rows_history_row_id_version_key;

-- Convert section_rows.id
ALTER TABLE section_rows ALTER COLUMN id DROP DEFAULT;
ALTER TABLE section_rows ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE section_rows ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS section_rows_id_seq;

-- Convert sections.row_id FK column
ALTER TABLE sections ALTER COLUMN row_id SET DATA TYPE UUID USING NULL;

-- Convert sections_history.row_id (no FK, just data)
ALTER TABLE sections_history ALTER COLUMN row_id SET DATA TYPE UUID USING NULL;

-- Convert section_rows_history columns
ALTER TABLE section_rows_history ALTER COLUMN row_id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE section_rows_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE section_rows_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE section_rows_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS section_rows_history_id_seq;

-- Re-add constraints
ALTER TABLE sections ADD CONSTRAINT sections_row_id_fkey
    FOREIGN KEY (row_id) REFERENCES section_rows(id) ON DELETE SET NULL;
ALTER TABLE section_rows_history ADD CONSTRAINT section_rows_history_row_id_fkey
    FOREIGN KEY (row_id) REFERENCES section_rows(id) ON DELETE CASCADE;
ALTER TABLE section_rows_history ADD CONSTRAINT section_rows_history_row_id_version_key
    UNIQUE (row_id, version);

-- ============================================================
-- 4. site_settings: id INT CHECK(id=1) -> UUID + singleton
-- ============================================================

ALTER TABLE site_settings DROP CONSTRAINT site_settings_id_check;
ALTER TABLE site_settings ALTER COLUMN id DROP DEFAULT;
ALTER TABLE site_settings ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE site_settings ALTER COLUMN id SET DEFAULT gen_random_uuid();
ALTER TABLE site_settings ADD COLUMN singleton BOOLEAN NOT NULL DEFAULT TRUE;
ALTER TABLE site_settings ADD CONSTRAINT site_settings_singleton_unique UNIQUE (singleton);
ALTER TABLE site_settings ADD CONSTRAINT site_settings_singleton_check CHECK (singleton = TRUE);

-- ============================================================
-- 5. Remaining history table PKs
-- ============================================================

-- sections_history.id
ALTER TABLE sections_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE sections_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE sections_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS sections_history_id_seq;

-- site_settings_history.id
ALTER TABLE site_settings_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE site_settings_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE site_settings_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS site_settings_history_id_seq;

-- users_history.id
ALTER TABLE users_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE users_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE users_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS users_history_id_seq;

-- roles_history.id
ALTER TABLE roles_history ALTER COLUMN id DROP DEFAULT;
ALTER TABLE roles_history ALTER COLUMN id SET DATA TYPE UUID USING gen_random_uuid();
ALTER TABLE roles_history ALTER COLUMN id SET DEFAULT gen_random_uuid();
DROP SEQUENCE IF EXISTS roles_history_id_seq;

COMMIT;
