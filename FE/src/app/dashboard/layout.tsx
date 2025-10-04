"use server";

import { redirect } from "next/navigation";
import { Toaster } from "sonner";
import NavHeader from "~/components/ui/nav-header";
import { auth } from "~/server/auth";

export default async function DashboardLayout({ children }: { children: React.ReactNode }) {

    // const user = await db.user.findUniqueOrThrow({
    //     where: {
    //         id: session.user.id,
    //     },
    //     select: {
    //         email: true,
    //         credits: true,
    //     },
    // });
    return (
        <div className="flex min-h-screen flex-col ">
            {/* <NavHeader credits={user.credits} email={user.email} /> */}
            <main className="flex-1 container py-6 mx-auto ">
                {children}
            </main>
            <Toaster />
        </div>
    )
}