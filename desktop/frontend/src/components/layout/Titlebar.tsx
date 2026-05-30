import { Moon, Settings2, Sun } from "lucide-react";
import type { ThemeName } from "../../lib/theme";
import { hasLightVariant, isLightTheme } from "../../lib/theme";

const stepLabels = ["Pick", "Config", "Preview", "Done"];

const stepBorder: Record<number, string> = {
  0: "border-sky-400/40",
  1: "border-amber-400/40",
  2: "border-violet-400/40",
  3: "border-emerald-400/40",
};

const stepDot: Record<number, string> = {
  0: "bg-sky-400",
  1: "bg-amber-400",
  2: "bg-violet-400",
  3: "bg-emerald-400",
};

interface TitlebarProps {
  theme: ThemeName;
  onToggleTheme: () => void;
  onOpenSettings: () => void;
  stepIndex?: number;
}

export function Titlebar({
  theme,
  onToggleTheme,
  onOpenSettings,
  stepIndex,
}: TitlebarProps) {
  const canToggle = hasLightVariant(theme);

  return (
    <header className="flex h-12 items-center justify-between border-b border-border bg-sidebar/95 px-4 backdrop-blur-sm">
      <div className="flex min-w-0 items-center gap-3">
        <img src="/icon.png" alt="" className="h-8 w-8 rounded-lg border border-border object-cover" />
        <div className="min-w-0 leading-tight">
          <div className="truncate text-sm font-semibold text-text-primary">NomNom</div>
        </div>
      </div>

      {/* Step indicators — center */}
      {stepIndex !== undefined && (
        <div className="flex items-center gap-1">
          {stepLabels.map((label, index) => {
            const active = index <= stepIndex;
            const done = index < stepIndex;
            return (
              <div key={label} className="flex items-center gap-1">
                {index > 0 && (
                  <div
                    className={`h-px w-4 transition-colors ${
                      done ? "bg-border" : "bg-border/30"
                    }`}
                  />
                )}
                <span
                  className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium transition-all ${
                    active
                      ? `${stepBorder[index]} bg-surface text-text-primary`
                      : "border-transparent text-muted/50"
                  }`}
                >
                  <span
                    className={`h-1.5 w-1.5 rounded-full ${
                      active ? stepDot[index] : "bg-border"
                    }`}
                  />
                  {label}
                </span>
              </div>
            );
          })}
        </div>
      )}

      <div className="flex items-center gap-2">
        <button
          type="button"
          onClick={onOpenSettings}
          className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-border text-text-secondary transition-colors hover:bg-surface-2 hover:text-text-primary"
          aria-label="Open settings"
          title="Open settings"
        >
          <Settings2 className="h-4 w-4" />
        </button>
        {canToggle ? (
          <button
            type="button"
            onClick={onToggleTheme}
            className="inline-flex h-8 w-8 items-center justify-center rounded-lg border border-border text-text-secondary transition-colors hover:bg-surface-2 hover:text-text-primary"
            aria-label={isLightTheme(theme) ? "Switch to dark mode" : "Switch to light mode"}
            title={isLightTheme(theme) ? "Switch to dark mode" : "Switch to light mode"}
          >
            {isLightTheme(theme) ? <Moon className="h-4 w-4" /> : <Sun className="h-4 w-4" />}
          </button>
        ) : null}
      </div>
    </header>
  );
}
