import { act, renderHook, waitFor } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import type { DesktopConfig } from "../lib/types";
import { useConfig } from "./useConfig";

const wailsMock = vi.hoisted(() => ({
  getConfigPath: vi.fn(),
  getConfig: vi.fn(),
  saveConfig: vi.fn(),
  setConfigPath: vi.fn(),
}));

vi.mock("../lib/wails", () => ({
  wails: wailsMock,
}));

const initialConfig: DesktopConfig = {
  output: "rename",
  case: "snake_case",
  ai: {
    provider: "ollama",
    model: "llama3.1",
    api_key: "",
    max_tokens: 1024,
    temperature: 0.2,
    vision: {
      enabled: false,
      max_image_size: "2mb",
    },
    prompt: "Rename files clearly.",
  },
  file_handling: {
    max_size: "10mb",
    auto_approve: false,
    hot_rename: false,
    skip_dot_files: true,
  },
  content_extraction: {
    extract_text: true,
    extract_metadata: true,
    max_content_length: 4096,
    skip_large_files: true,
    read_context: false,
  },
  performance: {
    ai: {
      workers: 2,
      timeout: "30s",
      retries: 1,
    },
    file: {
      workers: 4,
      timeout: "10s",
      retries: 2,
    },
  },
  logging: {
    enabled: true,
    log_path: "/tmp/nomnom.log",
  },
};

const nextConfig: DesktopConfig = {
  ...initialConfig,
  output: "history",
};

describe("useConfig", () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it("loads the config path and config on mount", async () => {
    wailsMock.getConfigPath.mockResolvedValue("/tmp/nomnom.json");
    wailsMock.getConfig.mockResolvedValue(initialConfig);

    const { result } = renderHook(() => useConfig());

    await waitFor(() => expect(result.current.loading).toBe(false));

    expect(result.current.error).toBeNull();
    expect(result.current.configPath).toBe("/tmp/nomnom.json");
    expect(result.current.config).toEqual(initialConfig);
    expect(wailsMock.getConfigPath).toHaveBeenCalledTimes(1);
    expect(wailsMock.getConfig).toHaveBeenCalledTimes(1);
  });

  it("switches config paths and keeps the returned config", async () => {
    wailsMock.getConfigPath
      .mockResolvedValueOnce("/tmp/nomnom.json")
      .mockResolvedValueOnce("/tmp/alternate.json");
    wailsMock.getConfig.mockResolvedValue(initialConfig);
    wailsMock.setConfigPath.mockResolvedValue(nextConfig);

    const { result } = renderHook(() => useConfig());

    await waitFor(() => expect(result.current.loading).toBe(false));

    await act(async () => {
      await result.current.applyConfigPath("/tmp/alternate.json");
    });

    expect(wailsMock.setConfigPath).toHaveBeenCalledWith("/tmp/alternate.json");
    expect(result.current.configPath).toBe("/tmp/alternate.json");
    expect(result.current.config).toEqual(nextConfig);
  });
});