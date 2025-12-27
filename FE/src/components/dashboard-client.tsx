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
import { useEffect, useState } from "react";
import { Dropzone, DropzoneContent, DropzoneEmptyState } from "./ui/shadcn-io/dropzone";
import { Badge } from "./ui/badge";
import { ClipDisplay } from "./clip-display";
import { DropzoneVideoPreview } from "./DropzoneVideoPreview";
import { ClipConfigSchema, transformToApiData } from "~/schemas/clipConfigSchema";
import { zodResolver } from "@hookform/resolvers/zod";
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from "./ui/form";
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from "./ui/select";
import { useForm } from "react-hook-form";
import { Checkbox } from "./ui/checkbox";
import { useRefreshUploadedFiles, useUploadClip } from "~/hooks/useUpload";
import type { ClipConfig } from "~/interfaces/clipConfig";
import type { Clip } from "~/interfaces/clip";
import type { UploadFile } from "~/interfaces/uploadfile";
import { UploadedFilesTable } from "./uploaded-files-table";
import { useClips } from "~/hooks/useClip";


export function DashboardClient({
    uploadedFiles,
    clips: initialClips,
}: {
    uploadedFiles: UploadFile[];
    clips: Clip[];

}) {

    const suggestedPrompts = [
        "Funny moments",
        "Key takeaways",
        "Interesting questions",
        "Actionable advice",
        "Viral-worthy clips"
    ];
    const refreshUploadedFiles = useRefreshUploadedFiles();
    const { data: clipsData, refetch: refetchClips } = useClips();
    const clips = clipsData ?? initialClips;

    useEffect(() => {
        refetchClips();
        const token = localStorage.getItem("accessToken");
        document.cookie = `access_token=${token}; path=/; samesite=none; secure;`;

        const es = new EventSource(`${process.env.NEXT_PUBLIC_API_URL}/events`, {
            withCredentials: true,
        });

        es.addEventListener("video_status", () => {
            refreshUploadedFiles();
            refetchClips();
            console.log("video_status");
            setTimeout(() => {
                refreshUploadedFiles();
                refetchClips();
            }, 1000);
        });

        return () => es.close();
    }, []);

    const [files, setFiles] = useState<File[]>([]);

    const form = useForm<ClipConfig>({
        resolver: zodResolver(ClipConfigSchema),
        defaultValues: {
            prompt: "",
            clip_count: 1,
            aspect_ratio: "9:16",
            subtitle: false,
        },
    });

    const handleDrop = (acceptedFiles: File[]) => {
        if (acceptedFiles.length > 0) {
            setFiles(acceptedFiles);
        }
    };
    const { mutate, isPending } = useUploadClip();

    const handleUpload = async () => {

        const isValid = await form.trigger();
        if (!isValid || files.length === 0) return;

        try {
            const item = files[0] as File;
            const rawValues = form.getValues();
            const apiData = transformToApiData(rawValues);
            await mutate({ file: item, config: apiData });
        } finally {
        }
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
                                disabled={isPending}
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
                                <Card className="mt-6 animate-in fade-in slide-in-from-bottom-4 duration-300"  >
                                    <CardHeader>
                                        <CardTitle>Video Configuration</CardTitle>
                                        <CardDescription>
                                            Customize how clips will be generated.
                                        </CardDescription>
                                    </CardHeader>

                                    <CardContent  >
                                        <Form {...form}  >
                                            <form className="space-y-4">
                                                <fieldset disabled={isPending} className="space-y-4 opacity-100">
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
                                                                <div className="flex flex-wrap gap-2 pt-2">
                                                                    {suggestedPrompts.map((promptSuggestion) => (
                                                                        <Badge
                                                                            key={promptSuggestion}
                                                                            variant="outline"
                                                                            className="cursor-pointer"
                                                                            onClick={() => {
                                                                                const currentValue = form.getValues("prompt");
                                                                                const newValue = currentValue
                                                                                    ? `${currentValue}\n- ${promptSuggestion}`
                                                                                    : `- ${promptSuggestion}`;
                                                                                form.setValue("prompt", newValue, { shouldValidate: true });
                                                                            }}
                                                                        >
                                                                            {promptSuggestion}
                                                                        </Badge>
                                                                    ))}
                                                                </div>
                                                                <FormMessage />
                                                            </FormItem>
                                                        )}
                                                    />

                                                    {/* Number of clips */}
                                                    <FormField
                                                        control={form.control}
                                                        name="clip_count"
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
                                                        name="aspect_ratio"
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
                                                                        <SelectItem value="4:3">4:3 (Instagram)</SelectItem>
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
                                                </fieldset>
                                            </form>
                                        </Form>
                                    </CardContent>
                                </Card>
                            )}

                            <div className="flex items-start justify-between">

                                <Button className="mt-4" disabled={isPending || files.length === 0} onClick={handleUpload}>
                                    {isPending ? <>
                                        <Loader2 className="mr-2 h-4 w-4  animate-spin" /> Uploading</> : "Upload and Generate Clips"}
                                </Button>
                            </div>
                            {/* <UploadFiles /> */}
                            <UploadedFilesTable files={uploadedFiles} />
                        </CardContent>

                    </Card>
                </TabsContent>

                <TabsContent value="my-clips" forceMount
                    className="data-[state=inactive]:hidden"
                    overflow-y-scroll
                >
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
