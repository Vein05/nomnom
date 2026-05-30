import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { JobStatus, RenameEntry, RunJobOptions } from "../lib/types";
import { useJob } from "./useJob";

const wailsMock = vi.hoisted(() => ({
  scanDirectory: vi.fn(),
  getPlan: vi.fn(),
  getJobStatus: vi.fn(),
  runJob: vi.fn(),
  cancelJob: vi.fn(),
}));

vi.mock("../lib/wails", () => ({
  wails: wailsMock,
}));

const plan: RenameEntry[] = [
  {
    index: 1,
    original: "IMG_001.png",
    new_name: "receipt.png",
    type: "file",
    status: "pending",
  },
];

const scanStatus: JobStatus = {
  job_id: "job-123",
  state: "ready",
  done: 0,
  total: 1,
  current_file: "",
  message: "Ready to run",
  summary: {
    planned: 1,
    renamed: 0,
    skipped: 0,
    errors: 0,
  },
};

const runningStatus: JobStatus = {
  ...scanStatus,
  state: "running",
  done: 1,
  current_file: "IMG_001.png",
  message: "Renaming file",
};

const runOptions: RunJobOptions = {
  dry_run: false,
  log_session: true,
  auto_approve: false,
  move_files: false,
  organize: false,
};

describe("useJob", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("prepares a job from a directory scan", async () => {
    wailsMock.scanDirectory.mockResolvedValue("job-123");
    wailsMock.getPlan.mockResolvedValue(plan);
    wailsMock.getJobStatus.mockResolvedValue(scanStatus);

    const { result } = renderHook(() => useJob());

    let jobId = "";
    await act(async () => {
      jobId = await result.current.scan("/source");
    });

    expect(jobId).toBe("job-123");
    expect(result.current.jobID).toBe("job-123");
    expect(result.current.plan).toEqual(plan);
    expect(result.current.status).toEqual(scanStatus);
  });

  it("runs the prepared job and refreshes status", async () => {
    wailsMock.scanDirectory.mockResolvedValue("job-123");
    wailsMock.getPlan.mockResolvedValue(plan);
    wailsMock.getJobStatus
      .mockResolvedValueOnce(scanStatus)
      .mockResolvedValueOnce(runningStatus);
    wailsMock.runJob.mockResolvedValue("started");

    const { result } = renderHook(() => useJob());

    await act(async () => {
      await result.current.scan("/source");
    });

    let nextStatus: JobStatus | null = null;
    await act(async () => {
      nextStatus = await result.current.run(runOptions);
    });

    expect(wailsMock.runJob).toHaveBeenCalledWith("job-123", runOptions);
    expect(nextStatus).toEqual(runningStatus);
    expect(result.current.status).toEqual(runningStatus);
  });

  it("returns false when cancel is called before a scan", async () => {
    const { result } = renderHook(() => useJob());

    await expect(result.current.cancel()).resolves.toBe(false);
    expect(wailsMock.cancelJob).not.toHaveBeenCalled();
  });
});