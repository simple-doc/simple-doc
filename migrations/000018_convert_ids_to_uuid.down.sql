-- Down migration: drop all tables and re-apply would be needed.
-- For a docs app this is simpler than reversing each UUID conversion.
-- To rollback: restore from backup, then run migrations 1-17.

-- This is intentionally left as a no-op to avoid data loss.
-- Restore from a database backup if you need to revert.
SELECT 1;
