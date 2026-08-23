import { useState, useEffect } from "react";
import { NavLink, useNavigate, useLocation } from "react-router-dom";
import {
  Clock,
  CheckSquare,
  FileSpreadsheet,
  ShieldCheck,
  BarChart3,
  Settings,
  LogOut,
  Menu,
  X,
  Wifi,
  WifiOff,
  User as UserIcon,
} from "lucide-react";
import {
  getUserRole,
  getCurrentUserName,
  getCurrentUserEmail,
  logout,
  type UserRole,
} from "../lib/api/authClient";

interface NavItem {
  to: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  roles?: UserRole[];
}

const NAV_ITEMS: NavItem[] = [
  {
    to: "/attendance",
    label: "Attendance",
    icon: Clock,
  },
  {
    to: "/tasks",
    label: "Tasks",
    icon: CheckSquare,
  },
  {
    to: "/timesheets",
    label: "Timesheets",
    icon: FileSpreadsheet,
  },
  {
    to: "/approvals",
    label: "Approvals",
    icon: ShieldCheck,
    roles: ["manager", "admin"],
  },
  {
    to: "/dashboard",
    label: "Dashboard",
    icon: BarChart3,
    roles: ["manager", "admin"],
  },
  {
    to: "/admin",
    label: "Administration",
    icon: Settings,
    roles: ["admin"],
  },
];

