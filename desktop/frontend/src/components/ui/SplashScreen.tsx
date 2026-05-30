import { useEffect, useState } from "react";

interface SplashScreenProps {
  onDone: () => void;
}

export function SplashScreen({ onDone }: SplashScreenProps) {
  const [exiting, setExiting] = useState(false);

  useEffect(() => {
    const exitTimer = setTimeout(() => setExiting(true), 1100);
    const doneTimer = setTimeout(() => onDone(), 1500);
    return () => {
      clearTimeout(exitTimer);
      clearTimeout(doneTimer);
    };
  }, [onDone]);

  if (exiting) {
    return (
      <div
        className="fixed inset-0 z-[100] flex items-center justify-center pointer-events-none"
        style={{
          background: "#0b0e14",
          animation: "splash-fade-out 400ms ease-out forwards",
        }}
      >
        <img
          src="/icon.png"
          alt="NomNom"
          className="h-24 w-24 rounded-2xl object-cover"
          style={{ opacity: 0 }}
        />
      </div>
    );
  }

  return (
    <div className="fixed inset-0 z-[100] flex items-center justify-center" style={{ background: "#0b0e14" }}>
      <img
        src="/icon.png"
        alt="NomNom"
        className="h-24 w-24 rounded-2xl object-cover"
        style={{
          animation: "splash-logo-in 900ms cubic-bezier(0.34, 1.56, 0.64, 1) forwards",
        }}
      />
    </div>
  );
}
