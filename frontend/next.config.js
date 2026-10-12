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
    return [
      {
        source: "/api/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/api/:path*`,
      },
      {
        source: "/uploads/:path*",
        destination: `${process.env.NEXT_PUBLIC_API_URL || "http://localhost:8080"}/uploads/:path*`,
      },
    ];
  },
};

module.exports = nextConfig;


