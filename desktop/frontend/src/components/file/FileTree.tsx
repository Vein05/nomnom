import {
  ChevronRight,
  File,
  FileArchive,
  FileAudio,
  FileImage,
  FileText,
  FileVideo,
  Folder,
  FolderOpen,
} from "lucide-react";
import { useState } from "react";
import { Badge } from "../ui/Badge";
import { wails } from "../../lib/wails";
import type { RenameEntry } from "../../lib/types";

// ─── Tree building ──────────────────────────────────────────────────

interface TreeNode {
  name: string;
  path: string;
  isDir: boolean;
  children: TreeNode[];
  entry?: RenameEntry;
  depth: number;
}

function buildTree(entries: RenameEntry[], rootDir: string): TreeNode[] {
  const root: TreeNode = {
    name: rootDir.split("/").pop() || rootDir,
    path: rootDir,
    isDir: true,
    children: [],
    depth: 0,
  };

  for (const entry of entries) {
    const parts = entry.original.split("/");
    let node = root;

    for (let i = 0; i < parts.length; i++) {
      const isLast = i === parts.length - 1;
      const partName = parts[i];
      const nodePath = parts.slice(0, i + 1).join("/");

      if (isLast) {
        // This is a file — use the new_name as display, original as part path
        const fileNode: TreeNode = {
          name: entry.new_name || partName,
          path: entry.original,
          isDir: false,
          children: [],
          entry,
          depth: node.depth + 1,
        };
        node.children.push(fileNode);
      } else {
        // Directory
        let child = node.children.find(
          (c) => c.name === partName && c.isDir,
        );
        if (!child) {
          child = {
            name: partName,
            path: nodePath,
            isDir: true,
            children: [],
            depth: node.depth + 1,
          };
          node.children.push(child);
        }
        node = child;
      }
    }
  }

  // Sort: directories first, then files; each alphabetically
  const sortNodes = (nodes: TreeNode[]) => {
    nodes.sort((a, b) => {
      if (a.isDir !== b.isDir) return a.isDir ? -1 : 1;
      return a.name.localeCompare(b.name, undefined, { sensitivity: "base" });
    });
    for (const node of nodes) {
      if (node.children.length > 0) sortNodes(node.children);
    }
  };
  sortNodes(root.children);

  return root.children.length > 0 ? [root] : [];
}

// ─── Icons ──────────────────────────────────────────────────────────

function fileIcon(entry: RenameEntry) {
  const ext = entry.type?.toLowerCase() || "";
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

function iconColor(entry: RenameEntry): string {
  const ext = entry.type?.toLowerCase() || "";
  if (["jpg", "jpeg", "png", "gif", "bmp", "webp", "svg"].includes(ext))
    return "text-amber-400";
  if (["mp3", "wav", "flac", "m4a", "aac", "ogg"].includes(ext))
    return "text-violet-400";
  if (["mp4", "mov", "avi", "mkv", "wmv"].includes(ext))
    return "text-rose-400";
  if (["pdf"].includes(ext)) return "text-red-400";
  if (["doc", "docx", "txt", "md", "rtf"].includes(ext))
    return "text-sky-400";
  if (["zip", "rar", "7z", "tar", "gz"].includes(ext))
    return "text-yellow-400";
  return "text-text-secondary";
}

// ─── Formatting ─────────────────────────────────────────────────────

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
  return `${value.toFixed(value >= 10 ? 0 : 1)} ${units[unitIndex]}`;
}

// ─── Tree node component ───────────────────────────────────────────

