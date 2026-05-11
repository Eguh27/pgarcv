const fs = require('fs');
const path = require('path');

const filePath = 'd:\\pgarcv\\frontend\\lib\\api.ts';
let content = fs.readFileSync(filePath, 'utf8');

const oldUpload = `  upload: async (file: File): Promise<{ url: string; filename: string; thumbnail_url?: string }> => {
    const token = getToken();

    const form = new FormData();
    form.append("file", file);


    // PENTING: jangan set Content-Type manual untuk FormData
    // browser yang auto-set boundary-nya
    const headers: Record<string, string> = {};
    if (BASE.includes("ngrok")) {
      headers["ngrok-skip-browser-warning"] = "true";
    }
    if (token) {
      headers["Authorization"] = \`Bearer \${token}\`;
    }

    const res = await fetch(\`\${BASE}/api/admin/upload\`, {
      method: "POST",
      credentials: "include",
      headers,
      body: form,
    });

    if (!res.ok) {
      const text = await res.text();
      throw new Error(\`Upload gagal (\${res.status}): \${text}\`);
    }

    const json = await res.json();
    if (!json.success) throw new Error(json.message || "Upload gagal");
    return json.data;
  },`;

const newUpload = `  upload: async (file: File): Promise<{ url: string; filename: string; thumbnail_url?: string }> => {
    const MAX_SIZE = 500 * 1024 * 1024; // 500MB
    const ALLOWED_IMAGE = [".jpg", ".jpeg", ".png", ".gif"];
    const ALLOWED_VIDEO = [".mp4", ".webm", ".mov", ".avi"];

    // Validasi client-side
    const ext = "." + file.name.split(".").pop()?.toLowerCase();
    const allowed = [...ALLOWED_IMAGE, ...ALLOWED_VIDEO];
    if (!allowed.includes(ext)) {
      throw new Error(\`Format tidak didukung: \${ext}\`);
    }
    if (file.size > MAX_SIZE) {
      throw new Error("File terlalu besar (maks 500MB)");
    }
    if (file.size === 0) {
      throw new Error("File kosong");
    }

    const form = new FormData();
    form.append("file", file);

    const token = getToken();
    const headers: Record<string, string> = { ...NGROK_HEADER };
    if (token) headers["Authorization"] = \`Bearer \${token}\`;

    let res: Response;
    try {
      res = await fetch(\`\${BASE}/api/admin/upload\`, {
        method: "POST",
        credentials: "include",
        headers,
        body: form,
      });
    } catch (e) {
      throw new Error(\`Upload network error: \${e instanceof Error ? e.message : "gagal"}\`);
    }

    if (!res.ok) {
      let msg = \`Upload gagal (\${res.status})\`;
      try {
        const err = await res.json();
        msg = err.message || msg;
      } catch {}
      throw new Error(msg);
    }

    const json = await res.json();
    if (!json.success) throw new Error(json.message || "Upload gagal");
    return json.data;
  },`;

if (content.includes(oldUpload)) {
  content = content.replace(oldUpload, newUpload);
  fs.writeFileSync(filePath, content, 'utf8');
  console.log('File updated successfully');
} else {
  console.log('Old upload function not found');
}
