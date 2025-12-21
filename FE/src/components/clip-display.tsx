"use client";

import { useMemo, useState } from "react";
import type { Clip } from "~/interfaces/clip";
import { Button } from "./ui/button";
import { Download, Loader2, Play } from "lucide-react";
import { Select, SelectTrigger, SelectItem, SelectContent, SelectValue } from "./ui/select";


export function ClipDisplay({ clips }: { clips: Clip[] }) {
    const [filterFile, setFilterFile] = useState<string>("all");

    // Lấy danh sách unique uploaded_file_id
    const fileOptions = useMemo(() => {
        const ids = Array.from(
            new Set(
                clips
                    .map(c => c.source_name?.trim())
                    .filter(x => x && x !== "")
            )
        );
        return ids;
    }, [clips]);

    // Sort + Filter clips
    const filteredClips = useMemo(() => {
        let result = [...clips];

        // Filter theo uploaded_file_id
        if (filterFile !== "all") {
            result = result.filter(c => c.source_name === filterFile);
        }

        // Sort mới nhất trước
        result.sort(
            (a, b) => new Date(b.created_at).getTime() - new Date(a.created_at).getTime()
        );

        return result;
    }, [clips, filterFile]);

    if (filteredClips.length === 0) {
        return (
            <p className="text-muted-foreground p-4 text-center">
                No clips found.
            </p>
        );
    }

    return (
        <div className="space-y-3">
            {/* Filter Dropdown */}
            <div className="flex items-center gap-2">
                <p className="text-sm text-muted-foreground">Filter by file:</p>
                <Select onValueChange={setFilterFile} defaultValue="all">
                    <SelectTrigger className="w-[200px]">
                        <SelectValue placeholder="Select a file" />
                    </SelectTrigger>
                    <SelectContent>
                        <SelectItem value="all">All files</SelectItem>
                        {fileOptions.map((id) => (
                            <SelectItem key={id} value={id}>
                                {id}
                            </SelectItem>
                        ))}
                    </SelectContent>
                </Select>
            </div>

            {/* Display clips */}
            <div className="grid grid-cols-2 gap-2 sm:grid-cols-3 lg:grid-cols-4">
                {filteredClips.map((clip) => (
                    <ClipCard key={clip.id} clip={clip} />
                ))}
            </div>
        </div>
    );
}
function ClipCard({ clip }: { clip: Clip }) {
    const playUrl = clip.download_url;
    const [isLoadingUrl] = useState(false);



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
                <p className="text-xs text-muted-foreground">
                    Created: {new Date(clip.created_at).toLocaleString()}
                </p>
                <Button onClick={handleDownload} variant="outline" size="sm">
                    <Download className="mr-1.5 h-4 w-4" />
                    Download
                </Button>
            </div>
        </div>
    );
}
