import type { PropsWithChildren } from "react";

type BadgeTone = "pending" | "ai" | "done" | "error" | "skipped" | "warning";

interface BadgeProps {
  tone: BadgeTone;
}

const toneClass: Record<BadgeTone, string> = {
  pending: "border-border bg-surface-2 text-muted",
  ai: "border-accent/20 bg-accent-subtle text-accent",
  done: "border-success/20 bg-success/10 text-success",
  error: "border-danger/20 bg-danger/10 text-danger",
  skipped: "border-border bg-surface-2 text-muted italic",
  warning: "border-warning/20 bg-warning/10 text-warning",
};

export function Badge({ tone, children }: PropsWithChildren<BadgeProps>) {
  return (
    <span className={`inline-flex items-center rounded-full border px-2.5 py-0.5 text-[11px] font-semibold tracking-[0.06em] ${toneClass[tone]}`}>
      {children}
    </span>
  );
}
