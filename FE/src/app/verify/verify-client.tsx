'use client';

import { useSearchParams, useRouter } from 'next/navigation';
import { useEffect, useState } from 'react';

export default function VerifyClient() {
    const searchParams = useSearchParams();
    const token = searchParams.get('token') ?? '';
    const [status, setStatus] = useState<'idle' | 'ok' | 'fail'>('idle');

    useEffect(() => {
        if (!token) { setStatus('fail'); return; }
        // Gọi API xác minh
        (async () => {
            try {
                const res = await fetch(`${process.env.NEXT_PUBLIC_API_URL}/auth/verify-email?token=${token}`, {
                    method: 'POST',
                    credentials: 'include',
                });
                if (!res.ok) throw new Error('verify failed');
                setStatus('ok');
            } catch {
                setStatus('fail');
            }
        })();
    }, [token]);

    if (status === 'idle') return <p>Đang xác minh...</p>;
    if (status === 'ok') return <p>Xác minh thành công! Bạn có thể đăng nhập.</p>;
    return <p>Liên kết xác minh không hợp lệ hoặc đã hết hạn.</p>;
}
