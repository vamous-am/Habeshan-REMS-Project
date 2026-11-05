import {
  forwardRef,
  type ButtonHTMLAttributes,
  type InputHTMLAttributes,
  type ReactNode,
  type SelectHTMLAttributes,
  type TextareaHTMLAttributes,
} from "react";

/**
 * Light grouped form components.
 *
 * These borrow the presentation of the dark auth panel (one grouped
 * panel, label on the left, borderless control on the right, hairline
 * dividers, one pill shaped primary action) in the app's light paper
 * theme. They are used by the Add User, Create Team and Create Task
 * forms. The shared Button, Input, Select and Card components are
 * intentionally left unchanged.
 */

export function GroupedFormPanel({
  title,
  subtitle,
  children,
  className = "",
}: {
  title: string;
  subtitle?: string;
  children: ReactNode;
  className?: string;
}) {
  return (
    <section
      className={[
        "w-full max-w-xl rounded-[20px] border border-ink/10 bg-[#fffefb] p-6 shadow-card sm:p-7",
        className,
      ].join(" ")}
    >
      <header className="mb-5">
        <h3 className="font-display text-xl font-semibold text-ink">
          {title}
        </h3>
        {subtitle && (
          <p className="mt-1 text-sm text-ink500">{subtitle}</p>
        )}
      </header>
      {children}
    </section>
  );
}

/** Grouped panel that holds field rows, divided by hairlines. */
export function FormFieldGroup({ children }: { children: ReactNode }) {
  return (
    <div className="divide-y divide-ink/[0.07] overflow-hidden rounded-[20px] border border-ink/10 bg-paper">
      {children}
    </div>
  );
}

function FieldRow({
  label,
  error,
  hint,
  controlId,
  children,
}: {
  label: string;
  error?: string;
  hint?: string;
  controlId: string;
  children: ReactNode;
}) {
  return (
    <div>
      <div className="flex items-start gap-3 px-4 focus-within:bg-ink/[0.03] sm:px-5">
        <label
          htmlFor={controlId}
          className="w-28 shrink-0 py-3 text-sm font-medium leading-tight text-ink"
        >
          {label}
        </label>
        <div className="min-w-0 flex-1">{children}</div>
      </div>
      {error ? (
        <p
          id={`${controlId}-error`}
          className="px-4 pb-3 text-sm text-status-rejected sm:pl-[132px] sm:pr-5"
        >
          {error}
        </p>
      ) : hint ? (
        <p
          id={`${controlId}-hint`}
          className="px-4 pb-3 text-sm text-ink500 sm:pl-[132px] sm:pr-5"
        >
          {hint}
        </p>
      ) : null}
    </div>
  );
}

const controlClassName =
  "w-full bg-transparent py-3 font-body text-base text-ink900 placeholder:text-ink500/60 focus:outline-none";

interface GroupedInputFieldProps
  extends InputHTMLAttributes<HTMLInputElement> {
  label: string;
  error?: string;
  hint?: string;
}

export const GroupedInputField = forwardRef<
  HTMLInputElement,
  GroupedInputFieldProps
>(function GroupedInputField({ label, error, hint, id, ...props }, ref) {
  const inputId = id ?? label.toLowerCase().replace(/\s+/g, "-");
  return (
    <FieldRow label={label} error={error} hint={hint} controlId={inputId}>
      <input
        ref={ref}
        id={inputId}
        aria-invalid={error ? true : undefined}
        aria-describedby={
          error
            ? `${inputId}-error`
            : hint
              ? `${inputId}-hint`
              : undefined
        }
        className={controlClassName}
        {...props}
      />
    </FieldRow>
  );
});

interface SelectOption {
  value: string;
  label: string;
}

interface GroupedSelectFieldProps
  extends SelectHTMLAttributes<HTMLSelectElement> {
  label: string;
  options: SelectOption[];
  error?: string;
  hint?: string;
}

export const GroupedSelectField = forwardRef<
  HTMLSelectElement,
  GroupedSelectFieldProps
>(function GroupedSelectField(
  { label, options, error, hint, id, ...props },
  ref
) {
  const selectId = id ?? label.toLowerCase().replace(/\s+/g, "-");
  return (
    <FieldRow label={label} error={error} hint={hint} controlId={selectId}>
      <select
        ref={ref}
        id={selectId}
        aria-invalid={error ? true : undefined}
        aria-describedby={
          error
            ? `${selectId}-error`
            : hint
              ? `${selectId}-hint`
              : undefined
        }
        className={controlClassName}
        {...props}
      >
        {options.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </select>
    </FieldRow>
  );
});

interface GroupedTextareaFieldProps
  extends TextareaHTMLAttributes<HTMLTextAreaElement> {
  label: string;
  error?: string;
  hint?: string;
}

export const GroupedTextareaField = forwardRef<
  HTMLTextAreaElement,
  GroupedTextareaFieldProps
>(function GroupedTextareaField({ label, error, hint, id, ...props }, ref) {
  const textareaId = id ?? label.toLowerCase().replace(/\s+/g, "-");
  return (
    <FieldRow label={label} error={error} hint={hint} controlId={textareaId}>
      <textarea
        ref={ref}
        id={textareaId}
        aria-invalid={error ? true : undefined}
        aria-describedby={
          error
            ? `${textareaId}-error`
            : hint
              ? `${textareaId}-hint`
              : undefined
        }
        className={`${controlClassName} resize-y`}
        {...props}
      />
    </FieldRow>
  );
});

type GroupedFormButtonVariant = "primary" | "secondary";

interface GroupedFormButtonProps
  extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: GroupedFormButtonVariant;
  loading?: boolean;
}

const groupedButtonVariants: Record<GroupedFormButtonVariant, string> = {
  primary:
    "bg-ochre font-semibold text-paper hover:bg-ochre-dark disabled:opacity-60",
  secondary:
    "border border-ink/25 bg-transparent text-ink hover:bg-ink/5 disabled:opacity-50",
};

/** Pill shaped button in the grouped form style. */
export function GroupedFormButton({
  variant = "primary",
  loading = false,
  disabled,
  className = "",
  children,
  ...props
}: GroupedFormButtonProps) {
  return (
    <button
      disabled={disabled || loading}
      className={[
        "inline-flex items-center justify-center rounded-full px-5 py-3",
        "font-body text-base transition-colors",
        "focus-visible:outline focus-visible:outline-2 focus-visible:outline-offset-2 focus-visible:outline-ochre",
        "disabled:cursor-not-allowed",
        groupedButtonVariants[variant],
        className,
      ].join(" ")}
      {...props}
    >
      {loading ? "Please wait…" : children}
    </button>
  );
}

/** Small explanatory note shown under a field group. */
export function FormNote({
  title,
  children,
}: {
  title: string;
  children: ReactNode;
}) {
  return (
    <div className="mt-5 rounded-2xl bg-paper-dim px-4 py-3">
      <p className="text-sm font-semibold text-ink">{title}</p>
      <p className="mt-0.5 text-sm text-ink500">{children}</p>
    </div>
  );
}
