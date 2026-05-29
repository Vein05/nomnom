import { useEffect, useState } from "react";
import { useToast } from "../../components/ui/ToastProvider";
import { wails } from "../../lib/wails";
import type { AnalyticsSummary } from "../../lib/types";

const blankAnalytics: AnalyticsSummary = {
  sessions: 0,
  renamed: 0,
  tokens: 0,
  avg_per_run: 0,
  recent_runs: 0,
  unique_models: 0,
};

export function AnalyticsView() {
  const [stats, setStats] = useState<AnalyticsSummary>(blankAnalytics);
  const [error, setError] = useState<string | null>(null);
  const { notify } = useToast();

  useEffect(() => {
    void (async () => {
      try {
        setStats(await wails.getAnalytics());
      } catch (err) {
        const msg = err instanceof Error ? err.message : "Failed to load analytics";
        setError(msg);
        notify(msg, "error");
      }
    })();
  }, [notify]);

  const cards = [
    { label: "Sessions", value: stats.sessions },
    { label: "Renamed", value: stats.renamed },
    { label: "Tokens", value: stats.tokens },
    { label: "Avg / run", value: stats.avg_per_run },
  ];

  const hasSessions = stats.sessions > 0;

  return (
    <section className="space-y-4">
      <header className="space-y-1">
        <h1 className="text-[15px] font-semibold">Analytics</h1>
        <p className="text-sm text-text-secondary">Local usage summaries collected from the desktop workspace.</p>
      </header>

      {!hasSessions && !error ? (
        <div className="rounded-2xl border border-border bg-surface p-6 text-center shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <p className="text-sm text-text-secondary">No rename sessions yet.</p>
          <p className="mt-1 text-xs text-text-secondary">Run a renaming job to populate analytics.</p>
        </div>
      ) : (
        <>
          <div className="grid grid-cols-1 gap-3 sm:grid-cols-2 xl:grid-cols-4">
            {cards.map((card) => (
              <article key={card.label} className="rounded-2xl border border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
                <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">{card.label}</div>
                <div className="mt-2 text-3xl font-semibold text-text-primary">{card.value}</div>
              </article>
            ))}
          </div>
          <div className="grid gap-4 lg:grid-cols-[minmax(0,1.2fr)_minmax(280px,0.8fr)]">
            <div className="rounded-2xl border border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
              <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">Activity</div>
              <div className="mt-3 flex items-end gap-2">
                <div className="h-16 flex-1 rounded-xl border border-border bg-surface-raised/40" />
                <div className="h-12 flex-1 rounded-xl border border-border bg-surface-raised/40" />
                <div className="h-20 flex-1 rounded-xl border border-border bg-accent-subtle" />
                <div className="h-10 flex-1 rounded-xl border border-border bg-surface-raised/40" />
                <div className="h-24 flex-1 rounded-xl border border-border bg-accent-subtle" />
              </div>
              <div className="mt-3 text-xs text-text-secondary">Recent runs: {stats.recent_runs}</div>
            </div>

            <div className="rounded-2xl border border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
              <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">Summary</div>
              <div className="mt-3 space-y-2 text-sm text-text-secondary">
                <div className="flex items-center justify-between">
                  <span>Unique models</span>
                  <span className="text-text-primary">{stats.unique_models}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span>Renamed per run</span>
                  <span className="text-text-primary">{stats.avg_per_run}</span>
                </div>
                <div className="flex items-center justify-between">
                  <span>Total tokens</span>
                  <span className="text-text-primary">{stats.tokens}</span>
                </div>
              </div>
            </div>
          </div>
        </>
      )}

      {stats.history_error ? (
        <div className="rounded-lg border border-danger/30 bg-danger/5 p-3 text-xs text-danger">
          History write error: {stats.history_error}
        </div>
      ) : null}
    </section>
  );
}
