import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
    const token = request.cookies.get("access_token")?.value;

    const { pathname } = request.nextUrl;

    // // Nếu đang login mà vào /login hoặc /register → redirect dashboard
    // if (token && (pathname.startsWith("/login") || pathname.startsWith("/register"))) {
    //     return NextResponse.redirect(new URL("/dashboard", request.url));
    // }

    // // Nếu chưa login mà vào /dashboard → redirect login
    // if (!token && pathname.startsWith("/dashboard")) {
    //     return NextResponse.redirect(new URL("/login", request.url));
    // }

    return NextResponse.next();
}

// Áp dụng cho các route cụ thể
export const config = {
    matcher: ["/login", "/register", "/dashboard/:path*"],
};
