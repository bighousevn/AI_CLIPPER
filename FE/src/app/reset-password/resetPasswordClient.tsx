"use client";

import { useState } from "react";
import { useRouter, useSearchParams } from "next/navigation";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { zodResolver } from "@hookform/resolvers/zod";
import { useMutation } from "@tanstack/react-query";

import { Loader2 } from "lucide-react";
import {
    Card,
    CardHeader,
    CardTitle,
    CardDescription,
    CardContent,
} from "~/components/ui/card";
import { Label } from "~/components/ui/label";
import { Input } from "~/components/ui/input";
import { Button } from "~/components/ui/button";
import axiosClient from "~/lib/axiosClient";

const resetSchema = z
    .object({
        password: z.string().min(6, "Password must be at least 6 characters"),
        confirmPassword: z.string().min(6, "Password must be at least 6 characters"),
    })
    .refine((data) => data.password === data.confirmPassword, {
        message: "Passwords do not match",
        path: ["confirmPassword"],
    });

type ResetSchema = z.infer<typeof resetSchema>;

export default function ResetPasswordClient() {
    const params = useSearchParams();
    const token = params.get("token") ?? "";
    const [message, setMessage] = useState<string | null>(null);
    const router = useRouter();

    const form = useForm<ResetSchema>({
        resolver: zodResolver(resetSchema),
        defaultValues: { password: "", confirmPassword: "" },
    });

    const mutation = useMutation({
        mutationFn: async (data: ResetSchema) => {
            const res = await axiosClient.post("/auth/reset-password", {
                token,
                password: data.password,
            });
            return res.data;
        },
        onSuccess: () => {
            setMessage("Your password has been reset successfully.");
            router.push("/login");
        },
        onError: () => {
            setMessage("Something went wrong. Try again later.");
        },
    });

    const onSubmit = (data: ResetSchema) => {
        setMessage(null);
        mutation.mutate(data);
    };

    return (
        <div className="flex justify-center items-center min-h-screen bg-gray-100">
            <Card className="w-full max-w-md shadow-lg">
                <CardHeader>
                    <CardTitle>Reset Password</CardTitle>
                    <CardDescription>
                        Enter your new password to reset your account.
                    </CardDescription>
                </CardHeader>

                <CardContent>
                    <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
                        <div>
                            <Label>New Password</Label>
                            <Input
                                type="password"
                                {...form.register("password")}
                                placeholder="******"
                                disabled={mutation.isSuccess}
                            />
                            {form.formState.errors.password && (
                                <p className="text-red-500 text-sm">
                                    {form.formState.errors.password.message}
                                </p>
                            )}
                        </div>

                        <div>
                            <Label>Confirm New Password</Label>
                            <Input
                                type="password"
                                {...form.register("confirmPassword")}
                                placeholder="******"
                                disabled={mutation.isSuccess}
                            />
                            {form.formState.errors.confirmPassword && (
                                <p className="text-red-500 text-sm">
                                    {form.formState.errors.confirmPassword.message}
                                </p>
                            )}
                        </div>

                        {message && (
                            <p
                                className={`text-sm ${mutation.isSuccess ? "text-green-600" : "text-red-500"
                                    }`}
                            >
                                {message}
                            </p>
                        )}

                        <Button
                            type="submit"
                            disabled={mutation.isPending || mutation.isSuccess}
                            className="w-full"
                        >
                            {mutation.isPending && (
                                <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                            )}
                            {mutation.isSuccess ? "Password Reset" : "Reset Password"}
                        </Button>
                    </form>
                </CardContent>
            </Card>
        </div>
    );
}
