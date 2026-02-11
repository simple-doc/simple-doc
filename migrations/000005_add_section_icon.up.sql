ALTER TABLE sections ADD COLUMN icon TEXT NOT NULL DEFAULT 'document';
ALTER TABLE sections_history ADD COLUMN icon TEXT NOT NULL DEFAULT 'document';

-- Backfill existing sections with their hardcoded icons
UPDATE sections SET icon = 'download' WHERE id = 'ingest-api';
UPDATE sections SET icon = 'grid' WHERE id = 'animation-api';
UPDATE sections SET icon = 'book' WHERE id = 'catalog-metadata';
