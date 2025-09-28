import { redirect } from "next/navigation";
import { DashboardClient } from "~/components/dashboard-client";
import { auth } from "~/server/auth";
import { db } from "~/server/db";

export default async function Dashboard() {
    const session = await auth();
    if (!session?.user?.id) {
        redirect("/login");
    }

    const userData = await db.user.findUnique({
        where: {
            id: session.user.id,
        },
        select: {
            uploadedFiles: {
                where: {
                    uploaded: true
                },
                select: {
                    id: true,
                    s3Key: true,
                    displayName: true,
                    status: true,
                    createdAt: true,
                    _count: {
                        select: {
                            clips: true
                        }
                    }
                }
            },
            clips: {
                orderBy: {
                    createdAt: "desc"
                }
            }
        },
    })

    const formattedFiles = userData?.uploadedFiles.map((file) => ({
        id: file.id,
        s3Key: file.s3Key,
        filename: file.displayName ?? "Unknown filename",
        status: file.status,
        clipsCount: file._count.clips,
        createdAt: file.createdAt,
    }));

    return (
        <DashboardClient uploadedFiles={formattedFiles ?? []} clips={userData?.clips ?? []} />
    );
}
