"use client";

import { cn } from "~/lib/utils";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "./ui/card";
import { Input } from "./ui/input";
import { Label } from "./ui/label";
import { Button } from "./ui/button";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import Link from "next/link";
import { loginSchema, type LoginFormValues } from "~/schemas/auth";
import { useRouter } from "next/navigation";
import { login } from "~/services/authService";
import { useMutation } from "@tanstack/react-query";
import type { AxiosError } from "axios";

export function LoginForm({
    className,
    ...props
}: React.ComponentProps<"div">) {
    const router = useRouter();

    const {
        register,
        handleSubmit,
        formState: { errors },
    } = useForm<LoginFormValues>({
        resolver: zodResolver(loginSchema),
    });

    const mutation = useMutation({
        mutationFn: async (data: LoginFormValues) => await login(data),
        onSuccess: () => {
            router.push("/dashboard");
        },
        onError: () => { },
    });

    const onSubmit = (data: LoginFormValues) => {
        mutation.mutate(data);
    };

    return (
        <div className={cn("flex flex-col gap-6", className)} {...props}>
            <Card className="border border-border shadow-sm">
                <CardHeader className="space-y-2">
                    <CardTitle className="text-2xl font-semibold">Login</CardTitle>
                    <CardDescription>
                        Enter your credentials to access your dashboard
                    </CardDescription>
                </CardHeader>

                <CardContent>
                    <form onSubmit={handleSubmit(onSubmit)} className="space-y-6" noValidate>
                        <div className="space-y-3">
                            <Label htmlFor="email">Email</Label>
                            <Input
                                id="email"
                                type="email"
                                placeholder="m@example.com"
                                {...register("email")}
                            />
                            {errors.email && (
                                <p className="text-sm text-red-500">{errors.email.message}</p>
                            )}
                        </div>

                        <div className="space-y-3">
                            <div className="flex justify-between items-center">
                                <Label htmlFor="password">Password</Label>
                                <Link
                                    href="/forgot-password"
                                    className="text-xs text-primary underline-offset-4 hover:underline"
                                >
                                    Forgot?
                                </Link>
                            </div>
                            <Input
                                id="password"
                                type="password"
                                autoComplete="current-password"
                                {...register("password")}
                            />
                            {errors.password && (
                                <p className="text-sm text-red-500">{errors.password.message}</p>
                            )}
                        </div>

                        {/* Error từ mutation */}
                        {mutation.isError && (
                            <p className="text-sm text-red-600 bg-red-50 border border-red-200 rounded-md p-3">
                                {(mutation.error as AxiosError<{ message?: string }>)?.response?.data
                                    ?.message || "Invalid email or password."}
                            </p>
                        )}

                        <Button
                            type="submit"
                            className="w-full"
                            disabled={mutation.isPending}
                        >
                            {mutation.isPending ? "Logging in..." : "Login"}
                        </Button>

                        <div className="text-center text-sm pt-2">
                            Don’t have an account?{" "}
                            <Link
                                href="/signup"
                                className="font-medium underline underline-offset-4 text-primary"
                            >
                                Sign up
                            </Link>
                        </div>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
