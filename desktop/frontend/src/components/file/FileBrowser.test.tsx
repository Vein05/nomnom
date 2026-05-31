import { render, screen, fireEvent } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { FileBrowser } from "./FileBrowser";
import type { RenameEntry } from "../../lib/types";

const wailsMock = vi.hoisted(() => ({
  openFile: vi.fn(),
}));

vi.mock("../../lib/wails", () => ({
  wails: wailsMock,
}));

function entry(
  index: number,
  original: string,
  newName: string,
  type: string,
  status: string,
  sizeBytes?: number,
): RenameEntry {
  return { index, original, new_name: newName, type, status, size_bytes: sizeBytes };
}

const sampleEntries: RenameEntry[] = [
  entry(1, "photo.jpg", "vacation.jpg", "jpg", "Done", 204800),
  entry(2, "doc.pdf", "report.pdf", "pdf", "Pending", 1024000),
];

describe("FileBrowser", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("shows a scanning indicator when scanning is true", () => {
    render(
      <FileBrowser entries={[]} rootDir="" scanning={true} />,
    );
    expect(screen.getByText("Scanning files...")).toBeInTheDocument();
    // Scanning state has priority — empty-state text should not appear
    expect(screen.queryByText("No folder opened")).not.toBeInTheDocument();
  });

  it("shows the empty state when there is no rootDir and no entries", () => {
    render(
      <FileBrowser entries={[]} rootDir="" scanning={false} />,
    );
    expect(screen.getByText("No folder opened")).toBeInTheDocument();
  });

  it("shows the empty state when entries array is empty", () => {
    render(
      <FileBrowser entries={[]} rootDir="/test" scanning={false} />,
    );
    expect(screen.getByText("No folder opened")).toBeInTheDocument();
  });

  it("renders file entries with their display name", () => {
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
      />,
    );
    expect(screen.getByText("vacation.jpg")).toBeInTheDocument();
    expect(screen.getByText("report.pdf")).toBeInTheDocument();
  });

  it("renders the original name as secondary text for entries with renameEntry", () => {
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
      />,
    );
    expect(screen.getByText("photo.jpg")).toBeInTheDocument();
    expect(screen.getByText("doc.pdf")).toBeInTheDocument();
  });

  it("renders status badges for non-pending entries", () => {
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
      />,
    );
    // The first entry has status "Done"
    expect(screen.getByText("Done")).toBeInTheDocument();
    // The second entry has status "Pending" — no badge should render
    expect(screen.queryByText("Pending")).not.toBeInTheDocument();
  });

  it("calls wails.openFile when a file entry is clicked", async () => {
    wailsMock.openFile.mockResolvedValue(undefined);
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
      />,
    );

    // Click the first entry's button
    const fileButton = screen.getByText("vacation.jpg").closest("button")!;
    expect(fileButton).not.toBeNull();
    fireEvent.click(fileButton);

    // wails.openFile is called with the absolute path (rootDir + path)
    expect(wailsMock.openFile).toHaveBeenCalledWith("/test/photo.jpg");
  });

  it("renders the generating state with SlotName components", () => {
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
        generating={true}
      />,
    );
    // SlotName replaces the plain text name with an animated element.
    // The rest of the entry (badges, original names, sizes) still renders.
    expect(screen.getByText("Done")).toBeInTheDocument();
    expect(screen.getByText("200 KB")).toBeInTheDocument();
    expect(screen.getByText("1000 KB")).toBeInTheDocument();
    expect(screen.getByText("photo.jpg")).toBeInTheDocument();
    expect(screen.getByText("jpg")).toBeInTheDocument();
  });

  it("renders file sizes for non-directory entries", () => {
    render(
      <FileBrowser
        entries={sampleEntries}
        rootDir="/test"
        scanning={false}
      />,
    );
    // 204800 bytes = 200 KB
    expect(screen.getByText("200 KB")).toBeInTheDocument();
    // 1024000 bytes = 1000 KB
    expect(screen.getByText("1000 KB")).toBeInTheDocument();
  });
});
