import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";
import { getRefreshToken } from "./services/authService";

export function middleware(request: NextRequest) {
    const { pathname } = request.nextUrl;

    // Refresh token được lưu trong cookie (HTTP-only)
    const refreshToken = request.cookies.get("refresh_token")?.value;

    // Nếu đã login mà cố truy cập /login hoặc /register → redirect dashboard
    if (refreshToken && ["/login", "/register"].includes(pathname)) {
        const res = getRefreshToken();
        if (res !== null) return NextResponse.redirect(new URL("/dashboard", request.url));
    }

    // Ngược lại, cho phép đi tiếp
    return NextResponse.next();
}

// Áp dụng cho các route cần kiểm tra
export const config = {
    matcher: ["/login", "/register"],
};
