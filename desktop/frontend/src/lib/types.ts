import type { LucideIcon } from "lucide-react";

export type ViewRoute = "rename" | "history" | "analytics" | "settings";

export interface NavItem {
  route: ViewRoute;
  label: string;
  Icon: LucideIcon;
}

export interface RenameEntry {
  index: number;
  original: string;
  new_name: string;
  type: string;
  status: string;
  size_bytes?: number;
  reason?: string;
}

export interface JobSummary {
  planned: number;
  renamed: number;
  skipped: number;
  errors: number;
}

export interface JobStatus {
  job_id: string;
  state: string;
  done: number;
  total: number;
  current_file: string;
  message: string;
  summary: JobSummary;
  output_dir: string;
}

export interface RunJobOptions {
  dry_run: boolean;
  log_session: boolean;
  auto_approve: boolean;
  hot_rename: boolean;
  organize: boolean;
}

export interface Session {
  date: string;
  directory: string;
  files: string;
  model: string;
  mode: string;
  status: string;
}

export interface AnalyticsSessionPoint {
  date: string;
  renamed: number;
}

export interface AnalyticsSummary {
  sessions: number;
  renamed: number;
  tokens: number;
  avg_per_run: number;
  recent_runs: number;
  unique_models: number;
  recent_sessions: AnalyticsSessionPoint[];
  history_error?: string;
}

export interface DesktopConfig {
  output: string;
  case: string;
  ai: {
    provider: string;
    model: string;
    api_key?: string;
    max_tokens: number;
    temperature: number;
    vision: {
      enabled: boolean;
      max_image_size: string;
    };
    prompt: string;
  };
  file_handling: {
    max_size: string;
    auto_approve: boolean;
    hot_rename: boolean;
    skip_dot_files: boolean;
  };
  content_extraction: {
    extract_text: boolean;
    extract_metadata: boolean;
    max_content_length: number;
    skip_large_files: boolean;
    read_context: boolean;
  };
  performance: {
    ai: {
      workers: number;
      timeout: string;
      retries: number;
    };
    file: {
      workers: number;
      timeout: string;
      retries: number;
    };
  };
  logging: {
    enabled: boolean;
    log_path: string;
  };
}

export interface OllamaStatus {
  running: boolean;
  model_available: boolean;
  message: string;
}

export interface AIStatus {
  provider: string;
  model: string;
  ollama: OllamaStatus;
}

export interface OpenRouterKeyStatus {
  available: boolean;
  source: string;
}

export interface OpenRouterTestResult {
  ok: boolean;
  status_code: number;
  status_text: string;
  source: string;
  message: string;
  response: string;
}
