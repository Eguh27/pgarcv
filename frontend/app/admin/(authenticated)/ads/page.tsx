"use client";

import { useEffect, useState } from "react";
import { Pencil, Plus, Trash2 } from "lucide-react";
import { adminApi } from "@/lib/api";
import type { Ad } from "@/lib/api";

type AdForm = { name: string; ad_code: string; device_type: string; is_active: boolean };

const emptyForm: AdForm = { name: "", ad_code: "", device_type: "all", is_active: true };

const inputStyle: React.CSSProperties = {
  width: "100%",
  padding: "9px 12px",
  borderRadius: "8px",
  border: "1px solid var(--border)",
  background: "var(--bg-secondary)",
  color: "var(--text-primary)",
  fontSize: "14px",
  outline: "none",
  fontFamily: "var(--font-montserrat)",
};

export default function AdminAdsPage() {
  const [ads, setAds] = useState<Ad[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Ad | null>(null);
  const [form, setForm] = useState<AdForm>(emptyForm);
  const [saving, setSaving] = useState(false);

  useEffect(() => {
    let active = true;
    adminApi.ads
      .list()
      .then((res) => {
        if (active) setAds(res);
      })
      .catch(console.error)
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const refreshAds = async () => {
    setAds(await adminApi.ads.list());
  };

  const openAdd = () => {
    setEditing(null);
    setForm(emptyForm);
    setShowModal(true);
  };

  const openEdit = (ad: Ad) => {
    setEditing(ad);
    setForm({ name: ad.name, ad_code: ad.ad_code, device_type: ad.device_type, is_active: ad.is_active });
    setShowModal(true);
  };

  const handleSave = async () => {
    if (!form.name.trim()) {
      alert("Nama wajib diisi");
      return;
    }
    setSaving(true);
    try {
      if (editing) await adminApi.ads.update(editing.id, form);
      else await adminApi.ads.create(form);
      setShowModal(false);
      await refreshAds();
    } catch {
      alert("Gagal simpan");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Hapus iklan ini?")) return;
    await adminApi.ads.delete(id);
    await refreshAds();
  };

  const deviceLabel: Record<string, string> = { all: "Semua", mobile: "Mobile", desktop: "Desktop" };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <div>
          <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px" }}>Manajemen Iklan</h1>
          <p style={{ fontSize: "13px", color: "var(--text-muted)" }}>Kelola kode iklan per target device</p>
        </div>
        <button onClick={openAdd} style={{ display: "flex", alignItems: "center", gap: "8px", padding: "10px 18px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px" }}>
          <Plus size={16} /> Tambah Iklan
        </button>
      </div>

      <div style={{ background: "var(--bg-card)", border: "1px solid var(--border)", borderRadius: "14px", overflow: "hidden" }}>
        <table style={{ width: "100%", borderCollapse: "collapse" }}>
          <thead>
            <tr style={{ background: "var(--bg-secondary)" }}>
              {["Nama", "Device", "Status", "Aksi"].map((heading) => (
                <th key={heading} style={{ padding: "12px 16px", textAlign: "left", fontSize: "12px", fontWeight: 600, color: "var(--text-muted)", textTransform: "uppercase" }}>
                  {heading}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {loading ? (
              <tr>
                <td colSpan={4} style={{ padding: "32px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</td>
              </tr>
            ) : (
              ads.map((ad) => (
                <tr key={ad.id} style={{ borderTop: "1px solid var(--border)" }}>
                  <td style={{ padding: "14px 16px", fontWeight: 500, fontSize: "14px" }}>{ad.name}</td>
                  <td style={{ padding: "14px 16px" }}>
                    <span style={{ padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600, background: "var(--surface)", color: "var(--text-secondary)" }}>
                      {deviceLabel[ad.device_type] ?? ad.device_type}
                    </span>
                  </td>
                  <td style={{ padding: "14px 16px" }}>
                    <span style={{ padding: "3px 10px", borderRadius: "20px", fontSize: "12px", fontWeight: 600, background: ad.is_active ? "#d1fae5" : "#fef3c7", color: ad.is_active ? "#065f46" : "#92400e" }}>
                      {ad.is_active ? "Aktif" : "Nonaktif"}
                    </span>
                  </td>
                  <td style={{ padding: "14px 16px" }}>
                    <div style={{ display: "flex", gap: "8px" }}>
                      <button onClick={() => openEdit(ad)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", border: "none", cursor: "pointer", fontSize: "13px", color: "var(--text-primary)" }}>
                        <Pencil size={13} /> Edit
                      </button>
                      <button onClick={() => handleDelete(ad.id)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                        <Trash2 size={13} />
                      </button>
                    </div>
                  </td>
                </tr>
              ))
            )}
            {!loading && ads.length === 0 && (
              <tr>
                <td colSpan={4} style={{ padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Belum ada iklan.</td>
              </tr>
            )}
          </tbody>
        </table>
      </div>

      {showModal && (
        <div style={{ position: "fixed", inset: 0, zIndex: 200, display: "flex", alignItems: "center", justifyContent: "center", padding: "16px" }} onClick={(event) => { if (event.target === event.currentTarget) setShowModal(false); }}>
          <div style={{ position: "absolute", inset: 0, background: "var(--overlay)" }} />
          <div style={{ position: "relative", background: "var(--bg-card)", borderRadius: "16px", padding: "28px", width: "100%", maxWidth: "520px", border: "1px solid var(--border)", zIndex: 1 }}>
            <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "18px", marginBottom: "20px" }}>
              {editing ? "Edit Iklan" : "Tambah Iklan"}
            </h2>
            <div style={{ display: "flex", flexDirection: "column", gap: "14px" }}>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Nama Iklan *</label>
                <input value={form.name} onChange={(event) => setForm((prev) => ({ ...prev, name: event.target.value }))} style={inputStyle} placeholder="Mis: Google AdSense Utama" />
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Target Device</label>
                <select value={form.device_type} onChange={(event) => setForm((prev) => ({ ...prev, device_type: event.target.value }))} style={inputStyle}>
                  <option value="all">Semua Device</option>
                  <option value="mobile">Mobile Only</option>
                  <option value="desktop">Desktop Only</option>
                </select>
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Ad Code HTML</label>
                <textarea value={form.ad_code} onChange={(event) => setForm((prev) => ({ ...prev, ad_code: event.target.value }))} rows={5} placeholder='<ins class="adsbygoogle"></ins>' style={{ ...inputStyle, resize: "vertical", fontFamily: "monospace", fontSize: "12px" }} />
              </div>
              <div style={{ display: "flex", alignItems: "center", gap: "10px" }}>
                <button type="button" onClick={() => setForm((prev) => ({ ...prev, is_active: !prev.is_active }))} style={{ width: "44px", height: "24px", borderRadius: "12px", background: form.is_active ? "var(--accent)" : "var(--surface)", border: "none", cursor: "pointer", position: "relative", transition: "background 0.2s", flexShrink: 0 }}>
                  <span style={{ position: "absolute", top: "2px", left: form.is_active ? "22px" : "2px", width: "20px", height: "20px", borderRadius: "50%", background: "#fff", transition: "left 0.2s" }} />
                </button>
                <span style={{ fontSize: "13px", color: "var(--text-secondary)" }}>{form.is_active ? "Aktif - iklan akan ditayangkan" : "Nonaktif"}</span>
              </div>
            </div>
            <div style={{ display: "flex", gap: "10px", marginTop: "24px" }}>
              <button onClick={handleSave} disabled={saving} style={{ flex: 1, padding: "11px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: saving ? "not-allowed" : "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px", opacity: saving ? 0.7 : 1 }}>
                {saving ? "Menyimpan..." : "Simpan"}
              </button>
              <button onClick={() => setShowModal(false)} style={{ padding: "11px 20px", borderRadius: "10px", background: "var(--surface)", border: "none", cursor: "pointer", fontSize: "14px" }}>
                Batal
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
}
