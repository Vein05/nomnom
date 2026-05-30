import { ArrowLeft, ArrowRight } from "lucide-react";
import { useCallback, useEffect, useState } from "react";
import { Button } from "./Button";
import { Input } from "./Input";
import { Select } from "./Select";
import { wails } from "../../lib/wails";
import type { DesktopConfig } from "../../lib/types";
import { useConfig } from "../../hooks/useConfig";
import { useToast } from "./ToastProvider";

const STORAGE_KEY = "nomnom-onboarding-v2";

const caseOptions = [
  { value: "snake", label: "snake_case" },
  { value: "camelCase", label: "camelCase" },
  { value: "kebab-case", label: "kebab-case" },
  { value: "PascalCase", label: "PascalCase" },
  { value: "lowercase", label: "lowercase" },
  { value: "UPPERCASE", label: "UPPERCASE" },
];

const aiDefaults = {
  provider: "openrouter" as string,
  model: "",
  api_key: "",
  max_tokens: 128,
  temperature: 0.2,
  vision: { enabled: true, max_image_size: "10MB" },
  prompt: "",
};

function safeAi(cfg: DesktopConfig | null) {
  if (!cfg?.ai) return { ...aiDefaults };
  return {
    provider: cfg.ai.provider || aiDefaults.provider,
    model: cfg.ai.model || aiDefaults.model,
    api_key: cfg.ai.api_key || aiDefaults.api_key,
    max_tokens: cfg.ai.max_tokens ?? aiDefaults.max_tokens,
    temperature: cfg.ai.temperature ?? aiDefaults.temperature,
    vision: {
      enabled: cfg.ai.vision?.enabled ?? aiDefaults.vision.enabled,
      max_image_size: cfg.ai.vision?.max_image_size || aiDefaults.vision.max_image_size,
    },
    prompt: cfg.ai.prompt || aiDefaults.prompt,
  };
}

// ─── Step Props ──────────────────────────────────────────────────────

interface StepProps {
  config: DesktopConfig | null;
  onConfig: (patch: Record<string, any>) => void;
  onNext: () => void;
  onBack: () => void;
  onDone: () => void;
  onSkip: () => void;
  step: number;
  total: number;
}

// ─── Progress Dots ───────────────────────────────────────────────────

function ProgressDots({ step, total }: { step: number; total: number }) {
  return (
    <div className="flex items-center gap-1.5">
      {Array.from({ length: total }).map((_, i) => (
        <div
          key={i}
          className={`h-1.5 rounded-full transition-all duration-300 ${
            i === step ? "w-5 bg-accent" : i < step ? "w-1.5 bg-accent/40" : "w-1.5 bg-border"
          }`}
        />
      ))}
    </div>
  );
}

// ─── Step 1: AI Provider ────────────────────────────────────────────

