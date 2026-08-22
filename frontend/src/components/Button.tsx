import { forwardRef, type ButtonHTMLAttributes } from "react";

type ButtonVariant = "primary" | "secondary";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: ButtonVariant;
  loading?: boolean;
}

const variantClasses: Record<ButtonVariant, string> = {
  primary:
    "bg-ochre text-paper hover:bg-ochre-dark disabled:bg-ochre/60",
  secondary:
    "bg-transparent text-ink border border-ink hover:bg-ink/5 disabled:opacity-50",
};

export const Button = forwardRef<HTMLButtonElement, ButtonProps>(
  function Button(
    {
      variant = "primary",
      loading = false,
      disabled,
      className = "",
      children,
      ...props
    },
    ref
  ) {
    return (
      <button
        ref={ref}
        disabled={disabled || loading}
        className={[
          "inline-flex items-center justify-center rounded px-4 py-2",
          "text-sm font-medium font-body transition-colors",
          "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ink",
          "disabled:cursor-not-allowed",
          variantClasses[variant],
          className,
        ].join(" ")}
        {...props}
      >
        {loading ? "Please wait…" : children}
      </button>
    );
  }
);
