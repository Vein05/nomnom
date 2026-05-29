import { useEffect, useState } from "react";

const STORAGE_KEY = "nomnom:font-scale";
const BASE_PX = 16;
const MIN = 0.85;
const MAX = 1.15;
const STEP = 0.05;
const DEFAULT = 1;

function clamp(v: number) {
  return Math.max(MIN, Math.min(MAX, v));
}

function readStored(): number {
  try {
    const raw = window.localStorage.getItem(STORAGE_KEY);
    if (raw !== null) {
      const n = Number(raw);
      if (!Number.isNaN(n)) return clamp(n);
    }
  } catch {
    // ignore
  }
  return DEFAULT;
}

export function useFontScale() {
  const [scale, setScale] = useState(() => (typeof window === "undefined" ? DEFAULT : readStored()));

  useEffect(() => {
    document.documentElement.style.fontSize = `${BASE_PX * scale}px`;
    try {
      window.localStorage.setItem(STORAGE_KEY, String(scale));
    } catch {
      // ignore
    }
  }, [scale]);

  return { scale, setScale, min: MIN, max: MAX, step: STEP };
}
