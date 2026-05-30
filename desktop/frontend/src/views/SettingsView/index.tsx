import { useEffect, useState } from "react";
import {
  Bot,
  Boxes,
  ChevronDown,
  ChevronUp,
  Loader2,
  Eye,
  EyeOff,
  FileCog,
  FilePlus2,
  FolderOpen,
  Gauge,
  KeyRound,
  SlidersHorizontal,
  Sparkles,
} from "lucide-react";
import { Button } from "../../components/ui/Button";
import { Input } from "../../components/ui/Input";
import { Select } from "../../components/ui/Select";
import { useToast } from "../../components/ui/ToastProvider";
import { Toggle } from "../../components/ui/Toggle";
import { useConfig } from "../../hooks/useConfig";
import { useFontScale } from "../../hooks/useFontScale";
import type { DesktopConfig, OpenRouterKeyStatus, OpenRouterTestResult } from "../../lib/types";
import { themeOptions, type ThemeName } from "../../lib/theme";
import { wails } from "../../lib/wails";

interface SettingsViewProps {
  theme: ThemeName;
  onThemeChange: (theme: ThemeName) => void;
  onDirtyChange?: (dirty: boolean) => void;
}

const providerOptions = ["deepseek", "openrouter", "ollama"];
const caseOptions = ["snake", "snake_case", "camelCase", "kebab-case", "PascalCase", "lowercase", "UPPERCASE", "original"];

const modelSuggestions: Record<string, string[]> = {
  openrouter: [
    "google/gemini-2.5-flash-lite",
    "google/gemini-2.5-flash",
    "qwen/qwen3.6-flash",
    "mistralai/mistral-small-2603",
    "moonshotai/kimi-k2.6",
  ],
  deepseek: ["deepseek-chat", "deepseek-reasoner"],
  ollama: [],
};

function modelDefaultForProvider(provider: string): string {
  switch (provider) {
    case "deepseek":
      return "deepseek-chat";
    case "ollama":
      return "llama3.2";
    default:
      return "google/gemini-2.5-flash-lite";
  }
}

function isStaleModel(provider: string, model: string): boolean {
  if (!model) return true;
  switch (provider) {
    case "deepseek":
      return model.includes("/") || model.toLowerCase().includes("llama") || model.toLowerCase().includes("gemini");
    case "ollama":
      return model.includes("/") || model.toLowerCase().includes("deepseek") || model.toLowerCase().includes("gemini");
    default:
      return model.startsWith("deepseek") || model.toLowerCase().includes("llama");
  }
}

const themeDescriptions: Record<ThemeName, string> = {
  "nomnom-dark": "Balanced dark surfaces with cool blue accents.",
  "nomnom-light": "Clean light palette for bright workspaces.",
  "catppuccin-mocha": "Warm dark contrast with softer edges.",
  "catppuccin-latte": "Muted light tones with a gentle accent.",
  "dracula": "Iconic purple/pink on dark — the classic hacker theme.",
  "nord": "Arctic blue-grey — clean, minimal, and easy on the eyes.",
  "gruvbox": "Retro-warm earth tones with a vintage terminal feel.",
  "tokyo-night": "Moody purple-blue inspired by Tokyo at night.",
};

