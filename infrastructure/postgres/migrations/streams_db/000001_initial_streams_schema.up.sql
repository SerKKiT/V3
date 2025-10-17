-- infrastructure/postgres/migrations/streams_db/000001_initial_streams_schema.up.sql

BEGIN;

-- Create streams table
CREATE TABLE IF NOT EXISTS streams (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    stream_key VARCHAR(255) UNIQUE NOT NULL,
    title VARCHAR(255) NOT NULL,
    description TEXT,
    status VARCHAR(50) DEFAULT 'offline' NOT NULL,
    viewer_count INT DEFAULT 0 NOT NULL,
    started_at TIMESTAMP,
    ended_at TIMESTAMP,
    thumbnail_url TEXT,
    hls_url TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,

    -- Constraints
    CONSTRAINT title_length CHECK (char_length(title) >= 3),
    CONSTRAINT viewer_count_positive CHECK (viewer_count >= 0),
    CONSTRAINT valid_status CHECK (status IN ('offline', 'live', 'starting', 'stopping', 'error')),
    CONSTRAINT valid_timestamps CHECK (ended_at IS NULL OR ended_at >= started_at)
);

-- Create indexes for efficient queries (with IF NOT EXISTS)
CREATE INDEX IF NOT EXISTS idx_streams_user_id ON streams(user_id);
CREATE INDEX IF NOT EXISTS idx_streams_status ON streams(status);
CREATE INDEX IF NOT EXISTS idx_streams_stream_key ON streams(stream_key);
CREATE INDEX IF NOT EXISTS idx_streams_created_at ON streams(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_streams_started_at ON streams(started_at DESC) WHERE started_at IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_streams_live ON streams(status, created_at DESC) WHERE status = 'live';

-- Table comments
COMMENT ON TABLE streams IS 'Live streaming sessions metadata';
COMMENT ON COLUMN streams.stream_key IS 'Unique key for SRT streaming (format: user_{userid}_{streamid})';
COMMENT ON COLUMN streams.status IS 'Stream status: offline, live, starting, stopping, error';
COMMENT ON COLUMN streams.viewer_count IS 'Current number of viewers (updated in real-time)';
COMMENT ON COLUMN streams.hls_url IS 'HLS playlist URL for playback';
COMMENT ON COLUMN streams.thumbnail_url IS 'Stream thumbnail/preview image URL';

-- Create function to update updated_at automatically
CREATE OR REPLACE FUNCTION update_streams_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger
DROP TRIGGER IF EXISTS trigger_update_streams_timestamp ON streams;
CREATE TRIGGER trigger_update_streams_timestamp
BEFORE UPDATE ON streams
FOR EACH ROW
EXECUTE FUNCTION update_streams_updated_at();

COMMIT;
