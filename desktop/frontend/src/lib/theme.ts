export const themeOptions = [
  { value: "nomnom-dark", label: "NomNom Dark" },
  { value: "nomnom-light", label: "NomNom Light" },
  { value: "catppuccin-mocha", label: "Catppuccin Mocha" },
  { value: "catppuccin-latte", label: "Catppuccin Latte" },
  { value: "dracula", label: "Dracula" },
  { value: "nord", label: "Nord" },
  { value: "gruvbox", label: "Gruvbox" },
  { value: "tokyo-night", label: "Tokyo Night" },
] as const;

export type ThemeName = (typeof themeOptions)[number]["value"];

export const themeStorageKey = "nomnom:theme";
export const defaultTheme: ThemeName = "nomnom-dark";

export function isThemeName(value: string | null | undefined): value is ThemeName {
  return themeOptions.some((option) => option.value === value);
}

export function resolveTheme(value: string | null | undefined): ThemeName {
  if (value === "dark") {
    return "nomnom-dark";
  }

  if (value === "light") {
    return "nomnom-light";
  }

  return isThemeName(value) ? value : defaultTheme;
}

export function isLightTheme(theme: ThemeName): boolean {
  return theme === "nomnom-light" || theme === "catppuccin-latte";
}

export function getNextTheme(theme: ThemeName): ThemeName {
  switch (theme) {
    case "nomnom-dark":
      return "nomnom-light";
    case "nomnom-light":
      return "nomnom-dark";
    case "catppuccin-mocha":
      return "catppuccin-latte";
    case "catppuccin-latte":
      return "catppuccin-mocha";
    case "dracula":
      return "dracula";
    case "nord":
      return "nord";
    case "gruvbox":
      return "gruvbox";
    case "tokyo-night":
      return "tokyo-night";
  }
}

export function getThemeLabel(theme: ThemeName): string {
  return themeOptions.find((option) => option.value === theme)?.label ?? theme;
}

export function hasLightVariant(theme: ThemeName): boolean {
  switch (theme) {
    case "nomnom-dark":
    case "nomnom-light":
    case "catppuccin-mocha":
    case "catppuccin-latte":
      return true;
    default:
      return false;
  }
}