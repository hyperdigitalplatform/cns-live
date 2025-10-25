-- Tags table
CREATE TABLE IF NOT EXISTS tags (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    name VARCHAR(100) NOT NULL UNIQUE,
    category VARCHAR(100),
    color VARCHAR(20),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_tags_name ON tags(name);
CREATE INDEX IF NOT EXISTS idx_tags_category ON tags(category);

-- Video tags (many-to-many relationship between segments and tags)
CREATE TABLE IF NOT EXISTS video_tags (
    segment_id UUID NOT NULL,
    tag_id UUID NOT NULL REFERENCES tags(id) ON DELETE CASCADE,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    PRIMARY KEY (segment_id, tag_id)
);

CREATE INDEX IF NOT EXISTS idx_video_tags_segment ON video_tags(segment_id);
CREATE INDEX IF NOT EXISTS idx_video_tags_tag ON video_tags(tag_id);

-- Annotations table
CREATE TABLE IF NOT EXISTS annotations (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    segment_id UUID NOT NULL,
    timestamp_offset INTEGER NOT NULL,
    type VARCHAR(50) NOT NULL CHECK (type IN ('NOTE', 'MARKER', 'WARNING', 'EVIDENCE')),
    content TEXT NOT NULL,
    user_id VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_annotations_segment ON annotations(segment_id);
CREATE INDEX IF NOT EXISTS idx_annotations_timestamp ON annotations(segment_id, timestamp_offset);

-- Incidents table
CREATE TABLE IF NOT EXISTS incidents (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    title VARCHAR(255) NOT NULL,
    description TEXT,
    severity VARCHAR(20) NOT NULL CHECK (severity IN ('LOW', 'MEDIUM', 'HIGH', 'CRITICAL')),
    status VARCHAR(20) NOT NULL DEFAULT 'OPEN' CHECK (status IN ('OPEN', 'IN_PROGRESS', 'RESOLVED', 'CLOSED')),
    camera_ids UUID[] NOT NULL,
    start_time TIMESTAMP WITH TIME ZONE NOT NULL,
    end_time TIMESTAMP WITH TIME ZONE NOT NULL,
    tags TEXT[],
    assigned_to VARCHAR(255),
    created_by VARCHAR(255) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    closed_at TIMESTAMP WITH TIME ZONE,
    search_vector tsvector
);

-- Indexes for incidents
CREATE INDEX IF NOT EXISTS idx_incidents_status ON incidents(status);
CREATE INDEX IF NOT EXISTS idx_incidents_severity ON incidents(severity);
CREATE INDEX IF NOT EXISTS idx_incidents_time ON incidents(start_time, end_time);
CREATE INDEX IF NOT EXISTS idx_incidents_camera_ids ON incidents USING GIN(camera_ids);
CREATE INDEX IF NOT EXISTS idx_incidents_tags ON incidents USING GIN(tags);
CREATE INDEX IF NOT EXISTS idx_incidents_search ON incidents USING GIN(search_vector);

-- Trigger function to update search_vector
CREATE OR REPLACE FUNCTION incidents_search_trigger() RETURNS trigger AS $$
BEGIN
    NEW.search_vector :=
        setweight(to_tsvector('english', COALESCE(NEW.title, '')), 'A') ||
        setweight(to_tsvector('english', COALESCE(NEW.description, '')), 'B');
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically update search_vector
DROP TRIGGER IF EXISTS incidents_search_update ON incidents;
CREATE TRIGGER incidents_search_update
    BEFORE INSERT OR UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION incidents_search_trigger();

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS update_incidents_updated_at ON incidents;
CREATE TRIGGER update_incidents_updated_at
    BEFORE UPDATE ON incidents
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
