import type { ButtonHTMLAttributes, PropsWithChildren } from "react";

type Variant = "ghost" | "outline" | "solid" | "danger";

interface ButtonProps extends ButtonHTMLAttributes<HTMLButtonElement> {
  variant?: Variant;
}

const variantClass: Record<Variant, string> = {
  ghost: "border border-border bg-transparent text-muted hover:border-border/80 hover:bg-surface-2 hover:text-text",
  outline: "border border-accent/60 bg-transparent text-accent hover:bg-accent-subtle",
  solid: "border border-transparent bg-accent px-3 text-accent-foreground hover:opacity-90",
  danger: "border border-danger/60 bg-transparent text-danger hover:bg-danger/10",
};

export function Button({
  variant = "ghost",
  className = "",
  children,
  ...props
}: PropsWithChildren<ButtonProps>) {
  return (
    <button
      className={`inline-flex h-9 items-center justify-center gap-2 rounded-lg px-3.5 text-sm font-medium leading-none transition-all duration-150 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/20 disabled:pointer-events-none disabled:opacity-50 ${variantClass[variant]} ${className}`.trim()}
      {...props}
    >
      {children}
    </button>
  );
}
