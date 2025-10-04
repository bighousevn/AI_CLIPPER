import axiosClient from "~/lib/axiosClient";


export type LoginResponse = {
    message: string;
    token: string;
};

export async function login(email: string, password: string): Promise<LoginResponse> {
    const res = await axiosClient.post<LoginResponse>("/auth/login", {
        email,
        password,
    });

    return res.data;
}

export async function logout() {
    localStorage.removeItem("accessToken");
    localStorage.removeItem("refreshToken");
    window.location.href = "/login";
}

export async function getProfile() {
    const res = await axiosClient.get("/users/me");
    return res.data;
}
