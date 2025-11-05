-- Rollback grid_layout column addition

-- Remove grid_layout column from layout_preferences
ALTER TABLE layout_preferences
DROP COLUMN IF EXISTS grid_layout;
