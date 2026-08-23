import { NavLink, Outlet } from "react-router-dom";
import { Users, Shield, Sliders } from "lucide-react";

const navItems = [
  { to: "/admin/users", label: "User Directory", icon: Users },
  { to: "/admin/teams", label: "Team Management", icon: Shield },
  { to: "/admin/settings", label: "Organization Settings", icon: Sliders },
];

export function AdminLayout() {
  return (
    <div className="mx-auto max-w-7xl px-4 py-8 sm:px-6 lg:px-8">
      {/* Admin Subheader & Tabs */}
      <div className="mb-8 border-b border-ink/10 pb-5">
        <div className="flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
          <div>
            <h1 className="font-display text-2xl font-bold tracking-tight text-ink">
              Agency Administration
            </h1>
            <p className="mt-1 text-sm text-ink500">
              Manage organization users, team assignments, and regional configuration.
            </p>
          </div>
          <nav className="flex flex-wrap gap-2">
            {navItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={({ isActive }) =>
                    [
                      "flex items-center gap-2 rounded px-3.5 py-1.5 text-sm font-medium transition-all",
                      isActive
                        ? "bg-ink text-paper shadow-xs font-semibold"
                        : "border border-ink/10 bg-paper-dim text-ink hover:bg-ink/5",
                    ].join(" ")
                  }
                >
                  <Icon className="h-4 w-4 opacity-80" />
                  <span>{item.label}</span>
                </NavLink>
              );
            })}
          </nav>
        </div>
      </div>

      <div>
        <Outlet />
      </div>
    </div>
  );
}
