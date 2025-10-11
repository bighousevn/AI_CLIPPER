import { Suspense } from 'react';
import VerifyClient from './verify-client';

export const dynamic = 'force-dynamic';
export const revalidate = 0;

export default function Page() {
    return (
        <Suspense fallback={<p>Đang xác minh...</p>}>
            <VerifyClient />
        </Suspense>
    );
}
