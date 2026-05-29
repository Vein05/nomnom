import { ChevronDown } from "lucide-react";
import type { SelectHTMLAttributes } from "react";

interface SelectProps extends SelectHTMLAttributes<HTMLSelectElement> {
  mono?: boolean;
}

export function Select({ mono = false, className = "", children, ...props }: SelectProps) {
  return (
    <div className="relative">
      <select
        className={`h-9 w-full appearance-none rounded-lg border border-border bg-surface-2/80 px-3.5 pr-10 text-sm text-text transition-colors duration-150 focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/15 ${mono ? "mono" : ""} ${className}`.trim()}
        {...props}
      >
        {children}
      </select>
      <ChevronDown className="pointer-events-none absolute right-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted" />
    </div>
  );
}