export function Navbar() {
  const navigate = useNavigate();
  const location = useLocation();
  const [mobileMenuOpen, setMobileMenuOpen] = useState(false);
  const [isOnline, setIsOnline] = useState(
    typeof navigator !== "undefined" ? navigator.onLine : true
  );

  const role = getUserRole();
  const userName = getCurrentUserName();
  const userEmail = getCurrentUserEmail();

  const [prevPath, setPrevPath] = useState(location.pathname);
  if (location.pathname !== prevPath) {
    setPrevPath(location.pathname);
    setMobileMenuOpen(false);
  }

  // Monitor online / offline status
  useEffect(() => {
    const handleOnline = () => setIsOnline(true);
    const handleOffline = () => setIsOnline(false);

    window.addEventListener("online", handleOnline);
    window.addEventListener("offline", handleOffline);

    return () => {
      window.removeEventListener("online", handleOnline);
      window.removeEventListener("offline", handleOffline);
    };
  }, []);

  async function handleLogout() {
    await logout();
    navigate("/login");
  }

  // Filter items according to current role
  const visibleItems = NAV_ITEMS.filter((item) => {
    if (!item.roles) return true;
    if (!role) return false;
    return item.roles.includes(role);
  });

  const getRoleBadge = (r: UserRole | null) => {
    switch (r) {
      case "admin":
        return {
          label: "ADMIN",
          className:
            "bg-ochre/15 text-ochre-dark border border-ochre/30 font-mono font-semibold",
        };
      case "manager":
        return {
          label: "MANAGER",
          className:
            "bg-ink/10 text-ink border border-ink/20 font-mono font-semibold",
        };
      default:
        return {
          label: "EMPLOYEE",
          className:
            "bg-ink500/10 text-ink500 border border-ink500/20 font-mono",
        };
    }
  };

  const roleBadge = getRoleBadge(role);

  return (
    <header className="sticky top-0 z-40 border-b border-ink/10 bg-paper/95 backdrop-blur-md transition-all">
      <div className="mx-auto flex h-16 max-w-7xl items-center justify-between px-4 sm:px-6 lg:px-8">
        {/* Left: Brand Identity */}
        <div className="flex items-center gap-6">
          <NavLink
            to="/tasks"
            className="group flex items-center gap-2.5 text-decoration-none"
          >
            <div className="flex h-9 w-9 items-center justify-center rounded bg-ink text-paper shadow-sm transition-transform group-hover:scale-105">
              <span className="font-display text-base font-bold tracking-wider text-ochre-light">
                H
              </span>
            </div>
            <div>
              <div className="flex items-center gap-1.5">
                <span className="font-display text-base font-bold tracking-tight text-ink">
                  HABESHAN
                </span>
                <span className="font-display text-base font-medium tracking-tight text-ochre">
                  REMS
                </span>
              </div>
              <p className="hidden text-[10px] font-medium uppercase tracking-widest text-ink500 sm:block">
                Remote Workforce System
              </p>
            </div>
          </NavLink>

          {/* Desktop Navigation Links */}
          <nav className="hidden lg:flex lg:items-center lg:gap-1">
            {visibleItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={({ isActive }) =>
                    [
                      "flex items-center gap-2 rounded px-3 py-2 text-sm font-medium transition-all",
                      isActive
                        ? "bg-ink text-paper shadow-xs font-semibold"
                        : "text-ink hover:bg-ink/5 hover:text-ink-dark",
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

        {/* Right: Network Status, User Profile & Actions */}
        <div className="hidden sm:flex sm:items-center sm:gap-4">
          {/* Network status pill */}
          <div
            title={
              isOnline
                ? "Connected to server"
                : "Working offline. Changes are queued locally in IndexedDB."
            }
            className={`flex items-center gap-1.5 rounded-full px-2.5 py-1 text-xs font-medium transition-colors ${
              isOnline
                ? "bg-status-verified/10 text-status-verified border border-status-verified/20"
                : "bg-ochre/15 text-ochre-dark border border-ochre/30 animate-pulse"
            }`}
          >
            {isOnline ? (
              <>
                <Wifi className="h-3 w-3" />
                <span className="font-mono text-[11px]">Online</span>
              </>
            ) : (
              <>
                <WifiOff className="h-3 w-3" />
                <span className="font-mono text-[11px]">Offline Queue</span>
              </>
            )}
          </div>

          {/* User Profile Card */}
          <div className="flex items-center gap-3 border-l border-ink/10 pl-4">
            <div className="flex flex-col items-end">
              <div className="flex items-center gap-2">
                <span className="text-xs font-semibold text-ink">
                  {userName || userEmail || "Authenticated User"}
                </span>
                <span
                  className={`rounded px-1.5 py-0.5 text-[10px] tracking-wide ${roleBadge.className}`}
                >
                  {roleBadge.label}
                </span>
              </div>
              {userEmail && userName && (
                <span className="text-[11px] text-ink500">{userEmail}</span>
              )}
            </div>

            {/* Logout Button */}
            <button
              type="button"
              onClick={handleLogout}
              title="Sign out of Habeshan REMS"
              className="flex h-8 w-8 items-center justify-center rounded border border-ink/10 bg-paper-dim text-ink500 transition-colors hover:border-status-rejected/30 hover:bg-status-rejected/10 hover:text-status-rejected focus:outline-none"
            >
              <LogOut className="h-4 w-4" />
            </button>
          </div>
        </div>

        {/* Mobile Hamburger Button */}
        <div className="flex items-center gap-2 lg:hidden">
          <div
            className={`flex items-center gap-1 rounded px-2 py-0.5 text-[10px] font-mono sm:hidden ${
              isOnline
                ? "bg-status-verified/10 text-status-verified"
                : "bg-ochre/15 text-ochre-dark"
            }`}
          >
            {isOnline ? "Online" : "Offline"}
          </div>

          <button
            type="button"
            onClick={() => setMobileMenuOpen(!mobileMenuOpen)}
            className="flex h-9 w-9 items-center justify-center rounded border border-ink/10 bg-paper-dim text-ink transition-colors hover:bg-ink/5"
            aria-label="Toggle navigation menu"
          >
            {mobileMenuOpen ? (
              <X className="h-5 w-5" />
            ) : (
              <Menu className="h-5 w-5" />
            )}
          </button>
        </div>
      </div>

      {/* Mobile Drawer Navigation Menu */}
      {mobileMenuOpen && (
        <div className="border-b border-ink/10 bg-paper-dim px-4 pt-3 pb-5 shadow-lg lg:hidden">
          {/* User Info on Mobile */}
          <div className="mb-4 flex items-center justify-between border-b border-ink/10 pb-3">
            <div className="flex items-center gap-2.5">
              <div className="flex h-8 w-8 items-center justify-center rounded-full bg-ink text-xs font-bold text-paper">
                <UserIcon className="h-4 w-4" />
              </div>
              <div>
                <p className="text-xs font-semibold text-ink">
                  {userName || userEmail || "User"}
                </p>
                <span
                  className={`inline-block rounded px-1.5 py-0.2 text-[9px] ${roleBadge.className}`}
                >
                  {roleBadge.label}
                </span>
              </div>
            </div>
            <button
              type="button"
              onClick={handleLogout}
              className="flex items-center gap-1.5 rounded px-2.5 py-1.5 text-xs font-medium text-status-rejected hover:bg-status-rejected/10"
            >
              <LogOut className="h-3.5 w-3.5" />
              <span>Sign out</span>
            </button>
          </div>

          {/* Nav items */}
          <nav className="flex flex-col gap-1">
            {visibleItems.map((item) => {
              const Icon = item.icon;
              return (
                <NavLink
                  key={item.to}
                  to={item.to}
                  className={({ isActive }) =>
                    [
                      "flex items-center gap-3 rounded px-3 py-2.5 text-sm font-medium transition-all",
                      isActive
                        ? "bg-ink text-paper font-semibold"
                        : "text-ink hover:bg-ink/5",
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
      )}
    </header>
  );
}
