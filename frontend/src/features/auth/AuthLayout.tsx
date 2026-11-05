import { AlertCircle } from "lucide-react";
import type { ButtonHTMLAttributes, ReactNode } from "react";

/**
 * Dark, centred layout shared by the auth screens (login, register,
 * password reset). Auth-only: the rest of the app keeps the light
 * paper/ink theme and the shared Card/Button/Input components.
 */
export function AuthLayout({
  title,
  subtitle,
  children,
  footer,
}: {
  title: string;
  subtitle?: ReactNode;
  children: ReactNode;
  footer?: ReactNode;
}) {
  return (
    <div className="flex min-h-screen flex-col items-center justify-center bg-auth-bg px-4 py-10 selection:bg-ochre/30">
      <div className="flex w-full max-w-[430px] flex-col items-center">
        <img
          src="/icons/icon.svg"
          alt=""
          className="h-[72px] w-[72px] rounded-[20px] ring-1 ring-ochre/60"
        />
        <h1 className="mt-6 text-center font-display text-2xl font-semibold text-paper">
          {title}
        </h1>
        {subtitle && (
          <p className="mt-2 text-center text-base text-auth-muted">
            {subtitle}
          </p>
        )}
      </div>

      <div className="mt-8 w-full max-w-[430px]">{children}</div>

      {footer && (
        <div className="mt-8 text-center text-sm text-auth-muted">
          {footer}
        </div>
      )}

      <p className="mt-10 text-center text-xs text-auth-faint">
        © {new Date().getFullYear()} Habeshan REMS
      </p>
    </div>
  );
}

type AuthButtonVariant = "primary" | "secondary";

interface AuthButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: AuthButtonVariant;
  loading?: boolean;
}

const authButtonVariants: Record<AuthButtonVariant, string> = {
  primary:
    "bg-ochre font-semibold text-ink-dark hover:bg-ochre-light disabled:opacity-60",
  secondary:
    "border border-white/25 bg-transparent text-paper hover:bg-white/5 disabled:opacity-50",
};

/** Full pill button in the auth style. Loading mirrors the shared Button. */
export function AuthButton({
  variant = "primary",
  loading = false,
  disabled,
  className = "",
  children,
  ...props
}: AuthButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      className={[
        "inline-flex items-center justify-center rounded-full px-5 py-3",
        "font-body text-base transition-colors",
        "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ochre",
        "disabled:cursor-not-allowed",
        authButtonVariants[variant],
        className,
      ].join(" ")}
      {...props}
    >
      {loading ? "Please wait…" : children}
    </button>
  );
}

/** Error banner restyled for the dark auth background. */
export function AuthError({ message }: { message: string }) {
  return (
    <div
      role="alert"
      className="mb-5 flex gap-3 rounded-2xl border border-auth-error/35 bg-[#331b18] px-4 py-3"
    >
      <AlertCircle
        className="mt-0.5 h-5 w-5 shrink-0 text-auth-error"
        aria-hidden
      />
      <div>
        <p className="text-sm font-medium text-auth-error">
          Something went wrong
        </p>
        <p className="mt-1 text-sm text-auth-error">{message}</p>
      </div>
    </div>
  );
}
