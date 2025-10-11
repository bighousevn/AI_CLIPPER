import { redirect } from "next/navigation";

export default function HomePage() {
  console.log("Redirecting to /dashboard");
  redirect("/dashboard");
}
