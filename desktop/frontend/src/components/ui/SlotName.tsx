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
  const timerRef = useRef<ReturnType<typeof setInterval> | null>(null);
  const settledRef = useRef(name);

  useEffect(() => {
    if (!spinning) {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
      settledRef.current = name;
      setDisplay(name);
      return;
    }

    const length = Math.max(name.length || settledRef.current.length || 8, 5);

    timerRef.current = setInterval(() => {
      setDisplay(randomString(length));
    }, 50);

    return () => {
      if (timerRef.current) {
        clearInterval(timerRef.current);
        timerRef.current = null;
      }
    };
  }, [spinning, name]);

  return (
    <span
      className={
        spinning
          ? "inline-block rounded bg-accent-subtle px-1 font-medium text-accent"
          : "inline-block font-medium text-accent"
      }
      style={{ transition: "background 200ms, color 200ms" }}
    >
      {display}
    </span>
  );
}
