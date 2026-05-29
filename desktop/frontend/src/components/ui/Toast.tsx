import { AlertCircle, Check, Info, X } from "lucide-react";
import type { PropsWithChildren } from "react";

type ToastTone = "success" | "error" | "info";

interface ToastProps {
  tone: ToastTone;
  onClose?: () => void;
}

const toneClass: Record<ToastTone, string> = {
  success: "border-success/40",
  error: "border-danger/40",
  info: "border-accent/50",
};

const toneIcon: Record<ToastTone, JSX.Element> = {
  success: <Check className="h-4 w-4 text-success" />,
  error: <AlertCircle className="h-4 w-4 text-danger" />,
  info: <Info className="h-4 w-4 text-accent" />,
};

export function Toast({ tone, onClose, children }: PropsWithChildren<ToastProps>) {
  return (
    <div className={`flex h-10 w-80 items-center gap-2 rounded-lg border bg-surface px-3 text-sm text-text shadow-[0_1px_2px_rgba(0,0,0,0.12)] ${toneClass[tone]}`}>
      {toneIcon[tone]}
      <div className="flex-1 truncate">{children}</div>
      {onClose ? (
        <button className="rounded-md p-1 text-muted transition-colors hover:bg-surface-2 hover:text-text" onClick={onClose}>
          <X className="h-4 w-4" />
        </button>
      ) : null}
    </div>
  );
}
