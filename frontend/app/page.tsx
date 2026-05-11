// ✅ Root page.tsx hanya redirect ke (public) group
// Ini menghindari duplicate rendering dan memastikan layout.tsx digunakan
import { redirect } from "next/navigation";

export default function RootPage() {
  redirect("/?view=root");
  return null;
}


 