function StepAIProvider({ config, onConfig, onNext, onSkip, step, total }: StepProps) {
  const { notify } = useToast();
  const [testing, setTesting] = useState(false);
  const [ollamaStatus, setOllamaStatus] = useState("");

  const ai = safeAi(config);
  const provider = ai.provider;

  async function checkOllama() {
    setTesting(true);
    setOllamaStatus("");
    try {
      const status = await wails.getAIStatus();
      if (status.ollama.running) {
        setOllamaStatus(
          status.ollama.model_available
            ? "Ollama is running and the model is available."
            : "Ollama is running but the selected model is not found.",
        );
      } else {
        setOllamaStatus(status.ollama.message || "Ollama is not running.");
      }
    } catch {
      setOllamaStatus("Could not reach Ollama. Make sure it is installed and running.");
    } finally {
      setTesting(false);
    }
  }

  function setProvider(p: string) {
    onConfig({
      ai: { ...ai, provider: p, model: "", api_key: p === "openrouter" ? ai.api_key : "" },
    });
  }

  return (
    <div className="flex flex-col gap-6" style={{ animation: "view-enter 350ms ease-out" }}>
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-[-0.02em] text-text-primary">
          Choose your AI Provider
        </h2>
        <p className="text-sm leading-6 text-text-secondary">
          NomNom uses AI to generate smart file names. Pick how you want to power it.
        </p>
      </div>

      <div className="grid gap-3">
        <button
          type="button"
          onClick={() => setProvider("openrouter")}
          className={`flex flex-col gap-1 rounded-xl border p-4 text-left transition-all duration-200 ${
            provider === "openrouter"
              ? "border-accent bg-accent-subtle/40 shadow-[0_0_0_1px_rgba(118,181,224,0.12)]"
              : "border-border bg-surface-2/30 hover:border-accent/30 hover:bg-surface-2/50"
          }`}
        >
          <span className="text-sm font-semibold text-text-primary">OpenRouter</span>
          <span className="text-xs leading-5 text-text-secondary">
            Cloud AI — use any model via API key. Best quality, needs internet.
          </span>
        </button>

        <button
          type="button"
          onClick={() => setProvider("ollama")}
          className={`flex flex-col gap-1 rounded-xl border p-4 text-left transition-all duration-200 ${
            provider === "ollama"
              ? "border-accent bg-accent-subtle/40 shadow-[0_0_0_1px_rgba(118,181,224,0.12)]"
              : "border-border bg-surface-2/30 hover:border-accent/30 hover:bg-surface-2/50"
          }`}
        >
          <span className="text-sm font-semibold text-text-primary">Ollama</span>
          <span className="text-xs leading-5 text-text-secondary">
            Local AI — runs on your machine. Private, no internet needed.
          </span>
        </button>
      </div>

      {provider === "openrouter" && (
        <div className="space-y-3 rounded-xl border border-border bg-surface-2/20 p-4" style={{ animation: "view-enter 250ms ease-out" }}>
          <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">
            API Key
          </div>
          <Input
            mono
            value={ai.api_key}
            onChange={(e) => onConfig({ ai: { ...ai, api_key: e.target.value } })}
            placeholder="sk-or-..."
          />
          <p className="text-[11px] leading-5 text-text-secondary">
            Get your key at{" "}
            <span className="text-accent">openrouter.ai/keys</span>. Or set the{" "}
            <span className="mono text-text-primary">OPENROUTER_API_KEY</span> environment
            variable and it will be picked up automatically.
          </p>
        </div>
      )}

      {provider === "ollama" && (
        <div className="space-y-3 rounded-xl border border-border bg-surface-2/20 p-4" style={{ animation: "view-enter 250ms ease-out" }}>
          <div className="flex items-center justify-between">
            <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">
              Connection
            </div>
            <Button variant="outline" onClick={checkOllama} disabled={testing} className="text-[11px]">
              {testing ? "Checking..." : "Test Connection"}
            </Button>
          </div>
          {ollamaStatus && (
            <div
              className={`rounded-lg border px-3 py-2 text-xs ${
                ollamaStatus.includes("not") || ollamaStatus.includes("Could not")
                  ? "border-red-500/20 bg-red-500/5 text-red-400"
                  : "border-green-500/20 bg-green-500/5 text-green-400"
              }`}
            >
              {ollamaStatus}
            </div>
          )}
          <p className="text-[11px] leading-5 text-text-secondary">
            Install Ollama from{" "}
            <span className="text-accent">ollama.com</span>, then pull a model like{" "}
            <span className="mono text-text-primary">ollama pull llama3.2</span>.{" "}
            You can change the model later in Settings.
          </p>
        </div>
      )}

      <div className="flex items-center justify-between pt-2">
        <ProgressDots step={step} total={total} />
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={onSkip} className="text-xs">Skip</Button>
          <Button variant="solid" onClick={onNext} className="inline-flex items-center gap-1.5">
            Next <ArrowRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
}

// ─── Step 2: File Handling ──────────────────────────────────────────

function StepFileHandling({ config, onConfig, onNext, onBack, onSkip, step, total }: StepProps) {
  const hotRename = config?.file_handling?.hot_rename || false;
  const skipDotFiles = config?.file_handling?.skip_dot_files ?? true;

  const fh = config?.file_handling || { max_size: "100MB", auto_approve: true, skip_dot_files: true, hot_rename: false };

  return (
    <div className="flex flex-col gap-6" style={{ animation: "view-enter 350ms ease-out" }}>
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-[-0.02em] text-text-primary">
          How should files be handled?
        </h2>
        <p className="text-sm leading-6 text-text-secondary">
          Decide whether NomNom copies files to a safe output folder or renames them in place.
        </p>
      </div>

      <div className="grid gap-3">
        <button
          type="button"
          onClick={() => onConfig({ file_handling: { ...fh, hot_rename: false } })}
          className={`flex flex-col gap-1 rounded-xl border p-4 text-left transition-all duration-200 ${
            !hotRename
              ? "border-accent bg-accent-subtle/40 shadow-[0_0_0_1px_rgba(118,181,224,0.12)]"
              : "border-border bg-surface-2/30 hover:border-accent/30 hover:bg-surface-2/50"
          }`}
        >
          <span className="text-sm font-semibold text-text-primary">
            Copy to Output Folder
          </span>
          <span className="text-xs leading-5 text-text-secondary">
            Files are copied and renamed into a separate folder. Your originals stay
            untouched. Safe and reversible.
          </span>
        </button>

        <button
          type="button"
          onClick={() => onConfig({ file_handling: { ...fh, hot_rename: true } })}
          className={`flex flex-col gap-1 rounded-xl border p-4 text-left transition-all duration-200 ${
            hotRename
              ? "border-accent bg-accent-subtle/40 shadow-[0_0_0_1px_rgba(118,181,224,0.12)]"
              : "border-border bg-surface-2/30 hover:border-accent/30 hover:bg-surface-2/50"
          }`}
        >
          <span className="text-sm font-semibold text-text-primary">Rename In Place</span>
          <span className="text-xs leading-5 text-text-secondary">
            Files are renamed directly where they are. No copies made. Faster, but modifies
            your original files.
          </span>
        </button>
      </div>

      <div className="rounded-xl border border-amber-500/20 bg-amber-500/5 p-3 text-xs leading-5 text-text-secondary">
        <span className="font-medium text-text-primary">Tip:</span> Start with Copy mode. You
        can always switch to Rename In Place later in Settings once you trust the results.
      </div>

      <div className="flex items-center justify-between pt-2">
        <ProgressDots step={step} total={total} />
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={onBack} className="inline-flex items-center gap-1.5">
            <ArrowLeft className="h-3.5 w-3.5" /> Back
          </Button>
          <Button variant="ghost" onClick={onSkip} className="text-xs">Skip</Button>
          <Button variant="solid" onClick={onNext} className="inline-flex items-center gap-1.5">
            Next <ArrowRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
}

// ─── Step 3: Naming & Organize ──────────────────────────────────────

function StepQuickTour({ config, onConfig, onNext, onBack, onDone, onSkip, step, total }: StepProps) {
  const currentCase = config?.case || "snake";
  const skipLarge = config?.content_extraction?.skip_large_files ?? true;

  const ce = config?.content_extraction || {
    extract_text: true,
    extract_metadata: true,
    max_content_length: 5000,
    skip_large_files: true,
    read_context: true,
  };

  const fh = config?.file_handling || {
    max_size: "100MB",
    auto_approve: true,
    skip_dot_files: true,
    hot_rename: false,
  };

  return (
    <div className="flex flex-col gap-6" style={{ animation: "view-enter 350ms ease-out" }}>
      <div className="space-y-2">
        <h2 className="text-lg font-semibold tracking-[-0.02em] text-text-primary">
          A quick look at your settings
        </h2>
        <p className="text-sm leading-6 text-text-secondary">
          These are the most important options. Fine-tune everything later in Settings.
        </p>
      </div>

      <div className="space-y-4">
        {/* Case style */}
        <div className="rounded-xl border border-border bg-surface-2/20 p-4">
          <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">
            Naming Style
          </div>
          <p className="mt-2 text-xs leading-5 text-text-secondary">
            Choose how your renamed files will look.
          </p>
          <div className="mt-3">
            <Select
              value={currentCase}
              onChange={(e) => onConfig({ case: e.target.value })}
            >
              {caseOptions.map((opt) => (
                <option key={opt.value} value={opt.value}>{opt.label}</option>
              ))}
            </Select>
          </div>
          <div className="mt-2 text-xs text-text-secondary">
            Preview:{" "}
            <span className="mono text-text-primary">
              {currentCase === "snake" && "my_report_2025.pdf"}
              {currentCase === "camelCase" && "myReport2025.pdf"}
              {currentCase === "kebab-case" && "my-report-2025.pdf"}
              {currentCase === "PascalCase" && "MyReport2025.pdf"}
              {currentCase === "lowercase" && "my report 2025.pdf"}
              {currentCase === "UPPERCASE" && "MY REPORT 2025.PDF"}
            </span>
          </div>
        </div>

        {/* File limits */}
        <div className="rounded-xl border border-border bg-surface-2/20 p-4 space-y-4">
          <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">
            File Size &amp; Limits
          </div>

          <div className="grid gap-3 text-xs leading-5 text-text-secondary">
            <div className="flex items-center justify-between">
              <span>Max file size</span>
              <span className="mono text-text-primary">{fh.max_size || "100MB"}</span>
            </div>
            <div className="flex items-center justify-between">
              <span>Max content scan length</span>
              <span className="mono text-text-primary">{ce.max_content_length?.toLocaleString() ?? "5,000"} chars</span>
            </div>
            <div className="flex items-center justify-between">
              <span>Files larger than the limit are scanned by name only.</span>
            </div>
          </div>

          <div className="flex items-center justify-between border-t border-border pt-3">
            <div>
              <div className="text-xs font-medium text-text-primary">Skip large files</div>
              <p className="mt-0.5 text-[11px] leading-4 text-text-secondary">
                Automatically skip files that exceed the size limit.
              </p>
            </div>
            <button
              type="button"
              onClick={() => onConfig({
                content_extraction: {
                  ...ce,
                  skip_large_files: !skipLarge,
                },
              })}
              className={`relative inline-flex h-6 w-11 shrink-0 items-center rounded-full transition-colors duration-150 ${
                skipLarge ? "bg-accent" : "bg-border"
              }`}
            >
              <span
                className={`inline-block h-4 w-4 rounded-full bg-white shadow-sm transition-transform duration-150 ${
                  skipLarge ? "translate-x-[22px]" : "translate-x-1"
                }`}
              />
            </button>
          </div>
        </div>

        {/* Organize note */}
        <div className="rounded-xl border border-border bg-surface-2/20 p-4">
          <div className="text-xs font-semibold uppercase tracking-[0.08em] text-text-secondary">
            Organize into Folders
          </div>
          <p className="mt-2 text-xs leading-5 text-text-secondary">
            When you run a job, NomNom can sort renamed files into category folders like{" "}
            <span className="mono text-text-primary">documents/</span>,{" "}
            <span className="mono text-text-primary">images/</span>, and{" "}
            <span className="mono text-text-primary">other/</span>. This is
            enabled by default on the Run screen and can be toggled before each job.
          </p>
        </div>

      </div>

      <div className="flex items-center justify-between pt-2">
        <ProgressDots step={step} total={total} />
        <div className="flex items-center gap-2">
          <Button variant="ghost" onClick={onBack} className="inline-flex items-center gap-1.5">
            <ArrowLeft className="h-3.5 w-3.5" /> Back
          </Button>
          <Button variant="ghost" onClick={onSkip} className="text-xs">Skip</Button>
          <Button variant="solid" onClick={onNext} className="inline-flex items-center gap-1.5">
            Next <ArrowRight className="h-3.5 w-3.5" />
          </Button>
        </div>
      </div>
    </div>
  );
}

// ─── Step 4: All Set ────────────────────────────────────────────────

function StepDone({ onBack, onDone, step, total }: StepProps) {
  return (
    <div className="flex flex-col items-center justify-center gap-8 py-12 text-center" style={{ animation: "view-enter 350ms ease-out" }}>
      <div className="space-y-3">
        <h2 className="text-xl font-bold tracking-[-0.03em] text-text-primary">
          You are all set!
        </h2>
        <p className="max-w-sm text-sm leading-6 text-text-secondary">
          Open <span className="font-medium text-text-primary">Settings</span> in the sidebar
          anytime to change your AI model, output path, themes, and more.
        </p>
      </div>

      <Button variant="solid" onClick={onDone} className="px-8 py-2.5 text-base">
        Start NomNom-ing
      </Button>

      <div className="flex items-center justify-between w-full pt-4">
        <ProgressDots step={step} total={total} />
        <Button variant="ghost" onClick={onBack} className="inline-flex items-center gap-1.5">
          <ArrowLeft className="h-3.5 w-3.5" /> Back
        </Button>
      </div>
    </div>
  );
}

// ─── Wizard Container ────────────────────────────────────────────────

interface OnboardingWizardProps {
  onDone: () => void;
}

export function OnboardingWizard({ onDone }: OnboardingWizardProps) {
  const { config, save } = useConfig();
  const [step, setStep] = useState(0);
  const [exiting, setExiting] = useState(false);
  const [draft, setDraft] = useState<DesktopConfig | null>(null);

  useEffect(() => {
    if (config && !draft) {
      setDraft(structuredClone(config));
    }
  }, [config, draft]);

  const handleConfig = useCallback((patch: Record<string, any>) => {
    setDraft((prev) => {
      if (!prev) return prev;
      const next: Record<string, any> = structuredClone(prev as any);
      for (const [section, values] of Object.entries(patch)) {
        if (values && typeof values === "object") {
          next[section] = { ...next[section], ...values };
        } else {
          next[section] = values;
        }
      }
      return next as DesktopConfig;
    });
  }, []);

  async function handleDone() {
    if (draft) {
      try { await save(draft); } catch { /* best effort */ }
    }
    setExiting(true);
    setTimeout(() => {
      localStorage.setItem(STORAGE_KEY, "1");
      onDone();
    }, 400);
  }

  function handleSkip() {
    setExiting(true);
    setTimeout(() => {
      localStorage.setItem(STORAGE_KEY, "1");
      onDone();
    }, 400);
  }

  const total = 4;
  const stepProps: StepProps = {
    config: draft,
    onConfig: handleConfig,
    onNext: () => setStep((s) => Math.min(s + 1, total - 1)),
    onBack: () => setStep((s) => Math.max(s - 1, 0)),
    onDone: handleDone,
    onSkip: handleSkip,
    step,
    total,
  };

  return (
    <div
      className="fixed inset-0 z-[90] flex items-center justify-center"
      style={{
        background: "#0b0e14",
        animation: exiting ? "splash-fade-out 400ms ease-out forwards" : undefined,
      }}
    >
      {/* Subtle background glow */}
      <div className="pointer-events-none absolute inset-0 overflow-hidden">
        <div
          className="absolute -top-40 left-1/2 h-[500px] w-[500px] -translate-x-1/2 rounded-full bg-accent/5 blur-[120px]"
          style={{ animation: "view-enter 800ms ease-out" }}
        />
      </div>

      <div className="relative z-10 w-full max-w-lg px-6">
        <div className="mb-8 text-center" style={{ animation: "view-enter 400ms ease-out" }}>
          <h1 className="text-xl font-bold tracking-[-0.03em] text-text-primary">
            Welcome to NomNom
          </h1>
          <p className="mt-2 text-sm leading-6 text-text-secondary">
            Let us set things up in a few seconds so you get the best results.
          </p>
        </div>

        <div className="min-h-[380px]">
          {step === 0 && <StepAIProvider {...stepProps} />}
          {step === 1 && <StepFileHandling {...stepProps} />}
          {step === 2 && <StepQuickTour {...stepProps} />}
          {step === 3 && <StepDone {...stepProps} />}
        </div>
      </div>
    </div>
  );
}

export function hasCompletedOnboarding(): boolean {
  return localStorage.getItem(STORAGE_KEY) === "1";
}
