import { ChevronRight, FolderOpen, Loader2, Play, RefreshCw } from "lucide-react";
import { useEffect, useState } from "react";
import { Button } from "../../components/ui/Button";
import { FileBrowser } from "../../components/file/FileBrowser";
import { Input } from "../../components/ui/Input";
import { Toggle } from "../../components/ui/Toggle";
import { useToast } from "../../components/ui/ToastProvider";
import { useConfig } from "../../hooks/useConfig";
import { shortDir } from "../../lib/path";
import { useJob } from "../../hooks/useJob";
import { wails } from "../../lib/wails";

const stepLabels = ["Pick", "Configure", "Preview", "Done"];

function isScanPreview(state?: string) {
  return state === "preview-ready" || state === "ready" || state === "files-ready" || state === "generating";
}

function hasFilesLoaded(state?: string) {
  return state === "files-ready";
}

function isComparisonPreview(state?: string) {
  return state === "running" || state === "complete" || state === "canceled" || state === "failed";
}

function formatJobState(state?: string) {
  switch (state) {
    case "generating":
      return "Thinking...";
    case "preview-ready":
    case "ready":
      return "Ready";
    case "running":
      return "Running";
    case "complete":
      return "Complete";
    case "canceled":
      return "Canceled";
    case "failed":
      return "Failed";
    default:
      return "Idle";
  }
}

