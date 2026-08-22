import { forwardRef, type SelectHTMLAttributes } from "react";

interface SelectOption {
  value: string;
  label: string;
}

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  label?: string;
  options: SelectOption[];
  error?: string;
  hint?: string;
}

export const Select = forwardRef<HTMLSelectElement, SelectProps>(
  function Select(
    { label, options, error, hint, id, className = "", ...props },
    ref
  ) {
    const selectId =
      id ?? (label ? label.toLowerCase().replace(/\s+/g, "-") : undefined);

    return (
      <div className="flex flex-col gap-1">
        {label && (
          <label htmlFor={selectId} className="text-sm font-medium text-ink">
            {label}
          </label>
        )}
        <select
          ref={ref}
          id={selectId}
          aria-label={!label ? props["aria-label"] : undefined}
          aria-invalid={error ? true : undefined}
          className={[
            "rounded border bg-paper px-3 py-2 text-base text-ink900",
            "font-body focus-visible:outline focus-visible:outline-2",
            "focus-visible:outline-offset-0 focus-visible:outline-ink",
            error ? "border-status-rejected" : "border-ink/20",
            className,
          ].join(" ")}
          {...props}
        >
          {options.map((option) => (
            <option key={option.value} value={option.value}>
              {option.label}
            </option>
          ))}
        </select>
        {error ? (
          <p className="text-sm text-status-rejected">{error}</p>
        ) : hint ? (
          <p className="text-sm text-ink500">{hint}</p>
        ) : null}
      </div>
    );
  }
);
