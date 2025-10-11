"use client";
import { cn } from "~/lib/utils";
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from "./ui/card";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Button } from "./ui/button";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useState } from "react";
import Link from "next/link";
import { loginSchema, signupSchema, type LoginFormValues, type SignupFormValues } from "~/schemas/auth";
import { signIn } from "next-auth/react";
import { useRouter } from "next/navigation";
import axiosClient from "~/lib/axiosClient";
import { login } from "~/services/authService";
import type { AxiosError } from "axios";

export function LoginForm({
    className,
    ...props
}: React.ComponentProps<"div">) {

    const [error, setError] = useState<string | null>(null);
    const [submitting, setSubmitting] = useState(false);
    const router = useRouter();

    const { register, handleSubmit, formState: { errors } } = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
    })
    const onSubmit = async (data: LoginFormValues) => {
        try {
            setSubmitting(true);
            setError(null);
            const res = await login(data);


            // Redirect
            router.push("/dashboard");
        } catch (err) {
            const error = err as AxiosError<{ message?: string }>;
            throw new Error(error.response?.data?.message || "Invalid email or password.");
        } finally {
            setSubmitting(false);
        }
    };

    return (
        <div className={cn("flex flex-col gap-6", className)} {...props}>
            <Card>
                <CardHeader>
                    <CardTitle>Login</CardTitle>
                    <CardDescription>
                        Enter your email below to sign up to your account
                    </CardDescription>
                </CardHeader>
                <CardContent>
                    <form onSubmit={handleSubmit(onSubmit)} noValidate>
                        <div className="flex flex-col gap-6">
                            <div className="grid gap-3">
                                <Label htmlFor="email">Email</Label>
                                <Input
                                    id="email"
                                    type="email"
                                    placeholder="m@example.com"
                                    required
                                    {...register("email")}
                                />
                                {errors.email && <p className="text-red-500">{errors.email.message}</p>}
                            </div>
                            <div className="grid gap-3">
                                <div className="flex items-center">
                                    <Label htmlFor="password">Password</Label>

                                </div>
                                <Input id="password" type="password" required
                                    autoComplete="current-password"
                                    {...register("password")} />
                                {errors.password && <p className="text-red-500">{errors.password.message}</p>}
                            </div>
                            <div className="flex flex-col gap-3">
                                {error && <p className="text-red-500 rounded-md bg-red-50 p-3 text-sm">{error}</p>}

                                <Button type="submit" className="w-full" disabled={submitting}>
                                    {submitting ? "Logging in..." : "Login"}
                                </Button>

                            </div>
                        </div>
                        <div className="mt-4 text-center text-sm">
                            Don&apos;t have an account?{" "}
                            <Link href="/signup" className="underline underline-offset-4">
                                Sign up
                            </Link>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    )
}
