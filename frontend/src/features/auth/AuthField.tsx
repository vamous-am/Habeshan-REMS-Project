import {
  forwardRef,
  useState,
  type InputHTMLAttributes,
  type ReactNode,
} from "react";

/** Grouped panel that holds AuthField rows, divided by hairlines. */
export function AuthFieldGroup({ children }: { children: ReactNode }) {
  return (
    <div className="divide-y divide-white/[0.07] overflow-hidden rounded-[20px] border border-white/10 bg-auth-panel">
      {children}
    </div>
  );
}

interface AuthFieldProps extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
  /** Show a SHOW/HIDE toggle. Only meaningful for password fields. */
  passwordToggle?: boolean;
}

/**
 * One row of the grouped auth panel: label on the left, borderless
 * input on the right, in the style of the approved dark auth design.
 * Errors and hints render under the row, inside the panel.
 */
export const AuthField = forwardRef<HTMLInputElement, AuthFieldProps>(
  function AuthField(
    { label, error, hint, id, type, passwordToggle = false, ...props },
    ref
  ) {
    const [showPassword, setShowPassword] = useState(false);
    const inputId = id ?? label.toLowerCase().replace(/\s+/g, "-");
    const isPassword = type === "password";
    const inputType = isPassword && showPassword ? "text" : type;

    return (
      <div>
        <div className="flex items-center gap-3 px-4 focus-within:bg-white/[0.03] sm:px-5">
          <label
            htmlFor={inputId}
            className="w-28 shrink-0 py-3 text-sm font-medium leading-tight text-paper"
          >
            {label}
          </label>
          <input
            ref={ref}
            id={inputId}
            type={inputType}
            aria-invalid={error ? true : undefined}
            aria-describedby={
              error
                ? `${inputId}-error`
                : hint
                  ? `${inputId}-hint`
                  : undefined
            }
            className="min-w-0 flex-1 bg-transparent py-3 font-body text-base text-paper placeholder:text-auth-faint focus:outline-none"
            {...props}
          />
          {isPassword && passwordToggle && (
            <button
              type="button"
              onClick={() => setShowPassword((value) => !value)}
              aria-label={showPassword ? "Hide password" : "Show password"}
              className="shrink-0 rounded-full border border-white/20 px-2.5 py-1 text-[11px] font-medium tracking-wider text-auth-muted transition-colors hover:text-paper focus-visible:outline focus-visible:outline-2 focus-visible:outline-ochre"
            >
              {showPassword ? "HIDE" : "SHOW"}
            </button>
          )}
        </div>
        {error ? (
          <p
            id={`${inputId}-error`}
            className="-mt-1 px-4 pb-3 text-sm text-auth-error sm:px-5 sm:pl-[132px]"
          >
            {error}
          </p>
        ) : hint ? (
          <p
            id={`${inputId}-hint`}
            className="-mt-1 px-4 pb-3 text-sm text-auth-muted sm:px-5 sm:pl-[132px]"
          >
            {hint}
          </p>
        ) : null}
      </div>
    );
  }
);
