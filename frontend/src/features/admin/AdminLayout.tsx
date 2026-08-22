import { NavLink, Outlet } from "react-router-dom";

const navItems = [
  { to: "/admin/users", label: "Users" },
  { to: "/admin/teams", label: "Teams" },
  { to: "/admin/settings", label: "Settings" },
];

export function AdminLayout() {
  return (
    <div className="min-h-screen bg-paper">
      <header className="border-b border-ink/10 bg-paper-dim">
        <div className="mx-auto flex max-w-6xl flex-col gap-4 px-6 py-6 sm:flex-row sm:items-end sm:justify-between">
          <div>
            <p className="text-xs font-medium uppercase tracking-widest text-ink500">
              Administration
            </p>
            <h1 className="font-display text-2xl font-semibold text-ink">
              Habeshan REMS
            </h1>
          </div>
          <nav className="flex flex-wrap gap-2">
            {navItems.map((item) => (
              <NavLink
                key={item.to}
                to={item.to}
                className={({ isActive }) =>
                  [
                    "rounded px-3 py-1.5 text-sm font-medium transition-colors",
                    isActive
                      ? "bg-ink text-paper"
                      : "text-ink hover:bg-ink/5",
                  ].join(" ")
                }
              >
                {item.label}
              </NavLink>
            ))}
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-6xl px-6 py-8">
        <Outlet />
      </main>
    </div>
  );
}
