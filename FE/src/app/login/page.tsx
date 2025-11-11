"use client";
import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { LoginForm } from "~/components/login-form";
import { useAuth } from "~/hooks/useAuth";

export default function Page() {
    const router = useRouter();
    const { user, loading } = useAuth();

    useEffect(() => {
        if (!loading && user) {
            router.push("/dashboard");
        }
    }, [loading, user, router]);

    if (loading) {
        return (
            <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
                <p>Đang kiểm tra phiên đăng nhập...</p>
            </div>
        );
    }

    if (user) return null;

    return (
        <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
            <div className="w-full max-w-sm">
                <LoginForm />
            </div>
        </div>
    );
}
