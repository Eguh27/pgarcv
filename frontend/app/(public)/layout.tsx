import { Navbar } from "@/components/layout/Navbar";

export default function PublicLayout({ children }: { children: React.ReactNode }) {
  return (
    <>
      <Navbar />
      <main style={{
        maxWidth: "1400px",
        margin: "0 auto",
        padding: "16px 12px 32px",
      }}>
        {children}
      </main>
    </>
  );
}
