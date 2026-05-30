import { ChevronDown } from "lucide-react";
import { useEffect, useRef, useState } from "react";
import type { SelectHTMLAttributes, ReactNode } from "react";

interface SelectProps extends Omit<SelectHTMLAttributes<HTMLSelectElement>, "children"> {
  mono?: boolean;
  children: ReactNode;
}

export function Select({ mono = false, className = "", children, value, onChange, ...props }: SelectProps) {
  const [open, setOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // Extract <option> children into { value, label } pairs
  const options = extractOptions(children);
  const selectedValue = value != null ? String(value) : (options[0]?.value ?? "");
  const selectedLabel = options.find((o) => o.value === selectedValue)?.label ?? options[0]?.label ?? "";

  // Close on outside click
  useEffect(() => {
    if (!open) return;
    function handleClick(e: MouseEvent) {
      if (containerRef.current && !containerRef.current.contains(e.target as Node)) {
        setOpen(false);
      }
    }
    document.addEventListener("mousedown", handleClick);
    return () => document.removeEventListener("mousedown", handleClick);
  }, [open]);

  // Close on Escape
  useEffect(() => {
    if (!open) return;
    function handleKey(e: KeyboardEvent) {
      if (e.key === "Escape") setOpen(false);
    }
    document.addEventListener("keydown", handleKey);
    return () => document.removeEventListener("keydown", handleKey);
  }, [open]);

  const trigger = (
    <button
      type="button"
      onClick={() => setOpen((v) => !v)}
      className={`flex h-9 w-full items-center justify-between gap-2 rounded-lg border border-border bg-surface-2/80 px-3.5 text-sm text-text transition-colors duration-150 hover:border-accent/40 focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/15 ${mono ? "mono" : ""} ${className}`.trim()}
    >
      <span className="truncate">{selectedLabel || " "}</span>
      <ChevronDown className={`h-4 w-4 shrink-0 text-muted transition-transform duration-150 ${open ? "rotate-180" : ""}`} />
    </button>
  );

  // Hidden native select for form compatibility
  const native = (
    <select
      className="sr-only"
      value={selectedValue}
      onChange={onChange}
      tabIndex={-1}
      aria-hidden
      {...props}
    >
      {children}
    </select>
  );

  return (
    <div ref={containerRef} className="relative">
      {trigger}
      {native}
      {open && (
        <div
          className="absolute left-0 z-50 mt-1 max-h-56 w-full overflow-auto rounded-lg border border-border bg-surface p-1 shadow-lg shadow-black/20"
          style={{ animation: "view-enter 150ms ease-out" }}
        >
          {options.map((opt) => {
            const active = opt.value === selectedValue;
            return (
              <button
                key={opt.value}
                type="button"
                onClick={() => {
                  const event = {
                    target: { value: opt.value },
                    currentTarget: { value: opt.value },
                  } as React.ChangeEvent<HTMLSelectElement>;
                  onChange?.(event);
                  setOpen(false);
                }}
                className={`flex w-full items-center rounded-md px-3 py-2 text-left text-sm transition-colors ${
                  active
                    ? "bg-accent-subtle text-accent font-medium"
                    : "text-text-secondary hover:bg-surface-2 hover:text-text-primary"
                } ${mono ? "mono" : ""}`.trim()}
              >
                {opt.label}
              </button>
            );
          })}
        </div>
      )}
    </div>
  );
}

interface OptionPair {
  value: string;
  label: string;
}

function extractOptions(children: ReactNode): OptionPair[] {
  const opts: OptionPair[] = [];
  walkChildren(children, opts);
  return opts.length > 0 ? opts : [{ value: "", label: "" }];
}

function walkChildren(node: ReactNode, into: OptionPair[]) {
  if (node == null) return;
  if (Array.isArray(node)) {
    for (const child of node) walkChildren(child, into);
    return;
  }
  if (typeof node !== "object" || !("type" in node)) return;
  const el = node as any;
  if (el.type === "option") {
    into.push({
      value: el.props?.value ?? "",
      label: el.props?.children ?? "",
    });
  } else if (el.props?.children) {
    walkChildren(el.props.children, into);
  }
}
