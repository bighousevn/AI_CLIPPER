// ~/hooks/useAuth.ts
"use client";

import { useEffect, useState } from "react";
import { useRouter } from "next/navigation";
import axiosClient from "~/lib/axiosClient";
import type { UserResponse } from "~/interfaces/auth";



export function useAuth() {
    const [user, setUser] = useState<UserResponse | null>(null);
    const [loading, setLoading] = useState(true);
    const router = useRouter();

    const fetchUser = async () => {
        const token = localStorage.getItem("accessToken");

        // ❌ Chưa có access token → thử refresh
        if (!token) {
            try {
                const res = await axiosClient.post("/auth/refresh-token", {});
                const newAccessToken = res.data.access_token;
                localStorage.setItem("accessToken", newAccessToken);
            } catch {
                setLoading(false);
                router.push("/login");
                return;
            }
        }

        // ✅ Có token → gọi API lấy thông tin user
        try {
            const res = await axiosClient.get("/users/me");
            setUser(res.data);
        } catch (err) {
            console.error("❌ Lấy thông tin user thất bại:", err);
            localStorage.removeItem("accessToken");
            router.push("/login");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUser();
    }, []);

    return { user, loading, fetchUser };
}
