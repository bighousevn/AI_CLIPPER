"use client";

import { useEffect, useState } from "react";
import type { Clip } from "~/interfaces/clip";
import { Button } from "./ui/button";
import { Download, Loader2, Play } from "lucide-react";

function ClipCard({ clip }: { clip: Clip }) {
    const [playUrl, setPlayUrl] = useState<string | null>("https://rniuoasrouxefivcxxkg.supabase.co/storage/v1/object/sign/uploaded_files/user-c4764ef9-b56e-4004-a8c7-720f0a500b74/clips/clip_0.mp4?token=eyJraWQiOiJzdG9yYWdlLXVybC1zaWduaW5nLWtleV80MmVmNTA2Ni05YzIyLTQ1MjEtYmI3Yy1iNDA3NGYzM2U1ODIiLCJhbGciOiJIUzI1NiJ9.eyJ1cmwiOiJ1cGxvYWRlZF9maWxlcy91c2VyLWM0NzY0ZWY5LWI1NmUtNDAwNC1hOGM3LTcyMGYwYTUwMGI3NC9jbGlwcy9jbGlwXzAubXA0IiwiaWF0IjoxNzYyODU0MjI2LCJleHAiOjE3NjI4NTc4MjZ9.xCS9XSC0Ud6_dCvn4q9OXqvCcwqVQpS-3V6IRYOhGvQ");
    const [isLoadingUrl, setIsLoadingUrl] = useState(false);

    // useEffect(() => {
    //   async function fetchPlayUrl() {
    //     try {
    //       const result = await getClipPlayUrl(clip.id);
    //       if (result.succes && result.url) {
    //         setPlayUrl(result.url);
    //       } else if (result.error) {
    //         console.error("Failed to get play url: " + result.error);
    //       }
    //     } catch (error) {
    //     } finally {
    //       setIsLoadingUrl(false);
    //     }
    //   }

    //   void fetchPlayUrl();
    // }, [clip.id]);

    const handleDownload = () => {
        if (playUrl) {
            const link = document.createElement("a");
            link.href = playUrl;
            link.style.display = "none";
            document.body.appendChild(link);
            link.click();
            document.body.removeChild(link);
        }
    };

    return (
        <div className="flex max-w-52 flex-col gap-2">
            <div className="bg-muted">
                {isLoadingUrl ? (
                    <div className="flex h-full w-full items-center justify-center">
                        <Loader2 className="text-muted-foreground h-8 w-8 animate-spin" />
                    </div>
                ) : playUrl ? (
                    <video
                        src={playUrl}
                        controls
                        preload="metadata"
                        className="h-full w-full rounded-md object-cover"
                    />
                ) : (
                    <div className="flex h-full w-full items-center justify-center">
                        <Play className="text-muted-foreground h-10 w-10 opacity-50" />
                    </div>
                )}
            </div>
            <div className="flex flex-col gap-2">
                <Button onClick={handleDownload} variant="outline" size="sm">
                    <Download className="mr-1.5 h-4 w-4" />
                    Download
                </Button>
            </div>
        </div>
    );
}

export function ClipDisplay({ clips }: { clips: Clip[] }) {
    if (clips.length === 0) {
        return (
            <p className="text-muted-foreground p-4 text-center">
                No clips generated yet.
            </p>
        );
    }
    return (
        <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
            {clips.map((clip) => (
                <ClipCard key={clip.id} clip={clip} />
            ))}
        </div>
    );
}