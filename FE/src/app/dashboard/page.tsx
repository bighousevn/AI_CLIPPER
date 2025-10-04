

export default async function Dashboard() {

    // const userData = await db.user.findUnique({
    //     where: {
    //         id: session.user.id,
    //     },
    //     select: {
    //         uploadedFiles: {
    //             where: {
    //                 uploaded: true
    //             },
    //             select: {
    //                 id: true,
    //                 s3Key: true,
    //                 displayName: true,
    //                 status: true,
    //                 createdAt: true,
    //                 _count: {
    //                     select: {
    //                         clips: true
    //                     }
    //                 }
    //             }
    //         },
    //         clips: {
    //             orderBy: {
    //                 createdAt: "desc"
    //             }
    //         }
    //     },
    // })

    // const formattedFiles = userData?.uploadedFiles.map((file) => ({
    //     id: file.id,
    //     s3Key: file.s3Key,
    //     filename: file.displayName ?? "Unknown filename",
    //     status: file.status,
    //     clipsCount: file._count.clips,
    //     createdAt: file.createdAt,
    // }));

    return (

        <>nodata</>        // <DashboardClient uploadedFiles={formattedFiles ?? []} clips={userData?.clips ?? []} />
    );
}
