"use client";

import { useRouter } from "next/navigation";
import { useEffect } from "react";
import { SignupForm } from "~/components/signup-form";
import { useAuth } from "~/hooks/useAuth";

export default function Page() {
  const router = useRouter();
  const { user, loading } = useAuth();

  useEffect(() => {
    // 🔒 Nếu đã có user → chuyển đến dashboard
    if (!loading && user) {
      router.push("/dashboard");
    }
  }, [loading, user, router]);

  // ⏳ Hiển thị loading trong khi kiểm tra đăng nhập
  if (loading) {
    return (
      <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
        <p>Đang kiểm tra phiên đăng nhập...</p>
      </div>
    );
  }

  // 🚫 Nếu đã login thì không hiển thị gì (đang redirect)
  if (user) return null;

  // 🧾 Nếu chưa đăng nhập → hiển thị form đăng ký
  return (
    <div className="flex min-h-svh w-full items-center justify-center p-6 md:p-10">
      <div className="w-full max-w-sm">
        <SignupForm />
      </div>
    </div>
  );
}
