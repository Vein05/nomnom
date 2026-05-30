import type { InputHTMLAttributes } from "react";

interface ToggleProps extends Omit<InputHTMLAttributes<HTMLInputElement>, "type"> {
  label: string;
  description?: string;
}

export function Toggle({ label, description, className = "", ...props }: ToggleProps) {
  return (
    <label
      className={`flex cursor-pointer items-center justify-between gap-4 rounded-xl border border-border bg-surface-2/40 px-3 py-2.5 transition-colors duration-150 hover:bg-surface-2/70 ${className}`.trim()}
    >
      <span className="min-w-0">
        <span className="block text-sm font-medium text-text">{label}</span>
        {description ? <span className="mt-0.5 block text-[11px] leading-4 text-muted">{description}</span> : null}
      </span>
      <span className="relative inline-flex h-6 w-11 shrink-0 items-center">
        <input className="peer sr-only" type="checkbox" {...props} />
        <span className="absolute inset-0 rounded-full border border-border bg-bg/70 transition-colors duration-150 peer-checked:border-accent peer-checked:bg-accent-subtle peer-focus-visible:ring-2 peer-focus-visible:ring-accent/20" />
        <span className="absolute left-0.5 h-5 w-5 rounded-full border border-border bg-muted transition-transform duration-150 peer-checked:translate-x-5 peer-checked:border-accent peer-checked:bg-accent" />
      </span>
    </label>
  );
}