"use server";
import { Toaster } from "sonner";
import NavHeader from "~/components/ui/nav-header";
export default async function DashboardLayout({ children }: { children: React.ReactNode }) {

    return (
        <div className="flex min-h-screen flex-col ">
            <NavHeader />
            <main className="flex-1 container py-6 mx-auto ">
                {children}
            </main>
            <Toaster />
        </div>
    )
}