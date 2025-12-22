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
                        AI <span className="text-purple-400">CLIPPER</span>
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
                        Turn long videos into <br />
                        <span className="bg-clip-text text-transparent bg-gradient-to-r from-purple-400 via-pink-500 to-red-500">
                            viral Shorts
                        </span> with AI
                    </h1>

                    <p className="text-zinc-400 text-lg md:text-xl max-w-2xl mx-auto">
                        The first AI tool that understands visual content to automatically clip, add captions, and create viral clips in just one click.
                    </p>

                    <div className="flex flex-col sm:flex-row items-center justify-center gap-4 pt-4">
                        <div className="relative w-full max-w-md group">
                            <Link href="/dashboard" >
                                <Button className="bg-purple-600 hover:bg-purple-700 rounded-lg">
                                    Get Started
                                </Button>
                            </Link>
                        </div>
                    </div>

                    <p className="text-xs text-zinc-500">No credit card required • Free trial available</p>
                </div>
            </section>

            {/* --- Feature Section (Simplified) --- */}
            <section className="py-20 bg-zinc-950">
                <div className="max-w-7xl mx-auto px-4 grid md:grid-cols-3 gap-8">
                    <FeatureCard
                        icon={<Video className="text-blue-400" />}
                        title="Smart Scene Cutting"
                        description="Automatically identify and extract the most engaging highlights from your video."
                    />
                    <FeatureCard
                        icon={<Sparkles className="text-purple-400" />}
                        title="Auto Subtitles"
                        description="Generate accurate captions with customizable styles to boost engagement."
                    />
                    <FeatureCard
                        icon={<Wand2 className="text-pink-400" />}
                        title="Smart Reframe"
                        description="Automatically resize videos to 9:16 vertical format while keeping the subject in focus."
                    />
                </div>
            </section>

            {/* --- Final CTA --- */}
            <section className="py-32 text-center">
                <div className="max-w-3xl mx-auto px-4 space-y-6">
                    <h2 className="text-4xl font-bold">Ready to grow your channel?</h2>
                    <Link href="/login">
                        <Button size="lg" className="h-16 px-10 text-lg bg-white text-black hover:bg-zinc-200 rounded-full transition-all hover:scale-105">
                            Sign in to start <MoveRight className="ml-2" />
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