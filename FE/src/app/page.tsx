import { redirect } from "next/navigation";

export default function HomePage() {
  console.log("Home page");
  redirect("/dashboard");
}
