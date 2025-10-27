export interface Clip {
    id: string;
    title: string;
    s3Key: string;
    createdAt: Date;
    uploadedFileId: string;
    views: number;
    videoUrl: string;
    thumbnailUrl: string;
}