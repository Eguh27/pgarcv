/** @type {import('next').NextConfig} */

// Host LAN (opsional). Cukup diisi SEKALI di frontend/.env.local, misal
// LAN_HOST=192.168.100.78. Dipakai untuk:
//   1. remotePatterns next/image
//   2. CORS_ORIGIN di backend/.env (backend/: CORS_ORIGIN=...,${LAN_HOST}:3000)
//
// IP LAN sering berubah saat ganti hotspot/router. Karena itu JANGAN
// menuliskan IP langsung di dalam file ini — cukup pindahkan nilainya ke
// .env.local kalau addressnya berubah.
//
// Kosongkan (atau hapus) variabel ini kalau hanya diakses via localhost.
const LAN_HOST = process.env.LAN_HOST || "";

const nextConfig = {
  // Next.js 16 memblokir resource dev (_next/*, HMR) yang diakses dari origin
  // selain localhost. Tanpa ini, membuka situs via IP LAN membuat HMR dan
  // sebagian aset ditolak → gambar tidak muncul di HP despite URL-nya valid.
  //
  // Dipakai juga oleh HLS dev/playground Next.js. Hanya relevan di mode
  // development; production tidak menerapkan pembatasan ini.
  ...(LAN_HOST ? { allowedDevOrigins: [LAN_HOST] } : {}),

  images: {
    // Next.js 16 memblokir optimasi gambar yang di-host di IP privat/localhost
    // demi keamanan (SSRF). Tanpa ini, next/image menolak URL backend lokal
    // dengan 400 '"url" parameter is not allowed' — remotePatterns sudah benar,
    // tapi pengecekan IP privat berjalan sesudahnya.
    //
    // Aman selama backend hanya jalan di jaringan privat. Hapus flag ini
    // begitu backend dipindahkan ke domain publik.
    dangerouslyAllowLocalIP: true,

    remotePatterns: [
      {
        protocol: "http",
        hostname: "localhost",
        port: "8080",
        pathname: "/uploads/**",
      },
      // Host LAN hanya ditambahkan bila LAN_HOST diisi.
      ...(LAN_HOST
        ? [
            {
              protocol: "http",
              hostname: LAN_HOST,
              port: "8080",
              pathname: "/uploads/**",
            },
          ]
        : []),
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