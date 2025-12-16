//getClips
import { AxiosError } from "axios";
import type { Clip } from "~/interfaces/Clip";
import axiosClient from "~/lib/axiosClient";
export const getClips = async (): Promise<Clip[]> => {
    try {
        const res = await axiosClient.get(`/clips/me`);
        console.log(res.data);
        return res.data.data;
    } catch (err) {
        const error = err as AxiosError<{ message?: string }>;
        throw new Error(error.response?.data?.message || "Get clips failed");
    }
} 
