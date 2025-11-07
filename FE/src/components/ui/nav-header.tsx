"use client";

import Link from "next/link";
import { Badge } from "./badge";
import { Button } from "./button";
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger } from "./dropdown-menu";
import { Avatar, AvatarFallback } from "./avatar";
import { ModeToggle } from "../mode-toggle";
import { useAuth } from "~/hooks/useAuth";
import { Loader2 } from "lucide-react";
import { logout } from "~/services/authService";

const NavHeader = () => {
    const { user, loading } = useAuth();

    if (loading)
        return (
            <div className="flex justify-center items-center h-14">
                <Loader2 className="animate-spin h-5 w-5 text-muted-foreground" />
            </div>
        );

    if (!user) return null;
    return <header className="bg-background sticky top-0 z-10 flex justify-center border-b">
        <div className="container flex items-center justify-between px-4 py-2">
            <Link href="/dashboard" className="flex items-center">
                <div className="font-sans text-xl font-medium tracking-tight">
                    <span className="text-foreground">Podcast</span>
                    <span className="font-light text-grey-500">/</span>
                    <span className="text-foreground font-light">Clipper</span>

                </div>
            </Link>
            <div className="flex items.center gap-4">
                <div className="flex items-center gap-2">
                    <Badge variant={"secondary"} className="h-8 px-3 py-1.5 text-xs font-medium">
                        {user.credits} credits
                    </Badge>
                    <Button
                        variant="outline"
                        size="sm"
                        asChild
                        className="h-8 text-3 font-medium">
                        <Link href="/dashboard/billing">
                            Buy more
                        </Link>
                    </Button>
                </div>
                <div>
                    <ModeToggle />
                </div>
                <DropdownMenu>
                    <DropdownMenuTrigger asChild>
                        <Button
                            variant="ghost"
                            className="relative h-8 w-8 rounded-full p-0"
                        >
                            <Avatar>
                                <AvatarFallback>{user.username.charAt(0).toUpperCase()}</AvatarFallback>
                            </Avatar>
                        </Button>
                    </DropdownMenuTrigger>
                    <DropdownMenuContent align="end">
                        <DropdownMenuLabel>
                            <p className="text-muted-foreground text-xs">{user.username}</p>
                        </DropdownMenuLabel>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem asChild>
                            <Link href="/dashboard/billing">Billing</Link>
                        </DropdownMenuItem>
                        <DropdownMenuSeparator />
                        <DropdownMenuItem
                            onClick={() => logout()}
                            className="text-destructive cursor-pointer"
                        >
                            Sign out
                        </DropdownMenuItem>
                    </DropdownMenuContent>
                </DropdownMenu>

            </div>
        </div>
    </header>
}
export default NavHeader