"use client";
import { useState } from "react";
import { adminApi } from "@/lib/api";

export default function AdminLoginPage() {
  const [username, setUsername] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLogin = async () => {
    setLoading(true);
    setError("");
    try {
      // Pakai adminApi, bukan fetch manual: BASE di lib/api.ts sudah
      // membedakan browser (path relatif) dan server (URL absolut).
      // Versi lama memakai BASE sendiri dengan fallback http://localhost:8080
      // sehingga di HP mengarah ke loopback HP itu sendiri → "Failed to fetch".
      const res = await adminApi.auth.login(username, password);
      // adminApi.auth.login sudah menyimpan token ke localStorage.
      if (!res?.token) throw new Error("Token tidak ada di response");

      // Hard navigate — bukan router.push, supaya cookie baru ikut
      // dan proxy.ts membaca sesi saat render.
      window.location.href = "/admin/dashboard";

    } catch (err: unknown) {
      setError(err instanceof Error ? err.message : "Login gagal");
    } finally {
      setLoading(false);
    }
  };

  return (
    <div style={{ minHeight: "100vh", display: "flex", alignItems: "center", justifyContent: "center", background: "var(--bg-secondary)" }}>
      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "16px", padding: "40px", width: "100%", maxWidth: "400px", boxShadow: "var(--shadow-md)" }}>
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "24px", marginBottom: "8px", textAlign: "center" }}>Admin Panel</h1>
        <p style={{ textAlign: "center", color: "var(--text-secondary)", marginBottom: "32px", fontSize: "14px" }}>Masuk untuk mengelola konten</p>

        {error && (
          <div style={{ background: "#fee2e2", border: "1px solid #fca5a5", borderRadius: "8px", padding: "12px", marginBottom: "16px", color: "#dc2626", fontSize: "14px" }}>
            {error}
          </div>
        )}

        <div style={{ display: "flex", flexDirection: "column", gap: "16px" }}>
          <div>
            <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Username</label>
            <input
              value={username}
              onChange={(e) => { setUsername(e.target.value); setError(""); }}
              onKeyDown={(e) => e.key === "Enter" && handleLogin()}
              autoComplete="username"
              style={{ width: "100%", padding: "10px 14px", borderRadius: "8px", border: "1px solid var(--border)", background: "var(--bg-secondary)", color: "var(--text-primary)", fontSize: "14px", outline: "none", fontFamily: "var(--font-montserrat)" }}
            />
          </div>
          <div>
            <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Password</label>
            <input
              type="password"
              value={password}
              onChange={(e) => { setPassword(e.target.value); setError(""); }}
              onKeyDown={(e) => e.key === "Enter" && handleLogin()}
              autoComplete="current-password"
              style={{ width: "100%", padding: "10px 14px", borderRadius: "8px", border: "1px solid var(--border)", background: "var(--bg-secondary)", color: "var(--text-primary)", fontSize: "14px", outline: "none", fontFamily: "var(--font-montserrat)" }}
            />
          </div>
          <button
            type="button"
            onClick={handleLogin}
            disabled={loading}
            style={{ padding: "12px", borderRadius: "8px", background: "var(--accent)", color: "#fff", border: "none", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "15px", cursor: loading ? "not-allowed" : "pointer", opacity: loading ? 0.7 : 1 }}
          >
            {loading ? "Masuk..." : "Masuk"}
          </button>
        </div>
      </div>
    </div>
  );
}