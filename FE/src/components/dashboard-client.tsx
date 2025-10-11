"use client";

import type { Clip } from "@prisma/client";
import Link from "next/link";
import { Button } from "./ui/button";
import { Tabs, TabsContent, TabsList, TabsTrigger } from "./ui/tabs";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "./ui/card";
import { Loader2, UploadCloud } from "lucide-react";
import { useState } from "react";
import { toast } from "sonner";
import {
    Table,
    TableBody,
    TableCell,
    TableHead,
    TableHeader,
    TableRow,
} from "./ui/table";
import { Badge } from "./ui/badge";
import { useRouter } from "next/navigation";
import type { DropzoneState } from "react-dropzone";
import { Dropzone, DropzoneContent, DropzoneEmptyState } from "./ui/shadcn-io/dropzone";

export function DashboardClient({
    uploadedFiles,
    clips,
}: {
    uploadedFiles: {
        id: string;
        s3Key: string;
        filename: string;
        status: string;
        clipsCount: number;
        createdAt: Date;
    }[];
    clips: {
        id: string;
        title: string;
        s3Key: string;
        createdAt: Date;
        uploadedFileId: string;
        views: number;
        videoUrl: string;
        thumbnailUrl: string;
    }[]

}) {
    const [files, setFiles] = useState<File[]>([]);
    const [uploading, setUploading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const router = useRouter();

    const handleRefresh = async () => {
        setRefreshing(true);
        router.refresh();
        setTimeout(() => setRefreshing(false), 600);
    };

    const handleDrop = (acceptedFiles: File[]) => {
        setFiles(acceptedFiles);
    };


    return (
        <div className="mx-auto flex max-w-5xl flex-col space-y-6 px-4 py-8">
            {/* Header */}
            <div className="flex items-center justify-between">
                <div>
                    <h1 className="text-2xl font-semibold tracking-tight">
                        Podcast Clipper
                    </h1>
                    <p className="text-muted-foreground">
                        Upload your podcast and get AI-generated clips instantly
                    </p>
                </div>
                <Link href="/dashboard/billing">
                    <Button>Buy Credits</Button>
                </Link>
            </div>
            {/* content */}
            <Tabs defaultValue="upload">
                <TabsList>
                    <TabsTrigger value="upload">Upload</TabsTrigger>
                    <TabsTrigger value="my-clips">My Clips</TabsTrigger>
                </TabsList>

                <TabsContent value="upload">
                    <Card>
                        <CardHeader>
                            <CardTitle>Upload Podcast</CardTitle>
                            <CardDescription>
                                Upload your audio or video file to generate clips
                            </CardDescription>
                        </CardHeader>
                        <CardContent>
                            {/* <Dropzone
                                onDrop={handleDrop}
                                accept={{ "video/mp4": [".mp4"] }}
                                maxSize={500 * 1024 * 1024}
                                disabled={uploading}
                                maxFiles={1}
                            >
                                {(dropzone: DropzoneState) => (
                                    <>
                                        <div className="flex flex-col items-center justify-center space-y-4 rounded-lg p-10 text-center">
                                            <UploadCloud className="text-muted-foreground h-12 w-12" />
                                            <p className="font-medium">Drag and drop your file</p>
                                            <p className="text-muted-foreground text-sm">
                                                or click to browse (MP4 up to 500MB)
                                            </p>
                                            <Button
                                                className="cursor-pointer"
                                                variant="default"
                                                size="sm"
                                                disabled={uploading}
                                            >
                                                Select File
                                            </Button>
                                        </div>
                                    </>
                                )}
                            </Dropzone> */}
                            <Dropzone
                                accept={{ "video/mp4": [".mp4"] }}
                                disabled={uploading}
                                maxFiles={1}
                                maxSize={500 * 1024 * 1024}
                                minSize={1024}
                                onDrop={handleDrop}
                                onError={console.error}
                                src={files}
                            >
                                <DropzoneEmptyState />
                                <DropzoneContent />
                            </Dropzone>
                            <div className="flex items-start justify-between">
                                <div>
                                    {files.length > 0 && (
                                        <div className="space-u-1 text-sm">
                                            <p className="font-medium">selected file</p>
                                        </div>
                                    )}
                                </div>
                                <Button>
                                    {uploading ? <>
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Uploading</> : "Upload"}
                                </Button>
                            </div>
                        </CardContent>

                    </Card>
                </TabsContent>

                <TabsContent value="my-clips">
                    <Card>
                        <CardHeader>
                            <CardTitle>My Clips</CardTitle>
                            <CardDescription>
                                View and manage your generated clips here. Processing may take a
                                few minutes.
                            </CardDescription>
                        </CardHeader>

                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
