-- Add screen mode support to dashboards
ALTER TABLE dashboards ADD COLUMN type VARCHAR(32) DEFAULT 'dashboard' AFTER title;
CREATE INDEX idx_type ON dashboards(type);
UPDATE dashboards SET type = 'dashboard' WHERE type IS NULL;
