export default function StatusBanner({ tone = "default", children }) {
  const toneClass =
    tone === "error"
      ? "border-accent text-accent"
      : tone === "success"
        ? "border-paper text-paper"
        : "border-line text-muted";

  return (
    <div className={`border px-4 py-3 text-xs font-semibold uppercase tracking-[0.18em] ${toneClass}`}>
      {children}
    </div>
  );
}
