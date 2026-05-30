import {
  File,
  FileArchive,
  FileAudio,
  FileImage,
  FileText,
  FileVideo,
  Folder,
  Loader2,
} from "lucide-react";
import { Badge } from "../ui/Badge";
import { SlotName } from "../ui/SlotName";
import { wails } from "../../lib/wails";
import type { RenameEntry } from "../../lib/types";

// ─── Icons ──────────────────────────────────────────────────────────

function fileIcon(ext: string) {
  const imageExts = ["jpg", "jpeg", "png", "gif", "bmp", "webp", "svg", "ico"];
  const audioExts = ["mp3", "wav", "flac", "m4a", "aac", "ogg", "wma"];
  const videoExts = ["mp4", "mov", "avi", "mkv", "wmv", "webm"];
  const docExts = ["pdf", "doc", "docx", "txt", "md", "rtf", "csv", "xls", "xlsx", "ppt", "pptx"];
  const archiveExts = ["zip", "rar", "7z", "tar", "gz", "bz2"];

  if (imageExts.includes(ext)) return FileImage;
  if (audioExts.includes(ext)) return FileAudio;
  if (videoExts.includes(ext)) return FileVideo;
  if (docExts.includes(ext)) return FileText;
  if (archiveExts.includes(ext)) return FileArchive;
  return File;
}

function iconColor(ext: string): string {
  if (["jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"].includes(ext)) return "text-amber-400";
  if (["mp3", "wav", "flac", "m4a", "aac", "ogg"].includes(ext)) return "text-violet-400";
  if (["mp4", "mov", "avi", "mkv", "wmv"].includes(ext)) return "text-rose-400";
  if (["pdf"].includes(ext)) return "text-red-400";
  if (["doc", "docx", "txt", "md", "rtf"].includes(ext)) return "text-sky-400";
  if (["zip", "rar", "7z", "tar", "gz"].includes(ext)) return "text-yellow-400";
  return "text-text-secondary";
}

function formatFileSize(sizeBytes?: number) {
  if (typeof sizeBytes !== "number" || sizeBytes <= 0) return "";
  if (sizeBytes < 1024) return `${sizeBytes} B`;
  const units = ["KB", "MB", "GB", "TB"];
  let value = sizeBytes / 1024;
  let unitIndex = 0;
  while (value >= 1024 && unitIndex < units.length - 1) {
    value /= 1024;
    unitIndex += 1;
  }
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
}

// ─── Types ──────────────────────────────────────────────────────────

interface BrowserEntry {
  name: string;
  path: string;
  isDir: boolean;
  size: number;
  ext: string;
  renameEntry?: RenameEntry;
}

function buildBrowserEntries(
  plan: RenameEntry[],
  rootDir: string,
): BrowserEntry[] {
  if (plan.length === 0) return [];

  // Collect unique directories from the plan
  const dirSet = new Set<string>();
  const entries: BrowserEntry[] = [];

  for (const entry of plan) {
    const parts = entry.original.split("/");
    if (parts.length > 1) {
      dirSet.add(parts[0]);
    } else {
      const ext = (entry.type || "").toLowerCase();
      entries.push({
        name: entry.original,
        path: `${rootDir}/${entry.original}`,
        isDir: false,
        size: entry.size_bytes || 0,
        ext,
        renameEntry: entry,
      });
    }
  }

  // Add directories
  for (const dir of dirSet) {
    entries.push({
      name: dir,
      path: `${rootDir}/${dir}`,
      isDir: true,
      size: 0,
      ext: "",
    });
  }

  // Sort: dirs first, then files; each alphabetically
  entries.sort((a, b) => {
    if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
    return a.name.localeCompare(b.name, undefined, { sensitivity: "base" });
  });

  return entries;
}

// ─── Status badge ──────────────────────────────────────────────────

function entryStatusBadge(entry?: RenameEntry) {
  if (!entry) return null;
  const status = entry.status;
  if (status === "Pending") return null;
  const tone =
    status === "Done"
      ? "done"
      : status === "Error"
        ? "error"
        : status === "Skipped"
          ? "skipped"
          : status === "AI"
            ? "ai"
            : "pending";
  return <Badge tone={tone as any}>{status}</Badge>;
}

// ─── Component ─────────────────────────────────────────────────────

interface FileBrowserProps {
  entries: RenameEntry[];
  rootDir: string;
  scanning: boolean;
  generating?: boolean;
  processing?: boolean;
}

export function FileBrowser({ entries, rootDir, scanning, generating, processing }: FileBrowserProps) {
  const browserEntries = buildBrowserEntries(entries, rootDir);

  if (scanning) {
    return (
      <div className="flex flex-col items-center justify-center py-20">
        <Loader2 className="h-6 w-6 animate-spin text-accent" />
        <div className="mt-3 text-sm text-text-secondary">Scanning files...</div>
      </div>
    );
  }

  if (!rootDir || browserEntries.length === 0) {
    return (
      <div className="flex h-full flex-1 flex-col items-center justify-center py-20 text-center">
        <Folder className="mx-auto h-8 w-8 text-text-secondary/40" />
        <div className="mt-3 text-sm font-medium text-text-primary">
          No folder opened
        </div>
        <div className="mt-1 text-xs text-text-secondary">
          Browse to a folder and scan to see its contents here.
        </div>
      </div>
    );
  }

  async function handleClick(entry: BrowserEntry) {
    if (entry.isDir) return; // subdir navigation could be added later
    try {
      await wails.openFile(entry.path);
    } catch {
      // ignore
    }
  }

  return (
    <div className="space-y-0.5">
      {browserEntries.map((entry) => {
        const Icon = entry.isDir ? Folder : fileIcon(entry.ext);
        const colorCls = entry.isDir ? "text-accent" : iconColor(entry.ext);

        return (
          <button
            key={entry.name}
            type="button"
            onClick={() => handleClick(entry)}
            className={`group flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-surface-raised ${
              processing && entry.renameEntry?.status === "Pending"
                ? "animate-pulse"
                : ""
            }`}
          >
            <Icon className={`h-5 w-5 shrink-0 ${colorCls}`} />

            <div className="min-w-0 flex-1">
              <div className="flex items-center gap-2">
                {generating && entry.renameEntry ? (
                  <SlotName
                    name={entry.renameEntry.new_name || entry.name}
                    spinning
                  />
                ) : (
                  <span className="truncate text-sm font-medium text-text-primary">
                    {entry.renameEntry?.new_name || entry.name}
                  </span>
                )}
                {entry.renameEntry && entryStatusBadge(entry.renameEntry)}
              </div>
              {entry.renameEntry ? (
                <div className="mt-0.5 flex items-center gap-2">
                  <span className="truncate text-xs text-text-secondary">
                    {entry.renameEntry.original.split("/").pop()}
                  </span>
                  <span className="shrink-0 text-[11px] uppercase text-text-secondary/60">
                    {entry.ext || "-"}
                  </span>
                </div>
              ) : null}
              {!entry.isDir && !entry.renameEntry && (
                <div className="mt-0.5 flex items-center gap-2">
                  <span className="text-xs text-text-secondary">
                    {entry.ext || "file"}
                  </span>
                </div>
              )}
            </div>

            {!entry.isDir && (
              <span className="shrink-0 text-xs text-text-secondary/60 tabular-nums">
                {formatFileSize(entry.size)}
              </span>
            )}
          </button>
        );
      })}
    </div>
  );
}
