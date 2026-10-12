/** @type {import('next').NextConfig} */
const nextConfig = {
  images: {
    // Next.js 16 memblokir optimasi gambar yang di-host di IP privat/localhost
    // demi keamanan (SSRF). Tanpa ini, next/image menolak URL backend lokal
    // dengan 400 '"url" parameter is not allowed' — remotePatterns sudah benar,
    // tapi pengecekan IP privat terjadi sesudahnya.
    //
    // Aman karena backend ini hanya berjalan di jaringan privat. Hapus flag ini
    // kalau backend sudah dipindahkan ke domain publik.
    dangerouslyAllowLocalIP: true,

    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
      {
        // IP LAN untuk akses dari perangkat lain di jaringan yang sama.
        // hostname di bawah perlu disesuaikan dengan subnet kamu.
        protocol: "http",
        hostname: "192.168.1.16",
        port: "8080",
        pathname: "/uploads/**",
      },
      {
        protocol: "https",
        hostname: "*.ngrok-free.app",
        pathname: "/uploads/**",
      },
      // ✅ Support Cloudinary URLs
      {
        protocol: "https",
        hostname: "res.cloudinary.com",
        pathname: "/**",
      },
    ],
  },
  async rewrites() {
    // PENTING: rewrites dievaluasi saat BUILD dan hasilnya di-freeze ke
    // routes-manifest.json. Env var yang disetel saat runtime TIDAK terbaca
    // di sini — karena itu API_URL harus di-declare sebagai ARG build.
    //
    // API_URL (tanpa NEXT_PUBLIC_) sengaja dipakai supaya nilainya tidak
    // ikut ter-inline ke bundle browser: value ini hanya untuk sisi server.
    const API_URL =
      process.env.API_URL ||
      process.env.NEXT_PUBLIC_API_URL ||
      "http://localhost:8080";

    return [
      {
        source: "/api/:path*",
        destination: `${API_URL}/api/:path*`,
      },
      {
        source: "/uploads/:path*",
        destination: `${API_URL}/uploads/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;


