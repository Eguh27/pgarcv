"use client";
import { useEffect, useMemo } from "react";
import { useRouter, usePathname } from "next/navigation";

export function useAdminAuth() {
  const router = useRouter();
  const pathname = usePathname();
  const token = useMemo(() => {
    try {
      return localStorage.getItem("admin_token") || "";
    } catch {
      return "";
    }
  }, []);
  const isAuth = Boolean(token);

  useEffect(() => {
    // useEffect hanya jalan di client, localStorage pasti tersedia
    const isLoginPage = pathname === "/admin/login";

    if (!isAuth && !isLoginPage) {
      router.replace("/admin/login");
    }
  }, [isAuth, pathname, router]);

  return { checked: true, isAuth };
}

