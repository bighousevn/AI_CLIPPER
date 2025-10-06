// ~/lib/axiosClient.ts
import axios from "axios";

const axiosClient = axios.create({
    baseURL: process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080/api/v1/",
    headers: {
        "Content-Type": "application/json",
    },
    withCredentials: true, // gửi cookie (refresh_token)
});

// 🎯 Thêm accessToken vào request
axiosClient.interceptors.request.use(
    (config) => {
        const token =
            typeof window !== "undefined"
                ? localStorage.getItem("accessToken")
                : null;

        // Không gắn token cho các endpoint auth công khai
        const isAuthEndpoint =
            config.url?.includes("/auth/login") ||
            config.url?.includes("/auth/register") ||
            config.url?.includes("/auth/refresh-token");

        if (token && config.headers && !isAuthEndpoint) {
            config.headers.Authorization = `Bearer ${token}`;
        }

        return config;
    },
    (error) => Promise.reject(error)
);

// 🎯 Refresh token tự động khi access token hết hạn
axiosClient.interceptors.response.use(
    (response) => response,
    async (error) => {
        const originalRequest = error.config;

        // Nếu gặp lỗi 401 (Unauthorized) và chưa retry
        if (error.response?.status === 401 && !originalRequest._retry) {
            originalRequest._retry = true;

            try {
                // 👉 Gọi refresh token API (cookie được gửi kèm nhờ withCredentials)
                const res = await axios.post(
                    `${process.env.NEXT_PUBLIC_API_URL}/auth/refresh-token`,
                    {},
                    { withCredentials: true } // cookie refresh_token được gửi tự động
                );

                const newAccessToken = res.data.access_token;

                // Lưu lại access token mới
                localStorage.setItem("accessToken", newAccessToken);

                // Gắn lại token vào request gốc và thử gửi lại
                originalRequest.headers.Authorization = `Bearer ${newAccessToken}`;
                return axiosClient(originalRequest);
            } catch (err) {
                console.error("Refresh token failed:", err);
                // Nếu refresh token hết hạn hoặc lỗi -> xoá token + logout
                localStorage.removeItem("accessToken");
                window.location.href = "/login";
            }
        }

        return Promise.reject(error);
    }
);

export default axiosClient;
