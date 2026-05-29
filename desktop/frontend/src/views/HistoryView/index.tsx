import { useEffect, useState } from "react";
import { Clock3 } from "lucide-react";
import { Table } from "../../components/ui/Table";
import { useToast } from "../../components/ui/ToastProvider";
import { shortDir } from "../../lib/path";
import { wails } from "../../lib/wails";
import type { Session } from "../../lib/types";

export function HistoryView() {
  const [rows, setRows] = useState<Session[]>([]);
  const { notify } = useToast();

  useEffect(() => {
    void (async () => {
      try {
        setRows(await wails.getHistory());
      } catch (err) {
        notify(err instanceof Error ? err.message : "Failed to load history", "error");
      }
    })();
  }, [notify]);

  return (
    <section className="space-y-4">
      <header className="space-y-1">
        <h1 className="text-[15px] font-semibold">History</h1>
        <p className="text-sm text-text-secondary">Sessions are stored locally so you can inspect past runs and review their results.</p>
      </header>

      {rows.length === 0 ? (
        <div className="rounded-2xl border border-border bg-surface p-8 text-center shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <Clock3 className="mx-auto h-6 w-6 text-text-secondary" />
          <div className="mt-3 text-sm font-medium text-text-primary">No sessions yet</div>
          <div className="mt-1 text-sm text-text-secondary">Run a rename job to start building local history.</div>
        </div>
      ) : (
        <Table
          headers={[
            { key: "date", label: "Date" },
            { key: "dir", label: "Directory" },
            { key: "files", label: "Files" },
            { key: "model", label: "Model" },
            { key: "mode", label: "Mode" },
            { key: "status", label: "Status" },
          ]}
        >
          {rows.map((row, index) => (
            <tr key={`${row.date}-${index}`} className="h-10 border-b border-border/70 last:border-b-0 hover:bg-surface-raised">
              <td className="px-4 text-xs text-text-secondary">{row.date}</td>
              <td className="px-4 mono text-xs text-text-secondary">{shortDir(row.directory)}</td>
              <td className="px-4 text-xs text-text-secondary">{row.files}</td>
              <td className="px-4 text-xs text-text-secondary">{row.model}</td>
              <td className="px-4 text-xs text-text-secondary">{row.mode}</td>
              <td className="px-4 text-xs text-text-secondary">{row.status}</td>
            </tr>
          ))}
        </Table>
      )}
    </section>
  );
}
