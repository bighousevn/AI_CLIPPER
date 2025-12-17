// export interface UploadFile {
//     id: string;
//     s3Key: string;
//     filename: string;
//     status: string;
//     clipsCount: number;
//     createdAt: Date;
// }
export interface UploadFile {
    id: string;
    file_name: string;
    file_path: string;
    file_size: number;       // tính bằng bytes
    status: string;
    clip_count: number;
    created_at: string;      // ISO string hoặc "YYYY-MM-DD HH:mm:ss"
    updated_at: string;      // ISO string hoặc "YYYY-MM-DD HH:mm:ss"
}