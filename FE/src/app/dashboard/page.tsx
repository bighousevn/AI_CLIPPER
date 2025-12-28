"use client";
import { useRouter } from "next/navigation";
import { DashboardClient } from "~/components/dashboard-client";
import { useAuth } from "~/hooks/useAuth";
import { useClips } from "~/hooks/useClip";
import { useUploadedFiles } from "~/hooks/useUpload";



export default function Dashboard() {
    const { data: uploadedFiles, isLoading, isError } = useUploadedFiles();
    const { data: clips } = useClips();
    const { user } = useAuth();
    const router = useRouter();
    // if (user === null) {
    //     router.push("/login");
    // }


    if (isLoading) return <div>Loading...</div>;
    if (isError) return <div>Error</div>;
    return (

        <DashboardClient uploadedFiles={uploadedFiles || []} clips={clips || []} />
    );
}
