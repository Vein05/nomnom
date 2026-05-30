import { useCallback, useEffect, useState } from "react";
import { wails } from "../lib/wails";
import type { DesktopConfig } from "../lib/types";

export function useConfig() {
  const [config, setConfig] = useState<DesktopConfig | null>(null);
  const [configPath, setConfigPath] = useState("");
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const refresh = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const [path, result] = await Promise.all([wails.getConfigPath(), wails.getConfig()]);
      setConfigPath(path);
      setConfig(result);
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to load config";
      setError(message);
    } finally {
      setLoading(false);
    }
  }, []);

  const save = useCallback(async (nextConfig: DesktopConfig) => {
    await wails.saveConfig(nextConfig);
    setConfig(nextConfig);
  }, []);

  const applyConfigPath = useCallback(async (nextPath: string) => {
    setError(null);
    try {
      const nextConfig = await wails.setConfigPath(nextPath);
      const resolvedPath = await wails.getConfigPath();
      setConfigPath(resolvedPath);
      setConfig(nextConfig);
      return nextConfig;
    } catch (err) {
      const message = err instanceof Error ? err.message : "Failed to switch config";
      setError(message);
      throw err;
    }
  }, []);

  useEffect(() => {
    refresh();
  }, [refresh]);

  return {
    config,
    configPath,
    loading,
    error,
    refresh,
    save,
    applyConfigPath,
  };
}
