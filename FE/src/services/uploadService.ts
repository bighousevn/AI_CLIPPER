// export async function uploadFile(file: File) {
//     const formData = new FormData();
//     formData.append("file", file);

//     const res = await fetch("http://localhost:8080/api/v1/upload", {
//         method: "POST",
//         body: formData,
//     });

//     if (!res.ok) {
//         const text = await res.text();
//         throw new Error(`Upload failed: ${text}`);
//     }

//     return await res.json();
// }
import axiosClient from "~/lib/axiosClient";
import type { AxiosError } from "axios";
import type { UploadFile } from "~/interfaces/uploadfile";
import type { ClipConfig } from "~/interfaces/ClipConfig";

export const uploadFile = async (file: File, config: ClipConfig) => {
    const formData = new FormData();

    // File upload
    formData.append("file", file);

    // ClipConfig → JSON
    formData.append("config", JSON.stringify(config));

    try {
        const res = await axiosClient.post("/upload", formData, {
            headers: {
                "Content-Type": "multipart/form-data",
            },
        });

        return res.data;
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        throw new Error(error.response?.data?.message || "Upload failed");
    }
};

export const processingFile = async (file: UploadFile) => {
    try {
        const processRes = await axiosClient.post(`/files/${file.id}/process`);
        return processRes.data; // return the response data if needed
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        if (error.response?.status === 400) {
            throw new Error(error.response?.data?.message || "Upload failed");
        }
        throw error;
    }
}
//getUploadedFiles
export const getUploadedFiles = async (): Promise<UploadFile[]> => {
    try {
        const res = await axiosClient.get("/files/me");
        return res.data;
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        throw new Error(error.response?.data?.message || "Upload failed");
    }
};