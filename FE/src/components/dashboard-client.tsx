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
import type { Clip } from "~/interfaces/clip";
import type { UploadFile } from "~/interfaces/uploadfile";
import { ClipDisplay } from "./clip-display";
import { processingFile, uploadFile } from "~/services/uploadService";
import { DropzoneVideoPreview } from "./DropzoneVideoPreview";
import type z from "zod";
import { ClipConfigSchema } from "~/schemas/clipConfigSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "./ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { useForm } from "react-hook-form";
import { Checkbox } from "./ui/checkbox";
import type { ClipConfig } from "~/interfaces/clipConfig";
import { useUploadClip } from "~/hooks/useUpload";

export function DashboardClient({
    uploadedFiles,
    clips,
}: {
    uploadedFiles: UploadFile[];
    clips: Clip[]

}) {
    const [files, setFiles] = useState<File[]>([]);
    const [uploading, setUploading] = useState(false);
    const [refreshing, setRefreshing] = useState(false);
    const router = useRouter();

    const form = useForm<ClipConfig>({
        resolver: zodResolver(ClipConfigSchema),
        defaultValues: {
            prompt: "",
            clipCount: 3,
            aspectRatio: "9:16",
            subtitle: false,
        },
    });
    const handleRefresh = async () => {
        setRefreshing(true);
        router.refresh();
        setTimeout(() => setRefreshing(false), 600);
    };

    const handleDrop = (acceptedFiles: File[]) => {
        setFiles(acceptedFiles);
    };
    const { mutate } = useUploadClip();

    const handleUpload = async () => {
        setUploading(true);
        if (files.length > 0) {
            const item = files[0] as File;
            await mutate({ file: item, config: form.getValues() });
        }
        setUploading(false);
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
                                <DropzoneContent>
                                    <DropzoneVideoPreview />
                                </DropzoneContent>
                            </Dropzone>
                            {files.length > 0 && (
                                <Card className="mt-6 animate-in fade-in slide-in-from-bottom-4 duration-300">
                                    <CardHeader>
                                        <CardTitle>Video Configuration</CardTitle>
                                        <CardDescription>
                                            Customize how clips will be generated.
                                        </CardDescription>
                                    </CardHeader>

                                    <CardContent>
                                        <Form {...form}>
                                            <form className="space-y-4">

                                                {/* Prompt */}
                                                <FormField
                                                    control={form.control}
                                                    name="prompt"
                                                    render={({ field }) => (
                                                        <FormItem>
                                                            <FormLabel>Prompt</FormLabel>
                                                            <FormControl>
                                                                <textarea
                                                                    className="w-full min-h-[80px] resize-none rounded-md border p-3"
                                                                    placeholder="Describe the tone, topic, or highlight you want..."
                                                                    {...field}
                                                                />
                                                            </FormControl>
                                                            <FormMessage />
                                                        </FormItem>
                                                    )}
                                                />

                                                {/* Number of clips */}
                                                <FormField
                                                    control={form.control}
                                                    name="clipCount"
                                                    render={({ field }) => (
                                                        <FormItem>
                                                            <FormLabel>Number of clips</FormLabel>
                                                            <FormControl>
                                                                <input
                                                                    type="number"
                                                                    className="input"
                                                                    min={1}
                                                                    max={10}
                                                                    {...field}
                                                                    onChange={(e) => field.onChange(Number(e.target.value))}
                                                                />
                                                            </FormControl>
                                                            <FormMessage />
                                                        </FormItem>
                                                    )}
                                                />



                                                {/* Aspect Ratio */}
                                                <FormField
                                                    control={form.control}
                                                    name="aspectRatio"
                                                    render={({ field }) => (
                                                        <FormItem>
                                                            <FormLabel>Aspect Ratio</FormLabel>
                                                            <Select onValueChange={field.onChange} defaultValue={field.value}>
                                                                <FormControl>
                                                                    <SelectTrigger>
                                                                        <SelectValue placeholder="Select ratio" />
                                                                    </SelectTrigger>
                                                                </FormControl>
                                                                <SelectContent>
                                                                    <SelectItem value="9:16">9:16 (TikTok, Reels)</SelectItem>
                                                                    <SelectItem value="16:9">16:9 (YouTube)</SelectItem>
                                                                    <SelectItem value="1:1">1:1 (Square)</SelectItem>
                                                                </SelectContent>
                                                            </Select>
                                                        </FormItem>
                                                    )}
                                                />
                                                {/* toggle subtitle */}
                                                <FormField
                                                    control={form.control}
                                                    name="subtitle"
                                                    render={({ field }) => (
                                                        <FormItem className="flex flex-row items-start space-x-3 space-y-0">
                                                            <FormControl>
                                                                <Checkbox
                                                                    checked={field.value}
                                                                    onCheckedChange={field.onChange}
                                                                />
                                                            </FormControl>
                                                            <div className="space-y-1 leading-none">
                                                                <FormLabel>
                                                                    Include Subtitle
                                                                </FormLabel>
                                                            </div>
                                                        </FormItem>
                                                    )}
                                                />

                                            </form>
                                        </Form>
                                    </CardContent>
                                </Card>
                            )}

                            <div className="flex items-start justify-between">

                                <Button className="mt-4" disabled={uploading || files.length === 0} onClick={handleUpload}>
                                    {uploading ? <>
                                        <Loader2 className="mr-2 h-4 w-4  animate-spin" /> Uploading</> : "Upload and Generate Clips"}
                                </Button>
                            </div>
                            {
                                uploadedFiles.length > 0 && (
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
                        <CardContent>
                            <ClipDisplay clips={clips} />
                        </CardContent>
                    </Card>
                </TabsContent>
            </Tabs>
        </div>
    );
}
