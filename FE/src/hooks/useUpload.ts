import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ClipConfigAPI } from "~/interfaces/clipConfig";
import { getUploadedFiles, uploadFile } from "~/services/uploadService";

export function useUploadClip() {
    const queryClient = useQueryClient();

    return useMutation({

        mutationFn: async (payload: { file: File; config: ClipConfigAPI }) => {
            return await uploadFile(payload.file, payload.config);
        },

        onSuccess: () => {
            queryClient.invalidateQueries({ queryKey: ["uploaded-files"] });
        },

        onError: (err) => {
            console.error("Upload failed:", err);
        }
    });
}
export function useUploadedFiles() {
    return useQuery(
        {
            queryKey: ["uploaded-files"],
            queryFn: getUploadedFiles
        });
}


export function useRefreshUploadedFiles() {
    const queryClient = useQueryClient();
    return () =>
        queryClient.invalidateQueries({
            queryKey: ["uploaded-files"],
        });
}
