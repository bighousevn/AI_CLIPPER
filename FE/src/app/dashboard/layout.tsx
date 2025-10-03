"use server";

import { cookies } from "next/headers";
import { redirect } from "next/navigation";
import { Toaster } from "sonner";
import NavHeader from "~/components/ui/nav-header";

interface User {
    email: string;
    credits: number;
}

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {


    let user: User = { email: "Duc", credits: 10 };


    return (
        <div className="flex min-h-screen flex-col ">
            <NavHeader credits={user.credits} email={user.email} />
            <main className="flex-1 container py-6 mx-auto ">
                {children}
            </main>
            <Toaster />
        </div>
    )
}