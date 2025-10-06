"use client";

import { useEffect, useState } from "react";
import { useSearchParams, useRouter } from "next/navigation";
import { verifyEmail } from "~/services/authService";
import { Card, CardHeader, CardTitle, CardContent } from "~/components/ui/card";
import { Loader2, CheckCircle2, XCircle } from "lucide-react";

export default function VerifyPage() {
    const router = useRouter();
    const searchParams = useSearchParams();
    const [status, setStatus] = useState<"waiting" | "verifying" | "success" | "error">("waiting");
    const [message, setMessage] = useState<string>("");

    useEffect(() => {
        const token = searchParams.get("token");

        // Nếu người dùng mở từ link trong email
        if (token) {
            setStatus("verifying");
            setMessage("Đang xác thực email của bạn...");

            verifyEmail(token)
                .then(() => {
                    setStatus("success");
                    setMessage("Email của bạn đã được xác thực thành công!");
                    setTimeout(() => router.push("/login"), 3000);
                })
                .catch((err) => {
                    setStatus("error");
                    setMessage(err.message || "Token không hợp lệ hoặc đã hết hạn.");
                });
        } else {
            // Trường hợp người dùng chỉ mới đăng ký xong (chưa click link email)
            setStatus("waiting");
            setMessage("Vui lòng kiểm tra email của bạn để xác thực tài khoản.");
        }
    }, [searchParams, router]);

    return (
        <div className="flex items-center justify-center min-h-screen bg-background">
            <Card className="w-[400px] p-6 text-center shadow-md">
                <CardHeader>
                    <CardTitle className="text-2xl font-semibold">
                        {status === "waiting" && "Xác thực email"}
                        {status === "verifying" && "Đang xác thực..."}
                        {status === "success" && "Thành công"}
                        {status === "error" && "Lỗi xác thực"}
                    </CardTitle>
                </CardHeader>

                <CardContent className="flex flex-col items-center justify-center space-y-3">
                    {status === "waiting" && (
                        <>
                            <Loader2 className="w-6 h-6 animate-spin text-muted-foreground" />
                            <p className="text-gray-600 text-sm">{message}</p>
                        </>
                    )}

                    {status === "verifying" && (
                        <>
                            <Loader2 className="w-6 h-6 animate-spin text-blue-500" />
                            <p className="text-gray-700 text-sm">{message}</p>
                        </>
                    )}

                    {status === "success" && (
                        <>
                            <CheckCircle2 className="w-8 h-8 text-green-500" />
                            <p className="text-green-600 text-sm">{message}</p>
                            <p className="text-xs text-muted-foreground">(Sẽ chuyển hướng đến trang đăng nhập...)</p>
                        </>
                    )}

                    {status === "error" && (
                        <>
                            <XCircle className="w-8 h-8 text-red-500" />
                            <p className="text-red-600 text-sm">{message}</p>
                        </>
                    )}
                </CardContent>
            </Card>
        </div>
    );
}
