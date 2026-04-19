import { useState } from "react";
import { Link } from "react-router-dom";
import { useAuth } from "../context/AuthContext";

const linksByRole = {
  CUSTOMER: [{ label: "Shop", href: "/shop" }],
  WAREHOUSE_STAFF: [{ label: "Inventory", href: "/inventory" }],
  DELIVERY_PARTNER: [{ label: "Delivery", href: "/delivery" }],
  ADMIN: [{ label: "Admin", href: "/admin" }],
  SUPPORT_AGENT: [{ label: "Support", href: "/support" }],
};

function DotBurstIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-4 w-4"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="2"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M12 3v18M3 12h18M5.6 5.6l12.8 12.8M18.4 5.6 5.6 18.4" />
    </svg>
  );
}

function MenuIcon() {
  return (
    <svg
      aria-hidden="true"
      className="h-5 w-5"
      fill="none"
      viewBox="0 0 24 24"
      stroke="currentColor"
      strokeWidth="1.8"
    >
      <path strokeLinecap="round" strokeLinejoin="round" d="M4 7h16M4 12h16M4 17h16" />
    </svg>
  );
}

function NavLink({ item, mobile = false }) {
  const className = mobile
    ? "flex items-center justify-between text-sm font-semibold tracking-[0.18em] text-paper transition-opacity duration-200 hover:opacity-70"
    : "group inline-flex items-center gap-2 text-[11px] font-semibold tracking-[0.2em] text-paper transition-opacity duration-200 hover:opacity-65";

  return (
    <Link to={item.href} className={className}>
      <span className="relative">
        {item.label}
        {!mobile ? (
          <span className="absolute -bottom-1 left-0 h-px w-0 bg-paper transition-all duration-200 group-hover:w-full" />
        ) : null}
      </span>
    </Link>
  );
}

export default function Navbar() {
  const [open, setOpen] = useState(false);
  const { user, role, logout } = useAuth();
  const navItems = role ? linksByRole[role] ?? [] : [];
  const homeTarget = role ? navItems[0]?.href ?? "/login" : "/login";

  return (
    <header className="border-b border-line">
      <nav className="flex min-h-[70px] items-center justify-between px-5 sm:px-8 lg:px-12 xl:px-16">
        <Link
          to={homeTarget}
          className="shrink-0 border-r border-line pr-6 text-[11px] font-semibold tracking-[0.42em] text-paper sm:pr-10"
        >
          CARTGO
        </Link>

        <div className="hidden min-w-0 flex-1 items-center justify-center lg:flex">
          <ul className="flex items-center gap-8 xl:gap-12">
            {navItems.map((item) => (
              <li key={item.label}>
                <NavLink item={item} />
              </li>
            ))}
          </ul>
        </div>

        <div className="hidden items-center gap-4 border-l border-line pl-6 lg:flex">
          <div className="text-right">
            <p className="text-[10px] font-semibold uppercase tracking-[0.28em] text-muted">
              {role?.replaceAll("_", " ")}
            </p>
            <p className="max-w-[12rem] truncate text-sm font-semibold text-paper">
              {user?.email}
            </p>
          </div>
          <button
            type="button"
            onClick={logout}
            className="inline-flex items-center gap-2 bg-paper px-4 py-3 text-sm font-medium text-surface transition-colors duration-200 hover:bg-white"
          >
            Sign out
            <span className="text-accent">
              <DotBurstIcon />
            </span>
          </button>
        </div>

        <button
          type="button"
          onClick={() => setOpen((value) => !value)}
          className="inline-flex items-center justify-center rounded-full border border-line p-3 text-paper transition-colors duration-200 hover:border-paper lg:hidden"
          aria-label="Toggle menu"
          aria-expanded={open}
        >
          <MenuIcon />
        </button>
      </nav>

      {open ? (
        <div className="border-t border-line px-5 py-5 sm:px-8 lg:hidden">
          <div className="flex flex-col gap-4">
            {navItems.map((item) => (
              <NavLink key={item.label} item={item} mobile />
            ))}
            <button
              type="button"
              onClick={logout}
              className="mt-2 inline-flex w-fit items-center gap-2 bg-paper px-4 py-3 text-sm font-medium text-surface transition-colors duration-200 hover:bg-white"
            >
              Sign out
              <span className="text-accent">
                <DotBurstIcon />
              </span>
            </button>
          </div>
        </div>
      ) : null}
    </header>
  );
}
