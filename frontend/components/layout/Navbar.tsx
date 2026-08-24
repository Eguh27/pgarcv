"use client";
import Link from "next/link";
import { Search, X } from "lucide-react";
import { useState, useRef, useEffect } from "react";
import { useRouter } from "next/navigation";
import { ThemeToggle } from "@/components/ui/ThemeToggle";

export function Navbar() {
  const [search, setSearch] = useState("");
  const [searchOpen, setSearchOpen] = useState(false);
  const router = useRouter();
  const inputRef = useRef<HTMLInputElement>(null);

  const handleSearch = (e: React.FormEvent) => {
    e.preventDefault();
    if (search.trim()) {
      router.push(`/search?q=${encodeURIComponent(search.trim())}`);
      setSearchOpen(false);
    }
  };

  useEffect(() => {
    if (searchOpen && inputRef.current) {
      inputRef.current.focus();
    }
  }, [searchOpen]);

  return (
    <>
      <style>{`
        .navbar {
          position: sticky;
          top: 0;
          z-index: 100;
          background: var(--bg-primary);
          border-bottom: 1px solid var(--border);
          backdrop-filter: blur(8px);
        }
        .navbar-inner {
          max-width: 1400px;
          margin: 0 auto;
          padding: 0 16px;
          height: 56px;
          display: flex;
          align-items: center;
          gap: 12px;
        }
        .navbar-logo {
          font-family: var(--font-poppins);
          font-weight: 700;
          font-size: 20px;
          color: var(--accent);
          text-decoration: none;
          flex-shrink: 0;
        }
        .navbar-search-desktop {
          flex: 1;
          max-width: 480px;
          position: relative;
          display: flex;
        }
        .navbar-search-input {
          width: 100%;
          padding: 8px 16px 8px 40px;
          border-radius: 24px;
          border: 1px solid var(--border);
          background: var(--bg-secondary);
          color: var(--text-primary);
          font-size: 14px;
          outline: none;
          font-family: var(--font-montserrat);
          transition: border-color 0.2s;
        }
        .navbar-search-input:focus {
          border-color: var(--accent);
        }
        .navbar-search-icon {
          position: absolute;
          left: 14px;
          top: 50%;
          transform: translateY(-50%);
          color: var(--text-muted);
          pointer-events: none;
        }
        .navbar-actions {
          margin-left: auto;
          display: flex;
          align-items: center;
          gap: 8px;
        }
        .navbar-search-btn-mobile {
          display: none;
          background: none;
          border: none;
          cursor: pointer;
          color: var(--text-secondary);
          padding: 8px;
          border-radius: 8px;
        }
        /* Mobile search overlay */
        .mobile-search-overlay {
          display: none;
          position: fixed;
          inset: 0;
          z-index: 200;
          background: var(--bg-primary);
          padding: 12px 16px;
          align-items: center;
          gap: 12px;
          pointer-events: none;
        }
        .mobile-search-overlay.open {
          display: flex;
          pointer-events: auto;
        }
        .mobile-search-overlay input {
          flex: 1;
          padding: 10px 16px;
          border-radius: 24px;
          border: 1px solid var(--accent);
          background: var(--bg-secondary);
          color: var(--text-primary);
          font-size: 15px;
          outline: none;
          font-family: var(--font-montserrat);
        }
        @media (max-width: 640px) {
          .navbar-search-desktop { display: none; }
          .navbar-search-btn-mobile { display: flex !important; align-items: center; justify-content: center; }
          .navbar-logo { font-size: 18px; }
          .navbar-inner { height: 52px; }
        }
      `}</style>

      {/* Mobile search overlay */}
      <div className={`mobile-search-overlay${searchOpen ? " open" : ""}`}>
        <form onSubmit={handleSearch} style={{ display: "flex", flex: 1, gap: "8px", alignItems: "center" }}>
          <input
            ref={inputRef}
            value={search}
            onChange={(e) => setSearch(e.target.value)}
            placeholder="Cari video..."
            autoComplete="off"
          />
          <button type="submit"
            style={{ padding: "10px 18px", borderRadius: "20px", background: "var(--accent)", color: "#fff", border: "none", cursor: "pointer", fontFamily: "var(--font-poppins)", fontWeight: 600, fontSize: "14px", whiteSpace: "nowrap" }}>
            Cari
          </button>
        </form>
        <button onClick={() => setSearchOpen(false)}
          style={{ background: "none", border: "none", cursor: "pointer", color: "var(--text-secondary)", padding: "8px", flexShrink: 0 }}>
          <X size={22} />
        </button>
      </div>

      <nav className="navbar">
        <div className="navbar-inner">
          <Link href="/" className="navbar-logo">PGarcv</Link>

          {/* Desktop search */}
          <form onSubmit={handleSearch} className="navbar-search-desktop">
            <div style={{ position: "relative", width: "100%" }}>
              <input
                value={search}
                onChange={(e) => setSearch(e.target.value)}
                placeholder="Cari video..."
                className="navbar-search-input"
                autoComplete="off"
              />
              <Search size={16} className="navbar-search-icon" />
            </div>
          </form>

          <div className="navbar-actions">
            {/* Mobile search button */}
            <button
              className="navbar-search-btn-mobile"
              onClick={() => setSearchOpen(true)}
              aria-label="Cari"
            >
              <Search size={20} />
            </button>
            <ThemeToggle />
          </div>
        </div>
      </nav>
    </>
  );
}
