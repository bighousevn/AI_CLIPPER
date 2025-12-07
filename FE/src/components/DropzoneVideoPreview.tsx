import { useMemo } from "react";
import { useDropzoneContext } from "./ui/shadcn-io/dropzone";
import { cn } from "~/lib/utils";

export const DropzoneVideoPreview = ({ className }: { className?: string }) => {
    const { src } = useDropzoneContext();

    const file = src?.[0];
    const videoUrl = useMemo(() => {
        if (!file) return null;
        return URL.createObjectURL(file);
    }, [file]);

    if (!file) {
        return null; // return sau khi mọi hook đã được chạy
    }

    return (
        <div className={cn("w-full flex flex-col space-y-3", className)}>
            <video
                src={videoUrl!}
                controls
                className="w-full rounded-lg border shadow-sm max-h-[320px] object-contain"
            />

            <p className="text-sm font-medium text-left truncate w-full">
                {file.name}
            </p>

            <p className="text-xs text-muted-foreground text-center">
                Click or drag to replace video
            </p>
        </div>
    );
};
