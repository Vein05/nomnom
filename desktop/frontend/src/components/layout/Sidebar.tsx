import type { NavItem, ViewRoute } from "../../lib/types";

interface SidebarProps {
  route: ViewRoute;
  navItems: NavItem[];
  onRouteChange: (route: ViewRoute) => void;
}

export function Sidebar({ route, navItems, onRouteChange }: SidebarProps) {
  return (
    <aside className="flex w-16 shrink-0 flex-col items-center gap-2 border-r border-border bg-sidebar px-2 py-4">
      {navItems.map((item) => {
        const active = route === item.route;
        return (
          <button
            key={item.route}
            title={item.label}
            onClick={() => onRouteChange(item.route)}
            aria-current={active ? "page" : undefined}
            className={`group flex h-11 w-11 items-center justify-center rounded-2xl border transition-all duration-150 ${
              active
                ? "border-border bg-sidebar-active text-text shadow-[0_8px_18px_rgba(0,0,0,0.16)] ring-1 ring-accent/10"
                : "border-transparent text-muted hover:border-border hover:bg-surface-2 hover:text-text"
            }`}
          >
            <item.Icon className="h-4 w-4 transition-transform duration-150 group-hover:-translate-y-px" />
          </button>
        );
      })}
    </aside>
  );
}
