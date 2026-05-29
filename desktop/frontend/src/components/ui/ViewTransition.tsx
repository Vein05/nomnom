import type { PropsWithChildren } from "react";

export function ViewTransition({ children }: PropsWithChildren) {
  return (
    <div
      style={{
        animation: "view-enter 250ms ease-out",
      }}
    >
      {children}
    </div>
  );
}
