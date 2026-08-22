import type { HTMLAttributes, ReactNode } from "react";

interface CardProps extends HTMLAttributes<HTMLDivElement> {
  title?: string;
  subtitle?: string;
  children: ReactNode;
}

export function Card({
  title,
  subtitle,
  children,
  className = "",
  ...props
}: CardProps) {
  return (
    <div
      className={[
        "w-full max-w-md rounded bg-paper p-8 shadow-card border border-ink/10",
        className,
      ].join(" ")}
      {...props}
    >
      {(title || subtitle) && (
        <header className="mb-6">
          {title && (
            <h1 className="font-display text-2xl font-semibold text-ink">
              {title}
            </h1>
          )}
          {subtitle && (
            <p className="mt-1 text-sm text-ink500">{subtitle}</p>
          )}
        </header>
      )}
      {children}
    </div>
  );
}
