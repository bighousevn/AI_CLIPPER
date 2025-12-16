import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";
import type { ClipConfig } from "~/interfaces/ClipConfig";
import { getUploadedFiles, uploadFile } from "~/services/uploadService";

export function useUploadClip() {
    const queryClient = useQueryClient();

    return useMutation({

        mutationFn: async (payload: { file: File; config: ClipConfig }) => {
            return await uploadFile(payload.file, payload.config);
        },

        onSuccess: () => {
            // Tự refresh danh sách uploadedFiles
            queryClient.invalidateQueries({ queryKey: ["uploaded-files"] });
        },

        onError: (err) => {
            console.error("Upload failed:", err);
        }
    });
}
//useUploadedFiles
export function useUploadedFiles() {
    return useQuery(
        {
            queryKey: ["uploaded-files"],
            queryFn: getUploadedFiles
        });
}