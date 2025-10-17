-- infrastructure/postgres/migrations/vod_db/000001_initial_vod_schema.up.sql
-- Migration: Initial VOD database schema
-- Description: Create recordings and videos tables

BEGIN;

-- Create ENUM types
CREATE TYPE recording_status AS ENUM ('recording', 'processing', 'completed', 'failed');
CREATE TYPE video_status AS ENUM ('pending', 'ready', 'failed', 'archived');
CREATE TYPE video_source AS ENUM ('recording', 'upload', 'import');
CREATE TYPE video_visibility AS ENUM ('public', 'private', 'unlisted');

-- ============================================================
-- RECORDINGS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS recordings (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    stream_id UUID,
    video_id UUID,
    file_path TEXT NOT NULL,
    thumbnail_path TEXT,
    duration INT DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    status recording_status DEFAULT 'recording',
    started_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    completed_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    
    -- Constraints
    CONSTRAINT duration_positive CHECK (duration >= 0),
    CONSTRAINT file_size_positive CHECK (file_size >= 0),
    CONSTRAINT valid_recording_timestamps CHECK (completed_at IS NULL OR completed_at >= started_at)
);

-- Indexes for recordings
CREATE INDEX idx_recordings_stream_id ON recordings(stream_id);
CREATE INDEX idx_recordings_status ON recordings(status);
CREATE INDEX idx_recordings_video_id ON recordings(video_id);
CREATE INDEX idx_recordings_created_at ON recordings(created_at DESC);
CREATE INDEX idx_recordings_completed_at ON recordings(completed_at DESC) WHERE completed_at IS NOT NULL;

-- Comments for recordings
COMMENT ON TABLE recordings IS 'Stream recordings metadata (stored in recordings MinIO bucket)';
COMMENT ON COLUMN recordings.file_path IS 'Path in recordings MinIO bucket';
COMMENT ON COLUMN recordings.status IS 'Recording status: recording, processing, completed, failed';

-- ============================================================
-- VIDEOS TABLE
-- ============================================================
CREATE TABLE IF NOT EXISTS videos (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    recording_id UUID,
    stream_id UUID,
    
    -- Metadata
    title VARCHAR(255) NOT NULL,
    description TEXT,
    category VARCHAR(100),
    
    -- Technical fields
    source video_source DEFAULT 'recording',
    status video_status DEFAULT 'ready',
    visibility video_visibility DEFAULT 'private',
    
    -- File information
    file_path TEXT NOT NULL,
    thumbnail_path TEXT,
    duration INT DEFAULT 0,
    file_size BIGINT DEFAULT 0,
    resolution VARCHAR(20),
    
    -- Statistics
    view_count INT DEFAULT 0,
    like_count INT DEFAULT 0,
    
    -- Timestamps
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    published_at TIMESTAMP,
    
    -- Constraints
    CONSTRAINT title_length CHECK (char_length(title) >= 3),
    CONSTRAINT duration_positive CHECK (duration >= 0),
    CONSTRAINT file_size_positive CHECK (file_size >= 0),
    CONSTRAINT view_count_positive CHECK (view_count >= 0),
    CONSTRAINT like_count_positive CHECK (like_count >= 0),
    CONSTRAINT fk_recording FOREIGN KEY (recording_id) REFERENCES recordings(id) ON DELETE SET NULL
);

-- Indexes for videos
CREATE INDEX idx_videos_user_id ON videos(user_id);
CREATE INDEX idx_videos_recording_id ON videos(recording_id);
CREATE INDEX idx_videos_stream_id ON videos(stream_id);
CREATE INDEX idx_videos_status ON videos(status);
CREATE INDEX idx_videos_visibility ON videos(visibility);
CREATE INDEX idx_videos_category ON videos(category);
CREATE INDEX idx_videos_created_at ON videos(created_at DESC);
CREATE INDEX idx_videos_view_count ON videos(view_count DESC);
CREATE INDEX idx_videos_published_at ON videos(published_at DESC) WHERE published_at IS NOT NULL;

-- Full-text search index
CREATE INDEX idx_videos_search ON videos 
USING GIN(to_tsvector('english', title || ' ' || COALESCE(description, '')));

-- Index for public videos (most common query)
CREATE INDEX idx_videos_public ON videos(visibility, published_at DESC) 
WHERE visibility = 'public' AND status = 'ready';

-- Comments for videos
COMMENT ON TABLE videos IS 'VOD library - manages video metadata (stored in vod-videos MinIO bucket)';
COMMENT ON COLUMN videos.source IS 'Source: recording (from stream), upload (direct), import (external)';
COMMENT ON COLUMN videos.visibility IS 'Visibility: public, private, unlisted';
COMMENT ON COLUMN videos.status IS 'Processing status: pending, ready, failed, archived';
COMMENT ON COLUMN videos.file_path IS 'Path in vod-videos MinIO bucket';

-- ============================================================
-- TRIGGERS
-- ============================================================

-- Auto-update updated_at for recordings
CREATE OR REPLACE FUNCTION update_recordings_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_recordings_timestamp
BEFORE UPDATE ON recordings
FOR EACH ROW
EXECUTE FUNCTION update_recordings_updated_at();

-- Auto-update updated_at for videos
CREATE OR REPLACE FUNCTION update_videos_updated_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_videos_timestamp
BEFORE UPDATE ON videos
FOR EACH ROW
EXECUTE FUNCTION update_videos_updated_at();

COMMIT;
