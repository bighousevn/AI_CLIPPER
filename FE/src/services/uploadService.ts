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

export const uploadFile = async (file: File) => {
    const formData = new FormData();
    formData.append("file", file);

    try {
        const res = await axiosClient.post("/upload", formData, { headers: { "Content-Type": "multipart/form-data" } });
        return res.data; // return the response data if needed
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        throw new Error(error.response?.data?.message || "Upload failed");
    }
};