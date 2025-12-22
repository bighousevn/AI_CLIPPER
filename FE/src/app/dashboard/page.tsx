"use client";
import { DashboardClient } from "~/components/dashboard-client";
import { useClips } from "~/hooks/useClip";
import { useUploadedFiles } from "~/hooks/useUpload";



export default function Dashboard() {
    const { data: uploadedFiles, isLoading, isError } = useUploadedFiles();
    const { data: clips } = useClips();



    if (isLoading) return <div>Loading...</div>;
    if (isError) return <div>Error</div>;
    return (

        <DashboardClient uploadedFiles={uploadedFiles || []} clips={clips || []} />
    );
}
