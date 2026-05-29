import { createContext, useCallback, useContext, useEffect, useRef, useState } from "react";
import type { PropsWithChildren } from "react";
import { Toast } from "./Toast";

type ToastTone = "success" | "error" | "info";

interface ToastItem {
  id: number;
  message: string;
  tone: ToastTone;
  leaving: boolean;
}

interface ToastContextValue {
  notify: (message: string, tone?: ToastTone) => void;
}

const ToastContext = createContext<ToastContextValue | null>(null);

export function useToast() {
  const ctx = useContext(ToastContext);
  if (!ctx) throw new Error("useToast must be used within ToastProvider");
  return ctx;
}

let nextId = 0;

export function ToastProvider({ children }: PropsWithChildren) {
  const [toasts, setToasts] = useState<ToastItem[]>([]);
  const timers = useRef<Map<number, ReturnType<typeof setTimeout>>>(new Map());

  const removeToast = useCallback((id: number) => {
    setToasts((prev) => prev.filter((t) => t.id !== id));
    const timer = timers.current.get(id);
    if (timer) {
      clearTimeout(timer);
      timers.current.delete(id);
    }
  }, []);

  const startExit = useCallback(
    (id: number) => {
      setToasts((prev) => prev.map((t) => (t.id === id ? { ...t, leaving: true } : t)));
      const timer = setTimeout(() => removeToast(id), 200);
      timers.current.set(id, timer);
    },
    [removeToast],
  );

  const notify = useCallback(
    (message: string, tone: ToastTone = "success") => {
      const id = nextId++;
      setToasts((prev) => [...prev, { id, message, tone, leaving: false }]);
      const timer = setTimeout(() => startExit(id), 3000);
      timers.current.set(id, timer);
    },
    [startExit],
  );

  useEffect(() => {
    return () => {
      timers.current.forEach((t) => clearTimeout(t));
    };
  }, []);

  return (
    <ToastContext.Provider value={{ notify }}>
      {children}
      <div className="fixed right-4 top-4 z-50 flex flex-col gap-2" aria-live="polite">
        {toasts.map((t) => (
          <div
            key={t.id}
            style={{
              animation: t.leaving ? "toast-fade-out 200ms ease-in forwards" : "toast-slide-in 300ms ease-out",
            }}
          >
            <Toast tone={t.tone} onClose={() => startExit(t.id)}>
              {t.message}
            </Toast>
          </div>
        ))}
      </div>
    </ToastContext.Provider>
  );
}
