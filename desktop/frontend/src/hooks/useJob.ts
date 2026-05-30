import { useCallback, useEffect, useState } from "react";
import { wails } from "../lib/wails";
import type { JobStatus, RenameEntry, RunJobOptions } from "../lib/types";

export function useJob() {
  const [jobID, setJobID] = useState<string | null>(null);
  const [scannedDirectory, setScannedDirectory] = useState<string | null>(null);
  const [plan, setPlan] = useState<RenameEntry[]>([]);
  const [status, setStatus] = useState<JobStatus | null>(null);

  const refreshSnapshotFor = useCallback(async (targetJobID: string) => {
    const [nextPlan, nextStatus] = await Promise.all([wails.getPlan(targetJobID), wails.getJobStatus(targetJobID)]);
    setPlan(nextPlan);
    setStatus(nextStatus);
    return nextStatus;
  }, []);

  const scan = useCallback(
    async (sourceDir: string) => {
      const normalizedSourceDir = sourceDir.trim();
      const id = await wails.scanDirectory(normalizedSourceDir);
      setJobID(id);
      setScannedDirectory(normalizedSourceDir);
      await refreshSnapshotFor(id);
      return id;
    },
    [refreshSnapshotFor],
  );

  const generateNames = useCallback(async () => {
    if (!jobID) {
      throw new Error("No job prepared. Scan a directory first.");
    }
    await wails.generateNames(jobID);
    return refreshSnapshotFor(jobID);
  }, [jobID, refreshSnapshotFor]);

  const run = useCallback(
    async (options: RunJobOptions) => {
      if (!jobID) {
        throw new Error("No job prepared. Scan a directory first.");
      }
      await wails.runJob(jobID, options);
      return refreshSnapshotFor(jobID);
    },
    [jobID, refreshSnapshotFor],
  );

  const refreshStatus = useCallback(async () => {
    if (!jobID) {
      return null;
    }
    return refreshSnapshotFor(jobID);
  }, [jobID, refreshSnapshotFor]);

  const cancel = useCallback(async () => {
    if (!jobID) {
      return false;
    }
    const canceled = await wails.cancelJob(jobID);
    await refreshSnapshotFor(jobID);
    return canceled;
  }, [jobID, refreshSnapshotFor]);

  const isRunning = useCallback((state?: string) => {
    return state === "generating" || state === "preview-ready" || state === "running";
  }, []);

  useEffect(() => {
    if (!jobID || !isRunning(status?.state)) {
      return undefined;
    }

    let active = true;
    const interval = window.setInterval(() => {
      void (async () => {
        try {
          const nextStatus = await refreshSnapshotFor(jobID);
          if (!active || isRunning(nextStatus.state)) {
            return;
          }
          window.clearInterval(interval);
        } catch {
          // Ignore transient backend failures while polling.
        }
      })();
    }, 300);

    return () => {
      active = false;
      window.clearInterval(interval);
    };
  }, [jobID, refreshSnapshotFor, isRunning, status?.state]);

  return {
    jobID,
    scannedDirectory,
    plan,
    status,
    scan,
    generateNames,
    run,
    refreshStatus,
    cancel,
  };
}
