import { NextResponse, type NextRequest } from "next/server";

// Guard level-UX untuk rute admin: redirect ke login bila cookie admin_token
// tidak ada. Verifikasi JWT tetap di backend (AuthRequired middleware).
export function proxy(request: NextRequest) {
  const { pathname } = request.nextUrl;

  if (pathname === "/admin/login") {
    return NextResponse.next();
  }

  const token = request.cookies.get("admin_token")?.value;
  if (!token) {
    const loginUrl = new URL("/admin/login", request.url);
    return NextResponse.redirect(loginUrl);
  }

  return NextResponse.next();
}

export const config = {
  matcher: "/admin/:path*",
};
