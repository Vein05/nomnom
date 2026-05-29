import type { InputHTMLAttributes } from "react";

interface InputProps extends InputHTMLAttributes<HTMLInputElement> {
  mono?: boolean;
}

export function Input({ mono = false, className = "", ...props }: InputProps) {
  return (
    <input
      className={`h-9 w-full rounded-lg border border-border bg-surface-2/80 px-3.5 text-sm text-text placeholder:text-muted/70 transition-colors duration-150 focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/15 ${mono ? "mono" : ""} ${className}`.trim()}
      {...props}
    />
  );
}
