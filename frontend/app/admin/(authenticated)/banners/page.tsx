"use client";

import { useEffect, useState } from "react";
import { Pencil, Plus, ToggleLeft, ToggleRight, Trash2 } from "lucide-react";
import { adminApi, mediaUrl } from "@/lib/api";
import type { Banner } from "@/lib/api";

type BannerForm = {
  title: string;
  image_url: string;
  link_url: string;
  is_active: boolean;
  sort_order: number;
};

const emptyForm: BannerForm = { title: "", image_url: "", link_url: "", is_active: true, sort_order: 0 };

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

export default function AdminBannersPage() {
  const [banners, setBanners] = useState<Banner[]>([]);
  const [loading, setLoading] = useState(true);
  const [showModal, setShowModal] = useState(false);
  const [editing, setEditing] = useState<Banner | null>(null);
  const [form, setForm] = useState<BannerForm>(emptyForm);
  const [saving, setSaving] = useState(false);
  const [uploading, setUploading] = useState(false);

  useEffect(() => {
    let active = true;
    adminApi.banners
      .list()
      .then((res) => {
        if (active) setBanners(res);
      })
      .catch(console.error)
      .finally(() => {
        if (active) setLoading(false);
      });
    return () => {
      active = false;
    };
  }, []);

  const refreshBanners = async () => {
    setBanners(await adminApi.banners.list());
  };

  const openAdd = () => {
    setEditing(null);
    setForm(emptyForm);
    setShowModal(true);
  };

  const openEdit = (banner: Banner) => {
    setEditing(banner);
    setForm({
      title: banner.title,
      image_url: banner.image_url,
      link_url: banner.link_url,
      is_active: banner.is_active,
      sort_order: banner.sort_order,
    });
    setShowModal(true);
  };

  const handleUpload = async (event: React.ChangeEvent<HTMLInputElement>) => {
    const file = event.target.files?.[0];
    if (!file) return;
    setUploading(true);
    try {
      const res = await adminApi.upload(file);
      setForm((prev) => ({ ...prev, image_url: res.url }));
    } catch {
      alert("Upload gagal");
    } finally {
      setUploading(false);
    }
  };

  const handleSave = async () => {
    if (!form.title.trim()) {
      alert("Judul wajib diisi");
      return;
    }
    setSaving(true);
    try {
      if (editing) await adminApi.banners.update(editing.id, form);
      else await adminApi.banners.create(form);
      setShowModal(false);
      await refreshBanners();
    } catch {
      alert("Gagal simpan");
    } finally {
      setSaving(false);
    }
  };

  const handleDelete = async (id: number) => {
    if (!confirm("Hapus banner ini?")) return;
    await adminApi.banners.delete(id);
    await refreshBanners();
  };

  const toggleActive = async (banner: Banner) => {
    await adminApi.banners.update(banner.id, { is_active: !banner.is_active });
    setBanners((prev) => prev.map((item) => (item.id === banner.id ? { ...item, is_active: !item.is_active } : item)));
  };

  return (
    <div>
      <div style={{ display: "flex", justifyContent: "space-between", alignItems: "center", marginBottom: "20px" }}>
        <h1 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "22px" }}>Banner</h1>
        <button onClick={openAdd} style={{ display: "flex", alignItems: "center", gap: "8px", padding: "10px 18px", borderRadius: "10px", background: "var(--accent)", color: "#fff", border: "none", cursor: "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px" }}>
          <Plus size={16} /> Tambah Banner
        </button>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "repeat(auto-fill, minmax(280px, 1fr))", gap: "16px" }}>
        {loading ? (
          <div style={{ gridColumn: "1/-1", padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Memuat...</div>
        ) : (
          banners.map((banner) => (
            <div key={banner.id} style={{ borderRadius: "14px", background: "var(--bg-card)", border: "1px solid var(--border)", overflow: "hidden" }}>
              <div style={{ aspectRatio: "16/5", background: "var(--surface)", position: "relative", overflow: "hidden" }}>
                {banner.image_url && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={mediaUrl(banner.image_url)} alt="" style={{ width: "100%", height: "100%", objectFit: "cover" }} />
                )}
                {!banner.is_active && (
                  <div style={{ position: "absolute", inset: 0, background: "rgba(0,0,0,0.4)", display: "flex", alignItems: "center", justifyContent: "center" }}>
                    <span style={{ color: "#fff", fontSize: "12px", background: "rgba(0,0,0,0.6)", padding: "4px 10px", borderRadius: "20px" }}>Nonaktif</span>
                  </div>
                )}
              </div>
              <div style={{ padding: "14px" }}>
                <p style={{ fontSize: "14px", fontWeight: 600, color: "var(--text-primary)", marginBottom: "4px" }}>{banner.title}</p>
                <p style={{ fontSize: "12px", color: "var(--text-muted)", marginBottom: "12px" }}>Urutan: {banner.sort_order}</p>
                <div style={{ display: "flex", gap: "8px", flexWrap: "wrap" }}>
                  <button onClick={() => openEdit(banner)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: "var(--text-primary)", border: "none", cursor: "pointer", fontSize: "13px" }}>
                    <Pencil size={13} /> Edit
                  </button>
                  <button onClick={() => toggleActive(banner)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "var(--surface)", color: banner.is_active ? "#10b981" : "var(--text-muted)", border: "none", cursor: "pointer", fontSize: "13px" }}>
                    {banner.is_active ? <ToggleRight size={15} /> : <ToggleLeft size={15} />}
                    {banner.is_active ? "Aktif" : "Nonaktif"}
                  </button>
                  <button onClick={() => handleDelete(banner.id)} style={{ display: "flex", alignItems: "center", gap: "4px", padding: "6px 12px", borderRadius: "8px", background: "#fee2e2", color: "#dc2626", border: "none", cursor: "pointer", fontSize: "13px" }}>
                    <Trash2 size={13} />
                  </button>
                </div>
              </div>
            </div>
          ))
        )}
        {!loading && banners.length === 0 && (
          <div style={{ gridColumn: "1/-1", padding: "48px", textAlign: "center", color: "var(--text-muted)" }}>Belum ada banner.</div>
        )}
      </div>

      {showModal && (
        <div style={{ position: "fixed", inset: 0, zIndex: 200, display: "flex", alignItems: "center", justifyContent: "center", padding: "16px" }} onClick={(event) => { if (event.target === event.currentTarget) setShowModal(false); }}>
          <div style={{ position: "absolute", inset: 0, background: "var(--overlay)" }} />
          <div style={{ position: "relative", background: "var(--bg-card)", borderRadius: "16px", padding: "28px", width: "100%", maxWidth: "480px", border: "1px solid var(--border)", zIndex: 1 }}>
            <h2 style={{ fontFamily: "var(--font-poppins)", fontWeight: 700, fontSize: "18px", marginBottom: "20px" }}>
              {editing ? "Edit Banner" : "Tambah Banner"}
            </h2>
            <div style={{ display: "flex", flexDirection: "column", gap: "14px" }}>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Judul *</label>
                <input value={form.title} onChange={(event) => setForm((prev) => ({ ...prev, title: event.target.value }))} style={inputStyle} placeholder="Judul banner" />
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Gambar Banner</label>
                <div style={{ display: "flex", gap: "8px" }}>
                  <input value={form.image_url} onChange={(event) => setForm((prev) => ({ ...prev, image_url: event.target.value }))} style={{ ...inputStyle, flex: 1 }} placeholder="/uploads/banner.jpg" />
                  <label style={{ padding: "9px 14px", background: "var(--surface)", borderRadius: "8px", cursor: "pointer", fontSize: "13px", whiteSpace: "nowrap", border: "1px solid var(--border)" }}>
                    {uploading ? "..." : "Upload"}
                    <input type="file" accept="image/*" style={{ display: "none" }} onChange={handleUpload} />
                  </label>
                </div>
                {form.image_url && (
                  // eslint-disable-next-line @next/next/no-img-element
                  <img src={mediaUrl(form.image_url)} alt="" style={{ marginTop: "8px", maxHeight: "80px", borderRadius: "6px" }} />
                )}
              </div>
              <div>
                <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Link URL</label>
                <input value={form.link_url} onChange={(event) => setForm((prev) => ({ ...prev, link_url: event.target.value }))} style={inputStyle} placeholder="https://..." />
              </div>
              <div style={{ display: "flex", gap: "12px" }}>
                <div style={{ flex: 1 }}>
                  <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Urutan</label>
                  <input type="number" value={form.sort_order} onChange={(event) => setForm((prev) => ({ ...prev, sort_order: Number(event.target.value) }))} style={inputStyle} min={0} />
                </div>
                <div>
                  <label style={{ display: "block", fontSize: "13px", fontWeight: 500, marginBottom: "6px", color: "var(--text-secondary)" }}>Status</label>
                  <button type="button" onClick={() => setForm((prev) => ({ ...prev, is_active: !prev.is_active }))} style={{ padding: "9px 16px", borderRadius: "8px", border: "1px solid var(--border)", background: form.is_active ? "#d1fae5" : "var(--surface)", color: form.is_active ? "#065f46" : "var(--text-secondary)", cursor: "pointer", fontSize: "13px", fontWeight: 500 }}>
                    {form.is_active ? "Aktif" : "Nonaktif"}
                  </button>
                </div>
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
