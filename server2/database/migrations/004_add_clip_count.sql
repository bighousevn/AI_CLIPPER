-- Add clip_count column to uploaded_files table
ALTER TABLE uploaded_files 
ADD COLUMN IF NOT EXISTS clip_count INTEGER DEFAULT 0;
