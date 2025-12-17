"use client";

import { useEffect, useState } from "react";
import axiosClient from "~/lib/axiosClient";
import type { UserResponse } from "~/interfaces/auth";

export function useAuth() {
    const [user, setUser] = useState<UserResponse | null>(null);
    const [loading, setLoading] = useState(true);

    const fetchUser = async () => {
        const token = localStorage.getItem("accessToken");
        if (!token) {
            setLoading(false);
            return; // ❌ Không tự refresh ở đây
        }

        try {
            const res = await axiosClient.get("/users/me");
            setUser(res.data);
        } catch (err) {
            console.error(" Lấy thông tin user thất bại:", err);
            localStorage.removeItem("accessToken");
        } finally {
            setLoading(false);
        }
    };

    useEffect(() => {
        fetchUser();
    }, []);

    return { user, loading, fetchUser };
}
