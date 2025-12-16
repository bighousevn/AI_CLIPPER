"use client";
import { DashboardClient } from "~/components/dashboard-client";
import { useAuth } from "~/hooks/useAuth";
import { useClips } from "~/hooks/useClip";
import { useUploadedFiles } from "~/hooks/useUpload";
import type { UploadFile } from "~/interfaces/Uploadfile";



export default function Dashboard() {
    const { user } = useAuth();
    const { data: uploadedFiles, isLoading, isError } = useUploadedFiles();

    const { data: clips, isLoading: isClipsLoading, isError: isClipsError } = useClips();
    console.log("uploadedFiles", uploadedFiles);

    const userData = {

        clips: [
            { id: "clip1", title: "Funny moment 1", s3Key: "clip_key_1", createdAt: new Date(), uploadedFileId: "1", views: 100, videoUrl: "https://commondatastorage.googleapis.com/gtv-videos-bucket/sample/BigBuckBunny.mp4", thumbnailUrl: "https://example.com/thumb1.jpg" },
            { id: "clip2", title: "Amazing skill", s3Key: "clip_key_2", createdAt: new Date(Date.now() - 1000 * 60 * 5), uploadedFileId: "1", views: 250, videoUrl: "https://example.com/clip2.mp4", thumbnailUrl: "https://example.com/thumb2.jpg" },
            { id: "clip3", title: "Tutorial part 3", s3Key: "clip_key_3", createdAt: new Date(Date.now() - 1000 * 60 * 60 * 2), uploadedFileId: "3", views: 50, videoUrl: "https://example.com/clip3.mp4", thumbnailUrl: "https://example.com/thumb3.jpg" },
        ]
    };
    const fakeUploadedFiles: UploadFile[] = [
        {
            id: "file-001",
            s3Key: "uploads/2024/12/video-presentation.mp4",
            filename: "Marketing Presentation Q4.mp4",
            status: "completed",
            clipsCount: 5,
            createdAt: new Date("2024-12-10T09:30:00")
        },
        {
            id: "file-002",
            s3Key: "uploads/2024/12/product-demo.mp4",
            filename: "Product Demo Final.mp4",
            status: "processing",
            clipsCount: 0,
            createdAt: new Date("2024-12-14T14:20:00")
        },
        {
            id: "file-003",
            s3Key: "uploads/2024/12/tutorial-video.mp4",
            filename: "How to use our platform.mp4",
            status: "completed",
            clipsCount: 12,
            createdAt: new Date("2024-12-13T16:45:00")
        },
        {
            id: "file-004",
            s3Key: "uploads/2024/12/webinar-recording.mp4",
            filename: "Monthly Webinar - December.mp4",
            status: "failed",
            clipsCount: 0,
            createdAt: new Date("2024-12-12T11:15:00")
        },
        {
            id: "file-005",
            s3Key: "uploads/2024/12/training-session.mp4",
            filename: "Team Training Session.mp4",
            status: "completed",
            clipsCount: 8,
            createdAt: new Date("2024-12-11T08:00:00")
        }
    ];

    if (isLoading) return <div>Loading...</div>;
    if (isError) return <div>Error</div>;
    return (

        <DashboardClient uploadedFiles={fakeUploadedFiles} clips={userData?.clips ?? []} />
    );
}