export function SettingsView({ theme, onThemeChange, onDirtyChange }: SettingsViewProps) {
  const { config, configPath, loading, error, save, applyConfigPath } = useConfig();
  const { notify } = useToast();
  const { scale, setScale, min, max, step } = useFontScale();
  const [draft, setDraft] = useState<DesktopConfig | null>(null);
  const [saving, setSaving] = useState(false);
  const [configBusy, setConfigBusy] = useState(false);
  const [pathDraft, setPathDraft] = useState("");
  const [showApiKey, setShowApiKey] = useState(false);
  const visibleThemes = 2;
  const [themesExpanded, setThemesExpanded] = useState(false);
  const [openRouterKeyStatus, setOpenRouterKeyStatus] = useState<OpenRouterKeyStatus | null>(null);
  const [apiKeyTestResult, setApiKeyTestResult] = useState<OpenRouterTestResult | null>(null);
  const [showApiKeyResponse, setShowApiKeyResponse] = useState(false);
  const [testingApiKey, setTestingApiKey] = useState(false);

  useEffect(() => {
    if (config) {
      setDraft(config);
    }
  }, [config]);

  useEffect(() => {
    setPathDraft(configPath);
  }, [configPath]);

  useEffect(() => {
    let active = true;

    async function refreshOpenRouterKeyStatus() {
      if (!draft || draft.ai.provider !== "openrouter" || draft.ai.api_key?.trim()) {
        if (active) {
          setOpenRouterKeyStatus(null);
        }
        return;
      }

      try {
        const nextStatus = await wails.checkOpenRouterAPIKey();
        if (active) {
          setOpenRouterKeyStatus(nextStatus);
        }
      } catch {
        if (active) {
          setOpenRouterKeyStatus(null);
        }
      }
    }

    void refreshOpenRouterKeyStatus();

    return () => {
      active = false;
    };
  }, [draft?.ai.api_key, draft?.ai.provider]);

  useEffect(() => {
    setApiKeyTestResult(null);
    setShowApiKeyResponse(false);
  }, [draft?.ai.api_key, draft?.ai.model, draft?.ai.provider]);

  useEffect(() => {
    if (!config || !draft) return;
    const dirty = JSON.stringify(config) !== JSON.stringify(draft);
    onDirtyChange?.(dirty);
  }, [config, draft, onDirtyChange]);

  function updateDraft(updater: (current: DesktopConfig) => DesktopConfig) {
    setDraft((current) => (current ? updater(current) : current));
  }

  async function handleSave() {
    if (!draft) {
      return;
    }
    setSaving(true);
    try {
      await save(draft);
      notify("Settings saved");
    } catch (err) {
      notify(err instanceof Error ? err.message : "Failed to save settings", "error");
    } finally {
      setSaving(false);
    }
  }

  async function handleApplyConfigPath(nextPath = pathDraft) {
    const trimmedPath = nextPath.trim();
    setConfigBusy(true);
    try {
      const nextConfig = await applyConfigPath(trimmedPath);
      const resolvedPath = await wails.getConfigPath();
      setDraft(nextConfig);
      setPathDraft(resolvedPath);
      notify("Config source updated");
    } catch (err) {
      notify(err instanceof Error ? err.message : "Failed to update config source", "error");
    } finally {
      setConfigBusy(false);
    }
  }

  async function handleBrowseConfig() {
    setConfigBusy(true);
    try {
      const selectedPath = await wails.selectConfigFile(pathDraft || configPath);
      if (selectedPath) {
        await handleApplyConfigPath(selectedPath);
      }
    } catch (err) {
      notify(err instanceof Error ? err.message : "Failed to choose config file", "error");
    } finally {
      setConfigBusy(false);
    }
  }

  async function handleCreateConfig() {
    setConfigBusy(true);
    try {
      const selectedPath = await wails.createConfigFile(pathDraft || configPath);
      if (selectedPath) {
        await handleApplyConfigPath(selectedPath);
      }
    } catch (err) {
      notify(err instanceof Error ? err.message : "Failed to create config file", "error");
    } finally {
      setConfigBusy(false);
    }
  }

  async function handleTestOpenRouterKey() {
    if (!draft) {
      return;
    }

    setTestingApiKey(true);
    setApiKeyTestResult(null);
    setShowApiKeyResponse(false);
    try {
      const result: OpenRouterTestResult = await wails.testOpenRouterAPIKey(draft.ai.api_key ?? "", draft.ai.model ?? "");
      setApiKeyTestResult(result);
      setShowApiKeyResponse(true);
    } catch (err) {
      setApiKeyTestResult({
        ok: false,
        status_code: 0,
        status_text: "",
        source: "",
        message: err instanceof Error ? err.message : "OpenRouter test failed",
        response: "",
      });
      setShowApiKeyResponse(true);
    } finally {
      setTestingApiKey(false);
    }
  }

  if (loading) {
    return <section className="text-sm text-text-secondary">Loading settings...</section>;
  }

  if (!config || !draft) {
    return <section className="text-sm text-danger">{error ?? "No config available"}</section>;
  }

  return (
    <section className="w-full space-y-4">
      <header className="space-y-1.5">
        <h1 className="text-base font-semibold tracking-[-0.02em] text-text-primary">Settings</h1>
        <p className="text-xs leading-5 text-text-secondary md:text-sm md:leading-6">Tune AI, file handling, extraction, and logging behavior for the desktop app.</p>
      </header>

      <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
        <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
          <FolderOpen className="h-4 w-4" />
          Config Source
        </div>
        <div className="grid gap-2 lg:grid-cols-[minmax(0,1fr)_auto_auto_auto]">
          <Input
            mono
            value={pathDraft}
            onChange={(event) => setPathDraft(event.target.value)}
            placeholder="/path/to/nomnom-config.json"
          />
          <Button variant="outline" onClick={handleBrowseConfig} disabled={configBusy} className="inline-flex items-center gap-2">
            <FolderOpen className="h-4 w-4" />
            Open
          </Button>
          <Button variant="ghost" onClick={handleCreateConfig} disabled={configBusy} className="inline-flex items-center gap-2">
            <FilePlus2 className="h-4 w-4" />
            New
          </Button>
          <Button variant="solid" onClick={() => void handleApplyConfigPath()} disabled={configBusy}>
            Use Config
          </Button>
        </div>
        <div className="mt-2 text-xs text-text-secondary">
          This uses the desktop config format. Existing files are verified before switching.
        </div>
      </div>

      <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
        <div className="mb-3 flex items-start justify-between gap-3">
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">Theme</div>
            <div className="mt-1 text-xs text-text-secondary">Choose the exact palette. The app updates immediately.</div>
          </div>
          <div className="rounded-full border border-border bg-surface-2/60 px-2.5 py-1 text-[11px] font-medium text-text-secondary">Live</div>
        </div>
        <div className="grid gap-2 sm:grid-cols-2">
          {themeOptions.slice(0, themesExpanded ? themeOptions.length : visibleThemes).map((option) => {
            const selected = theme === option.value;

            return (
              <button
                key={option.value}
                type="button"
                onClick={() => onThemeChange(option.value)}
                aria-pressed={selected}
                className={`flex h-full flex-col rounded-xl border p-3 text-left transition-all duration-150 ${
                  selected ? "border-accent bg-accent-subtle/50 shadow-[0_0_0_1px_rgba(118,181,224,0.08)]" : "border-border bg-surface-2/35 hover:border-accent/40 hover:bg-surface-2/60"
                }`}
              >
                <div className="flex items-start justify-between gap-3">
                  <div>
                    <div className="text-sm font-semibold text-text-primary">{option.label}</div>
                    <div className="mt-1 text-[11px] leading-5 text-text-secondary">{themeDescriptions[option.value]}</div>
                  </div>
                  <span className={`mt-1 h-2.5 w-2.5 rounded-full ${selected ? "bg-accent" : "bg-border"}`} />
                </div>
                <div className="mt-3 flex items-center justify-between text-[11px] uppercase tracking-[0.08em] text-text-secondary">
                  <span>{selected ? "Active" : "Apply"}</span>
                  <span className="rounded-full border border-border bg-surface px-2 py-0.5 mono text-[10px]">{selected ? "Selected" : "Preview"}</span>
                </div>
              </button>
            );
          })}
        </div>
        {themeOptions.length > visibleThemes ? (
          <button
            type="button"
            onClick={() => setThemesExpanded((v) => !v)}
            className="mt-2 inline-flex items-center gap-1.5 text-xs text-text-secondary hover:text-text-primary transition-colors"
          >
            {themesExpanded ? <ChevronUp className="h-3.5 w-3.5" /> : <ChevronDown className="h-3.5 w-3.5" />}
            {themesExpanded ? "Show Less" : `View ${themeOptions.length - visibleThemes} More Themes`}
          </button>
        ) : null}
        <div className="mt-2 text-xs text-text-secondary">The homepage keeps only a quick light/dark toggle; this view controls the exact palette.</div>
      </div>

      <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
        <div className="mb-3 flex items-center justify-between gap-3">
          <div>
            <div className="text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">Font Scale</div>
            <div className="mt-1 text-xs text-text-secondary">Adjust UI text size. Scales the entire interface proportionally.</div>
          </div>
          <div className="rounded-full border border-border bg-surface-2/60 px-2.5 py-1 text-[11px] font-medium text-text-secondary">{Math.round(scale * 100)}%</div>
        </div>
        <div className="flex items-center gap-3">
          <span className="text-xs text-text-secondary">{Math.round(min * 100)}%</span>
          <input
            type="range"
            min={min}
            max={max}
            step={step}
            value={scale}
            onChange={(e) => setScale(Number(e.target.value))}
            className="h-1.5 flex-1 appearance-none rounded-full bg-surface-2 accent-accent cursor-pointer"
          />
          <span className="text-xs text-text-secondary">{Math.round(max * 100)}%</span>
        </div>
      </div>

      <div className="grid grid-cols-1 gap-3 xl:grid-cols-2">
        <div className="space-y-3">
          <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
            <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
              <FileCog className="h-4 w-4" />
              App Output + Case
            </div>
            <div className="grid grid-cols-1 gap-2.5 lg:grid-cols-2">
              <div>
                <div className="mb-1 text-xs text-text-secondary">Output Directory</div>
                <Input mono value={draft.output} onChange={(event) => updateDraft((current) => ({ ...current, output: event.target.value }))} />
              </div>
              <div>
                <div className="mb-1 text-xs text-text-secondary">Case</div>
                <Select value={draft.case} onChange={(event) => updateDraft((current) => ({ ...current, case: event.target.value }))}>
                  {caseOptions.map((option) => (
                    <option key={option} value={option}>
                      {option}
                    </option>
                  ))}
                </Select>
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
            <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
              <Boxes className="h-4 w-4" />
              File Handling
            </div>
            <div className="space-y-2.5">
              <div>
                <div className="mb-1 text-xs text-text-secondary">Max Size</div>
                <Input
                  value={draft.file_handling.max_size}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      file_handling: {
                        ...current.file_handling,
                        max_size: event.target.value,
                      },
                    }))
                  }
                />
              </div>
              <Toggle
                label="Auto Approve"
                checked={draft.file_handling.auto_approve}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    file_handling: {
                      ...current.file_handling,
                      auto_approve: event.target.checked,
                    },
                  }))
                }
                description="Skip confirmation when the plan is safe to execute."
              />
              <Toggle
                label="Move Files Instead of Copy"
                checked={draft.file_handling.move_files}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    file_handling: {
                      ...current.file_handling,
                      move_files: event.target.checked,
                    },
                  }))
                }
                description="Relocate files when the filesystem supports it."
              />
              <Toggle
                label="Skip Dot Files"
                checked={draft.file_handling.skip_dot_files}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    file_handling: {
                      ...current.file_handling,
                      skip_dot_files: event.target.checked,
                    },
                  }))
                }
                description="Ignore hidden files and folders that start with a dot."
              />
            </div>
          </div>

          <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
            <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
              <Gauge className="h-4 w-4" />
              Performance
            </div>
            <div className="space-y-2.5">
              <div className="text-xs text-text-secondary">AI Pipeline</div>
              <div className="grid grid-cols-3 gap-2">
                <Input
                  type="number"
                  value={draft.performance.ai.workers}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        ai: {
                          ...current.performance.ai,
                          workers: Number(event.target.value || 0),
                        },
                      },
                    }))
                  }
                  placeholder="Workers"
                />
                <Input
                  value={draft.performance.ai.timeout}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        ai: {
                          ...current.performance.ai,
                          timeout: event.target.value,
                        },
                      },
                    }))
                  }
                  placeholder="Timeout"
                />
                <Input
                  type="number"
                  value={draft.performance.ai.retries}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        ai: {
                          ...current.performance.ai,
                          retries: Number(event.target.value || 0),
                        },
                      },
                    }))
                  }
                  placeholder="Retries"
                />
              </div>
              <div className="text-xs text-text-secondary">File Pipeline</div>
              <div className="grid grid-cols-3 gap-2">
                <Input
                  type="number"
                  value={draft.performance.file.workers}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        file: {
                          ...current.performance.file,
                          workers: Number(event.target.value || 0),
                        },
                      },
                    }))
                  }
                  placeholder="Workers"
                />
                <Input
                  value={draft.performance.file.timeout}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        file: {
                          ...current.performance.file,
                          timeout: event.target.value,
                        },
                      },
                    }))
                  }
                  placeholder="Timeout"
                />
                <Input
                  type="number"
                  value={draft.performance.file.retries}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      performance: {
                        ...current.performance,
                        file: {
                          ...current.performance.file,
                          retries: Number(event.target.value || 0),
                        },
                      },
                    }))
                  }
                  placeholder="Retries"
                />
              </div>
            </div>
          </div>
        </div>

        <div className="space-y-3">
          <div className="mt-auto rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
            <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
              <Bot className="h-4 w-4" />
              AI Config
            </div>
            <div className="space-y-2.5">
              <div>
                <div className="mb-1 text-xs text-text-secondary">Provider</div>
                <Select
                  value={draft.ai.provider}
                  onChange={(event) => {
                    const nextProvider = event.target.value;
                    updateDraft((current) => {
                      const prevProvider = current.ai.provider;
                      const prevModel = current.ai.model;
                      const model =
                        prevProvider !== nextProvider && isStaleModel(nextProvider, prevModel)
                          ? modelDefaultForProvider(nextProvider)
                          : prevModel;
                      return {
                        ...current,
                        ai: {
                          ...current.ai,
                          provider: nextProvider,
                          model,
                        },
                      };
                    });
                  }}
                >
                  {providerOptions.map((provider) => (
                    <option key={provider} value={provider}>
                      {provider}
                    </option>
                  ))}
                </Select>
              </div>
              <div>
                <div className="mb-1 text-xs text-text-secondary">Model</div>
                <Input
                  value={draft.ai.model}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      ai: {
                        ...current.ai,
                        model: event.target.value,
                      },
                    }))
                  }
                  list="model-suggestions"
                />
                <datalist id="model-suggestions">
                  {(modelSuggestions[draft.ai.provider] ?? []).map((m) => (
                    <option key={m} value={m}>
                      {m.split("/").slice(-1)[0]}
                    </option>
                  ))}
                </datalist>
              </div>
              <div>
                <div className="mb-1 inline-flex items-center gap-1 text-xs text-text-secondary">
                  <KeyRound className="h-3.5 w-3.5" />
                  API Key
                </div>
                <div className="flex items-center gap-2">
                  <div className="relative flex-1">
                    <Input
                      type={showApiKey ? "text" : "password"}
                      value={draft.ai.api_key ?? ""}
                      onChange={(event) =>
                        updateDraft((current) => ({
                          ...current,
                          ai: {
                            ...current.ai,
                            api_key: event.target.value,
                          },
                        }))
                      }
                      className="pr-10"
                    />
                    <button
                      type="button"
                      onClick={() => setShowApiKey((current) => !current)}
                      className="absolute right-2 top-1/2 -translate-y-1/2 rounded-md p-1 text-text-secondary transition-colors hover:bg-surface-raised hover:text-text-primary"
                      aria-label={showApiKey ? "Hide API key" : "Show API key"}
                    >
                      {showApiKey ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                    </button>
                  </div>
                  <Button
                    type="button"
                    variant="outline"
                    onClick={() => void handleTestOpenRouterKey()}
                    disabled={testingApiKey || draft.ai.provider !== "openrouter"}
                    className={`h-9 shrink-0 px-3 text-xs ${
                      testingApiKey
                        ? "border-accent/50 bg-accent-subtle text-accent"
                        : apiKeyTestResult?.ok === true
                          ? "border-success/60 bg-success/10 text-success hover:bg-success/15"
                          : apiKeyTestResult?.ok === false
                            ? "border-danger/60 bg-danger/10 text-danger hover:bg-danger/15"
                            : ""
                    }`}
                  >
                    {testingApiKey ? <Loader2 className="h-3.5 w-3.5 animate-spin" /> : "Test"}
                  </Button>
                </div>
                {draft.ai.provider === "openrouter" ? (
                  <div className="mt-1 text-[11px] leading-5 text-text-secondary">
                    {draft.ai.api_key?.trim()
                      ? "This field overrides the environment key for OpenRouter requests."
                      : openRouterKeyStatus?.available
                        ? `No config key set. Using OPENROUTER_API_KEY from ${openRouterKeyStatus.source === "env" ? "the environment" : "config"}.`
                        : "No OpenRouter API key found in the config or OPENROUTER_API_KEY."}
                  </div>
                ) : null}
                {apiKeyTestResult ? (
                  <div className="mt-2 rounded-lg border border-border bg-surface-raised/30 px-3 py-2 text-[11px] leading-5 text-text-secondary">
                    <div className="flex items-start justify-between gap-3">
                      <div>
                        <div className="font-medium text-text-primary">
                          {apiKeyTestResult.ok ? "OpenRouter test succeeded" : "OpenRouter test failed"}
                        </div>
                        <div>
                          {apiKeyTestResult.ok
                            ? `${apiKeyTestResult.status_code} ${apiKeyTestResult.status_text || "OK"} via ${apiKeyTestResult.source || "unknown source"}.`
                            : apiKeyTestResult.message}
                        </div>
                      </div>
                      <button
                        type="button"
                        onClick={() => setShowApiKeyResponse((current) => !current)}
                        className="inline-flex items-center gap-1 rounded-md border border-border bg-surface px-2 py-1 text-[10px] font-medium uppercase tracking-[0.08em] text-text-secondary transition-colors hover:bg-surface-2"
                      >
                        {showApiKeyResponse ? <ChevronUp className="h-3 w-3" /> : <ChevronDown className="h-3 w-3" />}
                        {showApiKeyResponse ? "Hide" : "Response"}
                      </button>
                    </div>
                    {showApiKeyResponse ? (
                      <div className="mt-2 rounded-md border border-border bg-surface px-3 py-2 text-[10px] leading-5 text-text-secondary">
                        <div className="flex items-center justify-between gap-2 border-b border-border/70 pb-1 text-[10px] uppercase tracking-[0.08em] text-text-secondary">
                          <span>HTTP Response</span>
                          <span>{apiKeyTestResult.status_code ? `${apiKeyTestResult.status_code} ${apiKeyTestResult.status_text}` : "No status code"}</span>
                        </div>
                        <pre className="mt-2 max-h-40 overflow-auto whitespace-pre-wrap break-words mono text-[10px] leading-5 text-text-primary">
                          {apiKeyTestResult.response || "No response body returned."}
                        </pre>
                      </div>
                    ) : null}
                  </div>
                ) : null}
              </div>
              <div className="grid grid-cols-2 gap-2">
                <div>
                  <div className="mb-1 text-xs text-text-secondary">Max Tokens</div>
                  <Input
                    type="number"
                    value={draft.ai.max_tokens}
                    onChange={(event) =>
                      updateDraft((current) => ({
                        ...current,
                        ai: {
                          ...current.ai,
                          max_tokens: Number(event.target.value || 0),
                        },
                      }))
                    }
                  />
                </div>
                <div>
                  <div className="mb-1 text-xs text-text-secondary">Temperature</div>
                  <Input
                    type="number"
                    step="0.1"
                    value={draft.ai.temperature}
                    onChange={(event) =>
                      updateDraft((current) => ({
                        ...current,
                        ai: {
                          ...current.ai,
                          temperature: Number(event.target.value || 0),
                        },
                      }))
                    }
                  />
                </div>
              </div>
              <div>
                <div className="mb-1 text-xs text-text-secondary">Prompt</div>
                <textarea
                  value={draft.ai.prompt}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      ai: {
                        ...current.ai,
                        prompt: event.target.value,
                      },
                    }))
                  }
                  className="min-h-24 w-full rounded-lg border border-border bg-surface-raised/80 px-3.5 py-2.5 text-sm text-text-primary transition-colors duration-150 focus-visible:border-accent focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-accent/15"
                />
              </div>
              <div className="grid grid-cols-1 gap-2 lg:grid-cols-2">
                <Toggle
                  label="Vision Enabled"
                  checked={draft.ai.vision.enabled}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      ai: {
                        ...current.ai,
                        vision: {
                          ...current.ai.vision,
                          enabled: event.target.checked,
                        },
                      },
                    }))
                  }
                  description="Include image context when available."
                />
                <div>
                  <div className="mb-1 text-xs text-text-secondary">Max Image Size</div>
                  <Input
                    value={draft.ai.vision.max_image_size}
                    onChange={(event) =>
                      updateDraft((current) => ({
                        ...current,
                        ai: {
                          ...current.ai,
                          vision: {
                            ...current.ai.vision,
                            max_image_size: event.target.value,
                          },
                        },
                      }))
                    }
                  />
                </div>
              </div>
            </div>
          </div>

          <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)]">
            <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
              <Sparkles className="h-4 w-4" />
              Content Extraction
            </div>
            <div className="space-y-2.5">
              <Toggle
                label="Extract Text"
                checked={draft.content_extraction.extract_text}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    content_extraction: {
                      ...current.content_extraction,
                      extract_text: event.target.checked,
                    },
                  }))
                }
                description="Read file contents when available."
              />
              <Toggle
                label="Extract Metadata"
                checked={draft.content_extraction.extract_metadata}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    content_extraction: {
                      ...current.content_extraction,
                      extract_metadata: event.target.checked,
                    },
                  }))
                }
                description="Include filesystem metadata in the prompt."
              />
              <Toggle
                label="Skip Large Files"
                checked={draft.content_extraction.skip_large_files}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    content_extraction: {
                      ...current.content_extraction,
                      skip_large_files: event.target.checked,
                    },
                  }))
                }
                description="Ignore files over the configured threshold."
              />
              <Toggle
                label="Read Context"
                checked={draft.content_extraction.read_context}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    content_extraction: {
                      ...current.content_extraction,
                      read_context: event.target.checked,
                    },
                  }))
                }
                description="Use nearby files for better naming context."
              />
              <div>
                <div className="mb-1 text-xs text-text-secondary">Max Content Length</div>
                <Input
                  type="number"
                  value={draft.content_extraction.max_content_length}
                  onChange={(event) =>
                    updateDraft((current) => ({
                      ...current,
                      content_extraction: {
                        ...current.content_extraction,
                        max_content_length: Number(event.target.value || 0),
                      },
                    }))
                  }
                />
              </div>
            </div>
          </div>
        </div>

        <div className="rounded-xl border border-border bg-surface p-3.5 shadow-[0_1px_2px_rgba(0,0,0,0.10)] xl:col-span-2">
          <div className="mb-3 flex items-center gap-2 text-[11px] font-semibold uppercase tracking-[0.08em] text-text-secondary">
            <SlidersHorizontal className="h-4 w-4" />
            Logging
          </div>
          <div className="grid grid-cols-1 gap-2.5 lg:grid-cols-2">
            <Toggle
              label="Logging Enabled"
              checked={draft.logging.enabled}
              onChange={(event) =>
                updateDraft((current) => ({
                  ...current,
                  logging: {
                    ...current.logging,
                    enabled: event.target.checked,
                  },
                }))
              }
              description="Write detailed local logs for troubleshooting."
            />
            <div>
              <div className="mb-1 text-xs text-text-secondary">Log Path</div>
              <Input
                mono
                value={draft.logging.log_path}
                onChange={(event) =>
                  updateDraft((current) => ({
                    ...current,
                    logging: {
                      ...current.logging,
                      log_path: event.target.value,
                    },
                  }))
                }
              />
            </div>
          </div>
        </div>
      </div>

      <div className="flex flex-wrap items-center justify-end gap-2">
        <Button variant="danger" onClick={() => setDraft(config)}>
          Discard Draft
        </Button>
        <Button variant="solid" onClick={handleSave} disabled={saving || configBusy}>
          {saving ? "Saving..." : "Save Settings"}
        </Button>
      </div>
      {error ? <div className="text-sm text-danger">{error}</div> : null}
    </section>
  );
}
