//getClips
import { AxiosError } from "axios";
import type { Clip } from "~/interfaces/clip";
import axiosClient from "~/lib/axiosClient";
export const getClips = async (): Promise<Clip[]> => {
    try {
        const res = await axiosClient.get(`/clips/me`);
        return res.data.data;
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        throw new Error(error.response?.data?.message || "Get clips failed");
    }
} 
