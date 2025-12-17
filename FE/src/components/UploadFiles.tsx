import React, { useState } from "react";
import { Button } from "./ui/button";
import { Loader2 } from "lucide-react";
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from "./ui/table";
import { Badge } from "./ui/badge";


const UploadFiles = () => {
    const [uploadedFiles, setUploadedFiles] = useState<any[]>([]); // thay any bằng type thật
    const [refreshing, setRefreshing] = useState(false);

    const handleRefresh = () => {
        setRefreshing(true);
        // giả lập refresh
        setTimeout(() => setRefreshing(false), 1000);
    };

    return (
        <div className="pt-6">
            {uploadedFiles.length > 0 ? (
                <>
                    <div className="mb-2 flex items-center justify-between">
                        <h3 className="text-md mb-2 font-medium">Queue status</h3>
                        <Button variant="outline" size="sm" onClick={handleRefresh} disabled={refreshing}>
                            {refreshing && <Loader2 className="mr-2 h-4 w-4 animate-spin" />}
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
                                        <TableCell className="max-w-xs truncate font-medium">{item.filename}</TableCell>
                                        <TableCell className="text-muted-foreground text-sm">
                                            {new Date(item.createdAt).toLocaleDateString()}
                                        </TableCell>
                                        <TableCell>
                                            {item.status === "queued" && <Badge variant="outline">Queued</Badge>}
                                            {item.status === "processing" && <Badge variant="outline">Processing</Badge>}
                                            {item.status === "processed" && <Badge variant="outline">Processed</Badge>}
                                            {item.status === "no credits" && <Badge variant="destructive">No credits</Badge>}
                                            {item.status === "failed" && <Badge variant="destructive">Failed</Badge>}
                                        </TableCell>
                                        <TableCell>
                                            {item.clipsCount > 0 ? (
                                                <span>
                                                    {item.clipsCount} clip{item.clipsCount !== 1 ? "s" : ""}
                                                </span>
                                            ) : (
                                                <span className="text-muted-foreground">No clips yet</span>
                                            )}
                                        </TableCell>
                                    </TableRow>
                                ))}
                            </TableBody>
                        </Table>
                    </div>
                </>
            ) : (
                <div className="mb-2">
                    <h3 className="text-md font-medium">No files uploaded yet</h3>
                </div>
            )}
        </div>
    );
};

export default UploadFiles;
