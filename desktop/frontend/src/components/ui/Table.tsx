import type { PropsWithChildren, ReactNode } from "react";

interface HeaderCell {
  key: string;
  label: ReactNode;
  className?: string;
}

interface TableProps {
  headers: HeaderCell[];
}

export function Table({ headers, children }: PropsWithChildren<TableProps>) {
  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
      <table className="w-full border-collapse text-left text-sm">
        <thead className="border-b border-border bg-surface-raised/60 text-[11px] uppercase tracking-[0.08em] text-text-secondary">
          <tr>
            {headers.map((header) => (
              <th key={header.key} className={`px-4 py-3 font-medium ${header.className ?? ""}`}>
                {header.label}
              </th>
            ))}
          </tr>
        </thead>
        <tbody>{children}</tbody>
      </table>
    </div>
  );
}
