"use client";
import { DashboardClient } from "~/components/dashboard-client";
import { useAuth } from "~/hooks/useAuth";
import { useClips } from "~/hooks/useClip";
import { useUploadedFiles } from "~/hooks/useUpload";
import type { UploadFile } from "~/interfaces/uploadfile";



export default function Dashboard() {
    const { user } = useAuth();
    const { data: uploadedFiles, isLoading, isError } = useUploadedFiles();
    const { data: clips } = useClips();



    if (isLoading) return <div>Loading...</div>;
    if (isError) return <div>Error</div>;
    return (

        <DashboardClient uploadedFiles={uploadedFiles || []} clips={clips || []} />
    );
}
