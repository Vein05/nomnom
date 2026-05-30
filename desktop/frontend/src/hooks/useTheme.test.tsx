import { act, renderHook } from "@testing-library/react";
import { beforeEach, describe, expect, it, vi } from "vitest";
import { defaultTheme, themeStorageKey } from "../lib/theme";
import { useTheme } from "./useTheme";

describe("useTheme", () => {
  beforeEach(() => {
    window.localStorage.clear();
    document.documentElement.removeAttribute("data-theme");
    vi.clearAllMocks();
  });

  it("restores the saved theme and persists toggles", () => {
    window.localStorage.setItem(themeStorageKey, "light");

    const { result } = renderHook(() => useTheme());

    expect(result.current.theme).toBe("nomnom-light");
    expect(document.documentElement.getAttribute("data-theme")).toBe("nomnom-light");
    expect(window.localStorage.getItem(themeStorageKey)).toBe("nomnom-light");

    act(() => {
      result.current.toggleTheme();
    });

    expect(result.current.theme).toBe("nomnom-dark");
    expect(document.documentElement.getAttribute("data-theme")).toBe("nomnom-dark");
    expect(window.localStorage.getItem(themeStorageKey)).toBe("nomnom-dark");
  });

  it("falls back to the default theme for unknown values", () => {
    window.localStorage.setItem(themeStorageKey, "not-a-real-theme");

    const { result } = renderHook(() => useTheme());

    expect(result.current.theme).toBe(defaultTheme);
    expect(document.documentElement.getAttribute("data-theme")).toBe(defaultTheme);
  });
});