"use client";

import { useState } from "react";
import { z } from "zod";
import { useForm } from "react-hook-form";
import { useMutation } from "@tanstack/react-query";

import { Loader2, AlertCircle, CheckCircle2 } from "lucide-react";
import {
    Card,
    CardContent,
    CardDescription,
    CardHeader,
    CardTitle,
} from "~/components/ui/card";
import {
    Form,
    FormControl,
    FormField,
    FormItem,
    FormLabel,
    FormMessage,
} from "~/components/ui/form";
import { Input } from "~/components/ui/input";
import { zodResolver } from "@hookform/resolvers/zod";
import { Button } from "~/components/ui/button";
import axiosClient from "~/lib/axiosClient";
import { useRouter } from "next/navigation";

const schema = z.object({
    email: z.string().email("Invalid email address"),
});

type ForgotSchema = z.infer<typeof schema>;

export default function ForgotPasswordPage() {
    const [message, setMessage] = useState<string | null>(null);
    const form = useForm<ForgotSchema>({
        resolver: zodResolver(schema),
        defaultValues: { email: "" },
    });

    const mutation = useMutation({
        mutationFn: async (values: ForgotSchema) => {
            const res = await axiosClient.post("/auth/forgot-password", values);
            return res.data;
        },
        onSuccess: () => {
            setMessage("If that email exists, reset instructions were sent.");
        },
        onError: () => {
            setMessage("Something went wrong. Try again later.");
        },
    });

    const onSubmit = (values: ForgotSchema) => {
        setMessage(null);
        mutation.mutate(values);
    };

    return (
        <div className="flex min-h-screen items-center justify-center bg-muted/30 px-4">
            <Card className="w-full max-w-md">
                <CardHeader className="text-center">
                    <CardTitle className="text-xl font-semibold">Forgot Password</CardTitle>
                    <CardDescription>
                        Enter your email and we’ll send reset instructions.
                    </CardDescription>
                </CardHeader>

                <CardContent>
                    <Form {...form}>
                        <form className="space-y-4" onSubmit={form.handleSubmit(onSubmit)}>
                            <FormField
                                control={form.control}
                                name="email"
                                render={({ field }) => (
                                    <FormItem>
                                        <FormLabel>Email Address</FormLabel>
                                        <FormControl>
                                            <Input
                                                type="email"
                                                placeholder="name@example.com"
                                                disabled={mutation.isSuccess}
                                                {...field}
                                            />
                                        </FormControl>
                                        <FormMessage />
                                    </FormItem>
                                )}
                            />

                            {mutation.isSuccess && (
                                <div className="flex items-center gap-2 text-sm text-green-600">
                                    <CheckCircle2 className="h-4 w-4" />
                                    {message}
                                </div>
                            )}

                            {mutation.isError && (
                                <div className="flex items-center gap-2 text-sm text-red-500">
                                    <AlertCircle className="h-4 w-4" />
                                    {message}
                                </div>
                            )}

                            <Button
                                className="w-full"
                                type="submit"
                                disabled={mutation.isPending || mutation.isSuccess}
                            >
                                {mutation.isPending && (
                                    <Loader2 className="mr-2 h-4 w-4 animate-spin" />
                                )}
                                {mutation.isSuccess ? "Email Sent" : "Send Reset Link"}
                            </Button>
                        </form>
                    </Form>
                </CardContent>
            </Card>
        </div>
    );
}
