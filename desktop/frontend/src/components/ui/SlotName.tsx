import { useEffect, useRef, useState } from "react";

const CHARS = "abcdefghijklmnopqrstuvwxyz0123456789_-.";

function randomChar() {
  return CHARS[Math.floor(Math.random() * CHARS.length)];
}

function randomString(len: number) {
  let s = "";
  for (let i = 0; i < len; i++) s += randomChar();
  return s;
}

interface SlotNameProps {
  name: string;
  spinning: boolean;
}

export function SlotName({ name, spinning }: SlotNameProps) {
  const [display, setDisplay] = useState(name);
  const [settling, setSettling] = useState(false);
  const timerRef = useRef<ReturnType<typeof setTimeout> | null>(null);

  useEffect(() => {
    if (spinning) {
      setSettling(false);
      const length = Math.max(name.length, 5);
      let ticking = true;

      function tick() {
        if (!ticking) return;
        setDisplay(randomString(length));
        timerRef.current = setTimeout(tick, 50);
      }
      tick();

      return () => {
        ticking = false;
        if (timerRef.current) {
          clearTimeout(timerRef.current);
          timerRef.current = null;
        }
      };
    }

    // Settling — decelerate and reveal final name character by character
    const final = name || "";
    const length = Math.max(final.length, 5);
    setSettling(true);
    let step = 0;
    let cancelled = false;

    function settle() {
      if (cancelled) return;

      if (step >= length) {
        setDisplay(final);
        setSettling(false);
        return;
      }

      const locked = final.slice(0, step);
      const filler = randomString(Math.max(length - step, 0));
      setDisplay((locked + filler).slice(0, length));

      step++;
      // Progressively slower: 50ms → ~1500ms over the length
      const delay = 50 + step * (1400 / length);
      timerRef.current = setTimeout(settle, delay);
    }

    settle();

    return () => {
      cancelled = true;
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [spinning, name]);

  const active = spinning || settling;

  return (
    <span
      className={
        active
          ? "inline-block rounded bg-accent-subtle px-1 font-medium text-accent"
          : "inline-block font-medium text-accent"
      }
      style={{ transition: "background 200ms, color 200ms" }}
    >
      {display}
    </span>
  );
}
