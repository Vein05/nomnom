import type { PropsWithChildren } from "react";
import { Sidebar } from "./Sidebar";
import { Titlebar } from "./Titlebar";
import type { NavItem, ViewRoute } from "../../lib/types";
import type { ThemeName } from "../../lib/theme";

interface ShellProps {
  route: ViewRoute;
  navItems: NavItem[];
  onRouteChange: (route: ViewRoute) => void;
  theme: ThemeName;
  onToggleTheme: () => void;
  onOpenSettings: () => void;
  stepIndex?: number;
}

export function Shell({
  route,
  navItems,
  onRouteChange,
  theme,
  onToggleTheme,
  onOpenSettings,
  stepIndex,
  children,
}: PropsWithChildren<ShellProps>) {
  return (
    <div className="flex h-full flex-col overflow-hidden bg-bg text-text-primary">
      <Titlebar theme={theme} onToggleTheme={onToggleTheme} onOpenSettings={onOpenSettings} stepIndex={stepIndex} />
      <div className="flex min-h-0 flex-1 overflow-hidden bg-bg">
        <Sidebar route={route} navItems={navItems} onRouteChange={onRouteChange} />
        <main className="min-h-0 flex-1 overflow-auto px-5 py-5 lg:px-6 lg:py-6">{children}</main>
      </div>
    </div>
  );
}
