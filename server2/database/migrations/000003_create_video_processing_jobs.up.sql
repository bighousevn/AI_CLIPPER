-- Create video_processing_jobs table
CREATE TABLE IF NOT EXISTS video_processing_jobs (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    upload_file_id UUID NOT NULL REFERENCES uploaded_files(id) ON DELETE CASCADE,
    storage_path VARCHAR(500) NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    result_data TEXT,
    error_message TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_video_jobs_user_id ON video_processing_jobs(user_id);
CREATE INDEX IF NOT EXISTS idx_video_jobs_status ON video_processing_jobs(status);
CREATE INDEX IF NOT EXISTS idx_video_jobs_created_at ON video_processing_jobs(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_video_jobs_upload_file_id ON video_processing_jobs(upload_file_id);
