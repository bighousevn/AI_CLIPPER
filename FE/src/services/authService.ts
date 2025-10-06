import type { LoginResponse } from "~/interfaces/auth";
import axiosClient from "~/lib/axiosClient";
import { signupSchema, type LoginFormValues, type SignupFormValues } from "~/schemas/auth";

/**
 * Gọi API đăng nhập
 * Lưu accessToken vào localStorage
 */
export const login = async (data: LoginFormValues): Promise<LoginResponse> => {
    // Validate dữ liệu phía client bằng Zod trước khi gửi request
    const { email, password } = data;
    if (!email || !password) throw new Error("Email and password are required");

    const res = await axiosClient.post<LoginResponse>("/auth/login", data);
    const { access_token: accessToken, message: message } = res.data;
    localStorage.setItem("accessToken", accessToken);
    return res.data;
};

/**
 * Gọi API đăng ký
 */
export const signup = async (data: SignupFormValues) => {
    // Validate client-side trước khi gửi API
    const result = signupSchema.safeParse(data);
    if (!result.success) {
        const message = result.error.errors.map((e) => e.message).join(", ");
        throw new Error(message);
    }

    try {
        const res = await axiosClient.post("/auth/register", data);
        console.log("Signup response:", res.data);
        return res.data; // có thể trả về user info hoặc message
    } catch (err: any) {
        console.error("Signup error:", err);
        throw new Error(err.response?.data?.message || "Signup failed");
    }
};
/**
 * Gọi API logout, xóa accessToken localStorage
 */
export const logout = async () => {
    try {
        await axiosClient.post("/auth/logout");
    } catch (err) {
        console.error("Logout failed:", err);
    } finally {
        localStorage.removeItem("accessToken");
    }
};

/**
 * Refresh token từ cookie (HTTP-only)
 */
export const refreshToken = async () => {
    try {
        const res = await axiosClient.post("/auth/refresh", {}, { withCredentials: true });
        const newToken = res.data.accessToken;
        localStorage.setItem("accessToken", newToken);
        return newToken;
    } catch (err) {
        console.error("Refresh token failed:", err);
        localStorage.removeItem("accessToken");
        return null;
    }
};

export const verifyEmail = async (token: string) => {
    try {
        const res = await axiosClient.get(`/auth/verify-email?token=${token}`);
        return res.data; // có thể chứa message như "Email verified successfully"
    } catch (err: any) {
        throw new Error(err.response?.data?.message || "Xác thực thất bại");
    }
};
