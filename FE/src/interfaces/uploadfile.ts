export interface UploadFile {
    id: string;
    s3Key: string;
    filename: string;
    status: string;
    clipsCount: number;
    createdAt: Date;
}
// BODY request tham khảo
// {
//     "fileId": "abc123",
//     "prompt": "Find motivational quotes",
//     "clip_count": 4,
//     "clip_duration": 20,
//     "skip_start": 5,
//     "skip_end": 0,
//     "aspect_ratio": "9:16",
//     "output_format": "mp4",
//     "generate_subtitles": true,
//     "language": "auto",
//     "tone": "inspirational"
//   }