function TreeNodeRow({
  node,
  rootDir,
}: {
  node: TreeNode;
  rootDir: string;
}) {
  const [expanded, setExpanded] = useState(node.depth <= 1);
  const hasChildren = node.children.length > 0;
  const Icon = node.isDir
    ? expanded
      ? FolderOpen
      : Folder
    : node.entry
      ? fileIcon(node.entry)
      : File;

  const iconCls = node.isDir
    ? "text-accent"
    : node.entry
      ? iconColor(node.entry)
      : "text-text-secondary";

  async function handleClick() {
    if (node.isDir) {
      if (hasChildren) setExpanded(!expanded);
      return;
    }
    if (!node.entry) return;
    const absPath = `${rootDir}/${node.entry.original}`;
    try {
      await wails.openFile(absPath);
    } catch {
      // File might not exist or OS can't open it — ignore
    }
  }

  return (
    <div>
      <button
        type="button"
        onClick={handleClick}
        className="group flex w-full items-center gap-2 rounded-lg px-3 py-2 text-left transition-colors hover:bg-surface-raised"
        style={{ paddingLeft: `${node.depth * 16 + 12}px` }}
      >
        {/* Expand/collapse arrow for directories */}
        <span className="flex h-4 w-4 shrink-0 items-center justify-center">
          {node.isDir && hasChildren ? (
            <ChevronRight
              className={`h-3.5 w-3.5 text-text-secondary transition-transform ${
                expanded ? "rotate-90" : ""
              }`}
            />
          ) : null}
        </span>

        {/* Icon */}
        <Icon className={`h-4 w-4 shrink-0 ${iconCls}`} />

        {/* File info */}
        <div className="min-w-0 flex-1">
          <div className="flex items-center gap-2">
            <span
              className={`truncate text-sm font-medium ${
                node.isDir ? "text-text-primary" : "text-text-primary"
              }`}
            >
              {node.name}
            </span>
            {node.entry?.status && node.entry.status !== "Pending" && (
              <Badge
                tone={
                  node.entry.status === "Done"
                    ? "done"
                    : node.entry.status === "Error"
                      ? "error"
                      : node.entry.status === "Skipped"
                        ? "skipped"
                        : "ai"
                }
              >
                {node.entry.status}
              </Badge>
            )}
          </div>
          {node.entry && (
            <div className="mt-0.5 flex items-center gap-3">
              <span className="truncate text-xs text-text-secondary">
                was: {node.entry.original.split("/").pop()}
              </span>
              <span className="shrink-0 text-xs text-text-secondary">
                {formatFileSize(node.entry.size_bytes)}
              </span>
              <span className="shrink-0 text-[11px] uppercase text-text-secondary/60">
                {node.entry.type || "-"}
              </span>
            </div>
          )}
          {node.isDir && (
            <div className="mt-0.5 text-xs text-text-secondary">
              {hasChildren
                ? `${node.children.length} ${
                    node.children.length === 1 ? "item" : "items"
                  }`
                : "Empty"}
            </div>
          )}
        </div>
      </button>

      {/* Children */}
      {node.isDir && expanded && hasChildren && (
        <div>
          {node.children.map((child) => (
            <TreeNodeRow
              key={child.path + child.name}
              node={child}
              rootDir={rootDir}
            />
          ))}
        </div>
      )}
    </div>
  );
}

// ─── Main component ─────────────────────────────────────────────────

interface FileTreeProps {
  entries: RenameEntry[];
  rootDir: string;
}

export function FileTree({ entries, rootDir }: FileTreeProps) {
  if (entries.length === 0) {
    return (
      <div className="flex flex-col items-center justify-center rounded-2xl border border-border bg-surface px-6 py-16 text-center shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
        <Folder className="mx-auto h-8 w-8 text-text-secondary/50" />
        <div className="mt-4 text-sm font-medium text-text-primary">
          No files to show
        </div>
        <div className="mt-1 text-xs text-text-secondary">
          Scan a folder to populate the file tree.
        </div>
      </div>
    );
  }

  const tree = buildTree(entries, rootDir);

  if (tree.length === 0) {
    return null;
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-surface shadow-[0_1px_2px_rgba(0,0,0,0.12)]">
      {/* Header */}
      <div className="flex items-center justify-between border-b border-border px-4 py-3">
        <span className="text-xs font-medium text-text-secondary">
          {entries.length} {entries.length === 1 ? "file" : "files"}
        </span>
        <span className="truncate text-xs text-text-secondary">
          {rootDir.split("/").pop() || rootDir}
        </span>
      </div>

      {/* Tree */}
      <div className="py-1">
        {tree.map((node) => (
          <TreeNodeRow
            key={node.path}
            node={{ ...node, depth: 0 }}
            rootDir={rootDir}
          />
        ))}
      </div>
    </div>
  );
}
