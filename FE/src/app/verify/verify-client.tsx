"use client";

import { useSearchParams } from "next/navigation";
import { useEffect, useState } from "react";
import { CheckCircle2, XCircle } from "lucide-react";
import { Button } from "~/components/ui/button";

export default function VerifyClient() {
    const searchParams = useSearchParams();
    const token = searchParams.get("token") ?? "";
    const [status, setStatus] = useState<"idle" | "ok" | "fail">("idle");

    useEffect(() => {
        console.log(token);
        if (!token) return setStatus("fail");

        (async () => {
            try {
                const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/verify-email?token=${token}`, {
                    method: "GET",
                    credentials: "include",
                });
                if (!res.ok) throw new Error();
                setStatus("ok");
            } catch {
                setStatus("fail");
            }
        })();
    }, [token]);

    if (status === "idle") return <p className="text-center mt-10">Đang xác minh...</p>;

    if (status === "ok")
        return (
            <div className="flex flex-col items-center justify-center mt-20">
                <CheckCircle2 className="text-green-500 w-12 h-12 mb-2" />
                <p className="text-xl font-semibold">Xác minh thành công!</p>
                <Button asChild className="mt-4">
                    <a href="/login">Đăng nhập ngay</a>
                </Button>
            </div>
        );

    return (
        <div className="flex flex-col items-center justify-center mt-20">
            <XCircle className="text-red-500 w-12 h-12 mb-2" />
            <p className="text-xl font-semibold">Liên kết không hợp lệ hoặc đã hết hạn</p>
        </div>
    );
}
