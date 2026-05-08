ALTER TABLE dashboards ADD COLUMN share_token VARCHAR(64) DEFAULT NULL AFTER config;
ALTER TABLE dashboards ADD COLUMN share_enabled TINYINT DEFAULT 0 AFTER share_token;
CREATE UNIQUE INDEX idx_share_token ON dashboards(share_token);
