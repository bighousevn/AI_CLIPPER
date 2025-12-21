import { Suspense } from "react";
import ResetPasswordClient from "./resetPasswordClient";
export default function Page() {
    return (
        <Suspense fallback={<div>Loading...</div>}>
            <ResetPasswordClient />
        </Suspense>
    );
}
