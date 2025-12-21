import React from 'react';

import { MoveRight, Video, Sparkles, Wand2 } from "lucide-react";
import Link from 'next/link';
import { Button } from '~/components/ui/button';
import { Input } from '~/components/ui/input';
import { Badge } from '~/components/ui/badge';

export default function LandingPage() {
    return (
        <div className="min-h-screen bg-black text-white selection:bg-purple-500/30">
            {/* --- Navbar --- */}
            <nav className="fixed top-0 w-full z-50 border-b border-white/10 bg-black/50 backdrop-blur-md">
                <div className="max-w-7xl mx-auto px-4 h-16 flex items-center justify-between">
                    <div className="flex items-center gap-2 font-bold text-xl tracking-tighter">
                        <div className="w-8 h-8 bg-gradient-to-tr from-purple-600 to-blue-500 rounded-lg" />
                        OPUS<span className="text-purple-400">CLONE</span>
                    </div>
                    <div className="hidden md:flex gap-8 text-sm font-medium text-zinc-400">
                        <Link href="#" className="hover:text-white transition">Product</Link>
                        <Link href="#" className="hover:text-white transition">Features</Link>
                        <Link href="#" className="hover:text-white transition">Pricing</Link>
                    </div>
                    <Link href="/login">
                        <Button variant="ghost" className="hover:bg-zinc-800">Sign In</Button>
                    </Link>
                </div>
            </nav>

            {/* --- Hero Section --- */}
            <section className="pt-32 pb-20 px-4">
                <div className="max-w-4xl mx-auto text-center space-y-8">
                    <Badge variant="outline" className="border-purple-500/50 text-purple-400 px-4 py-1 rounded-full bg-purple-500/10">
                        ✨ Clip Anything is now in Beta
                    </Badge>

                    <h1 className="text-5xl md:text-7xl font-extrabold tracking-tight leading-[1.1]">
                        Biến video dài thành <br />
                        <span className="bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-pink-500 to-red-500">
                            Shorts triệu view
                        </span> bằng AI
                    </h1>

                    <p className="text-zinc-400 text-lg md:text-xl max-w-2xl mx-auto">
                        Công cụ AI đầu tiên có khả năng hiểu nội dung hình ảnh để tự động cắt ghép, thêm caption và tạo viral clip chỉ trong 1 cú click.
                    </p>

                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
                        <div className="relative w-full max-w-md group">
                            <Input
                                placeholder="Dán link YouTube, Drive hoặc Zoom tại đây..."
                                className="h-14 bg-zinc-900 border-zinc-800 rounded-xl pr-32 focus-visible:ring-purple-500"
                            />
                            <Link href="/login" className="absolute right-2 top-2">
                                <Button className="bg-purple-600 hover:bg-purple-700 rounded-lg">
                                    Bắt đầu ngay
                                </Button>
                            </Link>
                        </div>
                    </div>

                    <p className="text-xs text-zinc-500">Không cần thẻ tín dụng • 30 phút dùng thử miễn phí</p>
                </div>
            </section>

            {/* --- Feature Section (Simplified) --- */}
            <section className="py-20 bg-zinc-950">
                <div className="max-w-7xl mx-auto px-4 grid md:grid-cols-3 gap-8">
                    <FeatureCard
                        icon={<Video className="text-blue-400" />}
                        title="Clip Anything"
                        description="Tự động nhận diện các phân cảnh đắt giá nhất trong video của bạn."
                    />
                    <FeatureCard
                        icon={<Sparkles className="text-purple-400" />}
                        title="AI B-roll"
                        description="Thêm cảnh quay minh họa phù hợp với ngữ cảnh hoàn toàn tự động."
                    />
                    <FeatureCard
                        icon={<Wand2 className="text-pink-400" />}
                        title="Auto Curation"
                        description="Sắp xếp nội dung theo cấu trúc kể chuyện lôi cuốn nhất."
                    />
                </div>
            </section>

            {/* --- Final CTA --- */}
            <section className="py-32 text-center">
                <div className="max-w-3xl mx-auto px-4 space-y-6">
                    <h2 className="text-4xl font-bold">Sẵn sàng bùng nổ kênh của bạn?</h2>
                    <Link href="/login">
                        <Button size="lg" className="h-16 px-10 text-lg bg-white text-black hover:bg-zinc-200 rounded-full transition-all hover:scale-105">
                            Đăng nhập để trải nghiệm <MoveRight className="ml-2" />
                        </Button>
                    </Link>
                </div>
            </section>
        </div>
    );
}

function FeatureCard({ icon, title, description }: { icon: React.ReactNode, title: string, description: string }) {
    return (
        <div className="p-8 rounded-2xl border border-white/5 bg-zinc-900/50 hover:border-purple-500/30 transition group">
            <div className="mb-4 p-3 bg-black rounded-lg w-fit group-hover:scale-110 transition-transform">{icon}</div>
            <h3 className="text-xl font-semibold mb-2">{title}</h3>
            <p className="text-zinc-400 leading-relaxed">{description}</p>
        </div>
    );
}