function formatFileSize(sizeBytes?: number) {
  if (typeof sizeBytes !== "number") return "-";
  if (sizeBytes < 1024) return `${sizeBytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = sizeBytes / 1024;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  const digits = value >= 10 ? 0 : 1;
  return `${value.toFixed(digits)} ${units[unitIndex]}`;
}

function shortPath(absPath: string, baseDir: string): string {
  if (!absPath || !baseDir) return absPath;
  const normalizedAbs = absPath.replace(/\\/g, "/");
  const normalizedBase = baseDir.replace(/\\/g, "/");
  if (normalizedAbs.startsWith(normalizedBase)) {
    const relative = normalizedAbs.slice(normalizedBase.length);
    return relative.startsWith("/") ? relative.slice(1) : relative;
  }
  return absPath;
}

interface RenameViewProps {
  onOpenSettings: () => void;
  onStepChange: (index: number) => void;
}

export function RenameView({ onOpenSettings, onStepChange }: RenameViewProps) {
  const { config, loading: configLoading } = useConfig();
  const { plan, status, scan, run, scannedDirectory } = useJob();
  const { notify } = useToast();
  const [sourceDir, setSourceDir] = useState("");
  const [logSession, setLogSession] = useState(true);
  const [moveFiles, setMoveFiles] = useState(false);
  const [organize, setOrganize] = useState(true);
  const [scanBusy, setScanBusy] = useState(false);
  const [runBusy, setRunBusy] = useState(false);
  const hasSourceDir = sourceDir.trim().length > 0;
  const currentSourceDir = sourceDir.trim();
  const hasCurrentScan = hasSourceDir && scannedDirectory === currentSourceDir;
  const scanPreview = hasCurrentScan && isScanPreview(status?.state);
  const comparisonPreview = hasCurrentScan && isComparisonPreview(status?.state);
  const isRunning = status?.state === "running";
  const progress = status && status.total > 0 ? Math.min(100, Math.round((status.done / status.total) * 100)) : 0;
  const stepIndex = status?.state === "complete" ? 3 : comparisonPreview || scanPreview ? 2 : hasSourceDir ? 1 : 0;
  const outputPreview = config?.output?.trim() ? config.output : currentSourceDir ? `${currentSourceDir}/nomnom/renamed` : "source/nomnom/renamed";

  useEffect(() => {
    if (config) {
      setMoveFiles(config.file_handling.move_files);
    }
  }, [config]);

  useEffect(() => {
    onStepChange(stepIndex);
  }, [stepIndex, onStepChange]);

  async function handleScan() {
    if (!currentSourceDir) {
      notify("Set a source directory before scanning.", "error");
      return;
    }
    setScanBusy(true);
    try {
      await scan(currentSourceDir);
    } catch (err) {
      notify(err instanceof Error ? err.message : "Scan failed", "error");
    } finally {
      setScanBusy(false);
    }
  }

  async function handleBrowse() {
    try {
      const currentDir = sourceDir.trim();
      const selectedPath = await wails.selectFolder(currentDir);
      if (selectedPath && selectedPath.length > 0) {
        setSourceDir(selectedPath);
        await scan(selectedPath);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : "Folder picker failed";
      console.error("Browse error:", msg, err);
      notify(msg, "error");
    }
  }

  async function handleRun() {
    if (!hasCurrentScan || plan.length === 0) {
      notify("Scan the open folder before running.", "error");
      return;
    }
    setRunBusy(true);
    try {
      await run({
        dry_run: false,
        log_session: logSession,
        auto_approve: true,
        move_files: moveFiles,
        organize,
      });
    } catch (err) {
      notify(err instanceof Error ? err.message : "Run failed", "error");
    } finally {
      setRunBusy(false);
    }
  }

  const showFileBrowser = hasCurrentScan && (scanPreview || comparisonPreview);
  const fileBrowserRoot = scannedDirectory || currentSourceDir;
  const fileBrowserEntries = comparisonPreview || scanPreview ? plan : [];

  return (
    <section className="space-y-6">
      {/* Header */}
      <header className="space-y-3">
        <div className="flex items-start justify-between gap-4">
          <div className="space-y-1">
            <h1 className="text-base font-semibold tracking-[-0.02em] text-text-primary">Rename</h1>
            <p className="max-w-2xl text-xs leading-5 text-text-secondary md:text-sm md:leading-6">
              Open a folder, review the rename plan, and run the job.
            </p>
          </div>
          <div className="rounded-full border border-border bg-surface px-3 py-1 text-xs text-text-secondary">
            {formatJobState(status?.state)}
          </div>
        </div>
      </header>

      {/* Path bar */}
      <div className="flex flex-wrap gap-2">
        <Input
          mono
          value={sourceDir}
          onChange={(event) => setSourceDir(event.target.value)}
          placeholder="/path/to/files"
          className="flex-1"
        />
        <Button onClick={handleBrowse} disabled={scanBusy || runBusy} variant="outline" className="inline-flex items-center gap-2" title="Browse">
          <FolderOpen className="h-4 w-4" />
        </Button>
        {hasSourceDir ? (
          <Button onClick={handleScan} disabled={scanBusy || runBusy} variant="outline" className="inline-flex items-center gap-2" title="Rescan">
            {scanBusy ? <Loader2 className="h-4 w-4 animate-spin" /> : <RefreshCw className="h-4 w-4" />}
          </Button>
        ) : null}
        {hasSourceDir ? (
          <Button
            onClick={handleRun}
            disabled={runBusy || scanBusy || !hasCurrentScan || plan.length === 0}
            variant="solid"
            className="inline-flex items-center gap-2"
            title="Run"
          >
            {isRunning || runBusy ? <Loader2 className="h-4 w-4 animate-spin" /> : <Play className="h-4 w-4" />}
          </Button>
        ) : null}
      </div>

      {/* Main content: File Browser + Run Flags */}
      <div className="grid gap-4 xl:grid-cols-[minmax(0,3fr)_minmax(260px,1fr)]">
        {/* File Browser */}
        <section className="flex min-h-[400px] flex-col rounded-2xl border-2 border-dashed border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <div className="flex items-center justify-between gap-4">
            <div>
              <div className="text-xs font-semibold uppercase tracking-[0.1em] text-text-primary">
                {showFileBrowser ? shortDir(fileBrowserRoot) : "Source Directory"}
              </div>
              <p className="mt-1 text-[11px] leading-5 text-text-secondary md:text-xs">
                {showFileBrowser
                  ? `${plan.length} ${plan.length === 1 ? "file" : "files"}`
                  : "Pick a folder and scan to see its contents."}
              </p>
            </div>
            {showFileBrowser && (
              <div className="text-right">
                <div className="text-xs text-text-secondary">
                  Output: <span className="mono text-text-primary">{outputPreview}</span>
                </div>
                <div className="mt-1 text-[11px] text-text-secondary">
                  {config?.ai.provider} / {config?.ai.model}
                </div>
              </div>
            )}
          </div>

          <div className="mt-4 flex-1 overflow-auto">
            <FileBrowser
              entries={fileBrowserEntries}
              rootDir={fileBrowserRoot}
              scanning={scanBusy}
            />
          </div>

          {!showFileBrowser && (
            <div className="mt-4 flex items-center justify-between border-t border-border pt-4">
              <div className="text-xs text-text-secondary">
                {configLoading ? "Loading config..." : config ? `${config.ai.provider} / ${config.ai.model}` : "No config"}
              </div>
              <button
                type="button"
                onClick={onOpenSettings}
                className="inline-flex items-center gap-1 text-sm text-accent transition-colors hover:text-accent/80"
              >
                View config
                <ChevronRight className="h-4 w-4" />
              </button>
            </div>
          )}
        </section>

        {/* Run Flags */}
        <section className="flex h-full flex-col space-y-3 rounded-2xl border border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <div className="flex items-center justify-between">
            <div className="text-xs font-semibold uppercase tracking-[0.1em] text-text-primary">Run Flags</div>
            {isRunning || runBusy ? (
              <div className="rounded-full bg-accent-subtle px-2.5 py-0.5 text-[10px] font-medium text-accent">{formatJobState(status?.state)}</div>
            ) : null}
          </div>
          {isRunning || runBusy ? (
            <div className="rounded-xl border border-border bg-surface-raised/40 px-3 py-2 text-xs text-text-secondary">
              <div className="flex items-center justify-between">
                <span>Files</span>
                <span className="text-text-primary">{plan.length}</span>
              </div>
              <div className="mt-2 flex items-center justify-between">
                <span>Status</span>
                <span className="text-text-primary">{formatJobState(status?.state)}</span>
              </div>
            </div>
          ) : (
            <div className="flex h-full flex-col gap-3">
              <div className="space-y-1 rounded-xl border border-border bg-surface-raised/25 p-3">
                <Toggle label="Log Session" checked={logSession} onChange={(event) => setLogSession(event.target.checked)} description="Save a session log for history and revert." />
                <Toggle label="Move Files" checked={moveFiles} onChange={(event) => setMoveFiles(event.target.checked)} description="Move files instead of copying them." />
                <Toggle label="Organize" checked={organize} onChange={(event) => setOrganize(event.target.checked)} description="Group output into category folders." />
              </div>
              <div className="rounded-xl border border-border bg-surface-raised/40 px-3 py-2 text-xs text-text-secondary">
                <div className="flex items-center justify-between">
                  <span>Files</span>
                  <span className="text-text-primary">{plan.length}</span>
                </div>
                <div className="mt-2 flex items-center justify-between">
                  <span>Status</span>
                  <span className="text-text-primary">{formatJobState(status?.state)}</span>
                </div>
              </div>
            </div>
          )}
        </section>
      </div>

      {/* Progress bar */}
      {status && isRunning ? (
        <div className="rounded-2xl border border-border bg-surface p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="text-sm font-medium text-text-primary">Renaming in progress</div>
              <div className="mt-1 text-xs text-text-secondary">{status.current_file || status.message || "Processing..."}</div>
            </div>
            <div className="text-right text-xs text-text-secondary">
              <div>{status.done}/{status.total} files</div>
              <div>{progress}%</div>
            </div>
          </div>
          <div className="mt-3 h-1.5 overflow-hidden rounded-full bg-surface-raised">
            <div className="h-full rounded-full bg-accent transition-all duration-200" style={{ width: `${progress}%` }} />
          </div>
        </div>
      ) : null}

      {/* Completion notice */}
      {status && status.state === "complete" ? (
        <div className="rounded-2xl border border-accent/30 bg-accent-subtle p-4 shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
          <div className="flex items-start justify-between gap-4">
            <div>
              <div className="text-sm font-medium text-text-primary">Rename complete</div>
              <div className="mt-1 text-xs text-text-secondary">
                {status.summary.renamed} renamed · {status.summary.skipped} skipped · {status.summary.errors} errors
              </div>
            </div>
            <div className="text-xs text-text-secondary">
              Output: <span className="mono text-text-primary">{outputPreview}</span>
            </div>
          </div>
        </div>
      ) : null}
    </section>
  );
}
