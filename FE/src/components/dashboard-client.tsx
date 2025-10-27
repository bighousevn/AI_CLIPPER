"use client";

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
import { Loader2 } from "lucide-react";
import { useState } from "react";

import { useRouter } from "next/navigation";
import { Dropzone, DropzoneContent, DropzoneEmptyState } from "./ui/shadcn-io/dropzone";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";

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
    const handleUpload = async () => {
        if (files.length === 0) return;
        const file = files[0];
        setUploading(true);

        try {
            // upload file to s3

        }
        catch (error) {
            console.log(error);
        } finally {
            setUploading(false);
        }
    }

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
                                {/* <div>
                                    {files.length > 0 && (
                                        <div className="space-u-1 text-sm">
                                            <p className="font-medium">selected file</p>
                                        </div>
                                    )}
                                </div> */}
                                <Button className="mt-4" disabled={uploading || files.length === 0} onClick={handleUpload}>
                                    {uploading ? <>
                                        <Loader2 className="mr-2 h-4 w-4  animate-spin" /> Uploading</> : "Upload and Generate Clips"}
                                </Button>
                            </div>
                            {uploadedFiles.length > 0 && (
                                <div className="pt-6">
                                    <div className="mb-2 flex items-center justify-between">
                                        <h3 className="text-md mb-2 font-medium">Queue status</h3>
                                        <Button
                                            variant="outline"
                                            size="sm"
                                            onClick={handleRefresh}
                                            disabled={refreshing}
                                        >
                                            {refreshing && (
                                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                            )}
                                            Refresh
                                        </Button>
                                    </div>
                                    <div className="max-h-[300px] overflow-auto rounded-md border">
                                        <Table>
                                            <TableHeader>
                                                <TableRow>
                                                    <TableHead>File</TableHead>
                                                    <TableHead>Uploaded</TableHead>
                                                    <TableHead>Status</TableHead>
                                                    <TableHead>Clips created</TableHead>
                                                </TableRow>
                                            </TableHeader>
                                            <TableBody>
                                                {uploadedFiles.map((item) => (
                                                    <TableRow key={item.id}>
                                                        <TableCell className="max-w-xs truncate font-medium">
                                                            {item.filename}
                                                        </TableCell>
                                                        <TableCell className="text-muted-foreground text-sm">
                                                            {new Date(item.createdAt).toLocaleDateString()}
                                                        </TableCell>
                                                        <TableCell>
                                                            {item.status === "queued" && (
                                                                <Badge variant="outline">Queued</Badge>
                                                            )}
                                                            {item.status === "processing" && (
                                                                <Badge variant="outline">Processing</Badge>
                                                            )}
                                                            {item.status === "processed" && (
                                                                <Badge variant="outline">Processed</Badge>
                                                            )}
                                                            {item.status === "no credits" && (
                                                                <Badge variant="destructive">No credits</Badge>
                                                            )}
                                                            {item.status === "failed" && (
                                                                <Badge variant="destructive">Failed</Badge>
                                                            )}
                                                        </TableCell>
                                                        <TableCell>
                                                            {item.clipsCount > 0 ? (
                                                                <span>
                                                                    {item.clipsCount} clip
                                                                    {item.clipsCount !== 1 ? "s" : ""}
                                                                </span>
                                                            ) : (
                                                                <span className="text-muted-foreground">
                                                                    No clips yet
                                                                </span>
                                                            )}
                                                        </TableCell>
                                                    </TableRow>
                                                ))}
                                            </TableBody>
                                        </Table>
                                    </div>
                                </div>
                            )}
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
