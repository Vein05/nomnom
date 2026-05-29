import { useEffect, useState } from "react";
import { defaultTheme, getNextTheme, resolveTheme, themeStorageKey, type ThemeName } from "../lib/theme";

export function useTheme() {
  const [theme, setTheme] = useState<ThemeName>(() => {
    if (typeof window === "undefined") {
      return defaultTheme;
    }

    return resolveTheme(window.localStorage.getItem(themeStorageKey));
  });

  useEffect(() => {
    document.documentElement.setAttribute("data-theme", theme);
    window.localStorage.setItem(themeStorageKey, theme);
  }, [theme]);

  function toggleTheme() {
    setTheme((current) => getNextTheme(current));
  }

  return { theme, setTheme, toggleTheme };
}