"use client";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";
import { Loader2 } from "lucide-react";
import type { UploadFile } from "~/interfaces/uploadfile";

export function UploadedFilesTable({ files }: { files: UploadFile[] }) {
    if (!files.length)
        return <div className="mb-2 mt-4 text-md font-medium">No files uploaded yet</div>;

    return (
        <div className="max-h-[300px] overflow-auto rounded-md border mt-5">
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
                    {files.map((item) => (
                        <TableRow key={item.id}>
                            <TableCell className="max-w-xs truncate font-medium">{item.file_name}</TableCell>
                            <TableCell className="text-muted-foreground text-sm">
                                {new Date(item.created_at).toLocaleString()}
                            </TableCell>
                            <TableCell className="text-muted-foreground text-sm">
                                {item.status === "queued" && <Badge variant="outline">Queued</Badge>}
                                {item.status === "processing" && (
                                    <div className="flex items-center">
                                        <Loader2 className="mr-2 h-4 w-4 animate-spin" /> Processing
                                    </div>
                                )}
                                {item.status === "success" && <Badge className="bg-green-600">Success</Badge>}
                                {item.status === "no credits" && <Badge variant="destructive">No credits</Badge>}
                                {item.status === "failed" && <Badge variant="destructive">Failed</Badge>}
                            </TableCell>
                            <TableCell>
                                {item.clip_count > 0 ? (
                                    `${item.clip_count} clip${item.clip_count !== 1 ? "s" : ""}`
                                ) : (
                                    <span className="text-muted-foreground">No clips yet</span>
                                )}
                            </TableCell>
                        </TableRow>
                    ))}
                </TableBody>
            </Table>
        </div>
    );
}
