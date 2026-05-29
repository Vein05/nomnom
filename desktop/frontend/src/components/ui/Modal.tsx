import type { PropsWithChildren, ReactNode } from "react";
import { X } from "lucide-react";

interface ModalProps {
  open: boolean;
  title: string;
  onClose: () => void;
  footer?: ReactNode;
}

export function Modal({ open, title, onClose, footer, children }: PropsWithChildren<ModalProps>) {
  if (!open) {
    return null;
  }

  return (
    <div className="fixed inset-0 z-40 flex items-center justify-center bg-bg/75 p-4 backdrop-blur-sm">
      <div className="w-full max-w-2xl rounded-2xl border border-border bg-surface p-5 shadow-[0_12px_32px_rgba(0,0,0,0.24)]">
        <header className="mb-4 flex items-center justify-between gap-3">
          <h2 className="text-base font-semibold text-text">{title}</h2>
          <button onClick={onClose} className="rounded-md border border-border p-1.5 text-muted transition-colors hover:bg-surface-2 hover:text-text" aria-label="Close modal">
            <X className="h-4 w-4" />
          </button>
        </header>
        <div className="text-sm leading-6 text-muted">{children}</div>
        {footer ? <footer className="mt-5 flex justify-end gap-2">{footer}</footer> : null}
      </div>
    </div>
  );
}
