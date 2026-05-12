import { NextResponse } from "next/server";
import type { NextRequest } from "next/server";

export function middleware(request: NextRequest) {
  // No redirects — auth handled client-side via localStorage
  return NextResponse.next();
}

export const config = {
  matcher: [],
};