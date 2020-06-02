ALTER TABLE blob ADD COLUMN IF NOT EXISTS update_time timestamp default CURRENT_TIMESTAMP;
ALTER TABLE blob ADD COLUMN IF NOT EXISTS status varchar(255);
ALTER TABLE blob ADD COLUMN IF NOT EXISTS version BIGINT default 0;
CREATE INDEX IF NOT EXISTS idx_status ON blob (status);
CREATE INDEX IF NOT EXISTS idx_version ON blob (version);