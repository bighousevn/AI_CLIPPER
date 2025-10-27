import { DashboardClient } from "~/components/dashboard-client";


export default async function Dashboard() {

    const userData = {
        uploadedFiles: [
            {
                id: "1",
                s3Key: "key1",
                displayName: "My first video.mp4",
                status: "DONE",
                createdAt: new Date(),
                _count: {
                    clips: 5,
                },
            },
            {
                id: "2",
                s3Key: "key2",
                displayName: "Another video.mov",
                status: "PROCESSING",
                createdAt: new Date(Date.now() - 1000 * 60 * 60),
                _count: {
                    clips: 0,
                },
            },
            {
                id: "3",
                s3Key: "key3",
                displayName: "Holiday footage.mkv",
                status: "UPLOADED",
                createdAt: new Date(Date.now() - 1000 * 60 * 60 * 24),
                _count: {
                    clips: 12,
                },
            },
        ],
        clips: [
            { id: "clip1", title: "Funny moment 1", s3Key: "clip_key_1", createdAt: new Date(), uploadedFileId: "1", views: 100, videoUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", thumbnailUrl: "https://example.com/thumb1.jpg" },
            { id: "clip2", title: "Amazing skill", s3Key: "clip_key_2", createdAt: new Date(Date.now() - 1000 * 60 * 5), uploadedFileId: "1", views: 250, videoUrl: "https://example.com/clip2.mp4", thumbnailUrl: "https://example.com/thumb2.jpg" },
            { id: "clip3", title: "Tutorial part 3", s3Key: "clip_key_3", createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2), uploadedFileId: "3", views: 50, videoUrl: "https://example.com/clip3.mp4", thumbnailUrl: "https://example.com/thumb3.jpg" },
        ]
    };

    const formattedFiles = userData?.uploadedFiles.map((file) => ({
        id: file.id,
        s3Key: file.s3Key,
        filename: file.displayName ?? "Unknown filename",
        status: file.status,
        clipsCount: file._count.clips,
        createdAt: file.createdAt,
    }));

    return (
        <DashboardClient uploadedFiles={formattedFiles ?? []} clips={userData?.clips ?? []} />
    );
}
